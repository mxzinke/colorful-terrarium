package terrain

import (
	"fmt"
	"sync"
	"time"
)

// cacheEntry represents a cached elevation map with its expiration time
type cacheEntry struct {
	data      *ElevationMap
	expiresAt time.Time
}

// elevationCache is a thread-safe cache for elevation maps
type elevationCache struct {
	mapMu      sync.RWMutex // Mutex for protecting the entries map
	entries    map[string]cacheEntry
	keyMu      sync.RWMutex                  // Mutex for protecting the keyMutexes map
	keyMutexes map[string]*sync.RWMutex      // Per-key mutexes
	inFlightMu sync.RWMutex                  // Mutex for protecting the inFlight map
	inFlight   map[string]chan *ElevationMap // In-flight requests
	done       chan struct{}                 // Channel to signal cleanup goroutine to stop
}

// newElevationCache creates a new elevation cache and starts the cleanup routine
func newElevationCache() *elevationCache {
	cache := &elevationCache{
		entries:    make(map[string]cacheEntry),
		keyMutexes: make(map[string]*sync.RWMutex),
		inFlight:   make(map[string]chan *ElevationMap),
		done:       make(chan struct{}),
	}
	go cache.startCleanupRoutine()
	return cache
}

// getKeyMutex returns the mutex for a given key, creating it if it doesn't exist
func (c *elevationCache) getKeyMutex(key string) *sync.RWMutex {
	c.keyMu.Lock()
	defer c.keyMu.Unlock()

	if mu, exists := c.keyMutexes[key]; exists {
		return mu
	}

	mu := &sync.RWMutex{}
	c.keyMutexes[key] = mu
	return mu
}

// startCleanupRoutine starts a goroutine that periodically cleans up expired entries
func (c *elevationCache) startCleanupRoutine() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.done:
			return
		}
	}
}

// cleanup removes expired entries from the cache
func (c *elevationCache) cleanup() {
	now := time.Now()

	// First, get all keys that need to be checked
	c.mapMu.RLock()
	keys := make([]string, 0, len(c.entries))
	for key := range c.entries {
		keys = append(keys, key)
	}
	c.mapMu.RUnlock()

	// Check each key individually
	for _, key := range keys {
		keyMutex := c.getKeyMutex(key)
		keyMutex.Lock()

		c.mapMu.Lock()
		if entry, exists := c.entries[key]; exists && now.After(entry.expiresAt) {
			delete(c.entries, key)
			// Clean up the key mutex as well
			c.keyMu.Lock()
			delete(c.keyMutexes, key)
			c.keyMu.Unlock()
		}
		c.mapMu.Unlock()

		keyMutex.Unlock()
	}
}

// Stop stops the cleanup routine
func (c *elevationCache) Stop() {
	close(c.done)
}

// getCacheKey generates a unique key for a tile coordinate
func getCacheKey(coord TileCoord) string {
	return fmt.Sprintf("%d/%d/%d", coord.Z, coord.X, coord.Y)
}

var (
	// Global cache instance
	globalCache = newElevationCache()
	// Cache expiration duration - increased to 5 minutes since cleanup is now active
	cacheExpiration = 5 * time.Minute
)

// Get retrieves an elevation map from the cache if it exists and hasn't expired
func (c *elevationCache) Get(coord TileCoord) (*ElevationMap, bool) {
	key := getCacheKey(coord)
	keyMutex := c.getKeyMutex(key)

	// Use read lock for the specific key
	keyMutex.RLock()
	defer keyMutex.RUnlock()

	// Use read lock for the map access
	c.mapMu.RLock()
	entry, exists := c.entries[key]
	c.mapMu.RUnlock()

	if !exists {
		return nil, false
	}

	// Check if the entry has expired
	if time.Now().After(entry.expiresAt) {
		return nil, false
	}

	return entry.data, true
}

// Set stores an elevation map in the cache with an expiration time
func (c *elevationCache) Set(coord TileCoord, data *ElevationMap) {
	key := getCacheKey(coord)
	keyMutex := c.getKeyMutex(key)

	// Use write lock for the specific key
	keyMutex.Lock()
	defer keyMutex.Unlock()

	// Use write lock for the map access
	c.mapMu.Lock()
	c.entries[key] = cacheEntry{
		data:      data,
		expiresAt: time.Now().Add(cacheExpiration),
	}
	c.mapMu.Unlock()
}

// GetOrCreate gets a value from cache or creates it using the provided function
func (c *elevationCache) GetOrCreate(coord TileCoord, create func() (*ElevationMap, error)) (*ElevationMap, error) {
	key := getCacheKey(coord)

	// First try to get from cache
	if em, found := c.Get(coord); found {
		return em, nil
	}

	// Check for in-flight request
	c.inFlightMu.Lock()
	if ch, exists := c.inFlight[key]; exists {
		c.inFlightMu.Unlock()
		// Wait for the in-flight request to complete
		return <-ch, nil
	}

	// Create channel for this request
	ch := make(chan *ElevationMap, 1)
	c.inFlight[key] = ch
	c.inFlightMu.Unlock()

	// Create the value
	em, err := create()
	if err != nil {
		// On error, remove from in-flight and return error
		c.inFlightMu.Lock()
		delete(c.inFlight, key)
		c.inFlightMu.Unlock()
		return nil, err
	}

	// Store in cache
	c.Set(coord, em)

	// Notify any waiting goroutines and cleanup
	ch <- em
	c.inFlightMu.Lock()
	delete(c.inFlight, key)
	c.inFlightMu.Unlock()

	return em, nil
}
