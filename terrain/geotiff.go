package terrain

import (
	"bytes"
	"context"
	"fmt"
	"image/color"
	"io"
	"net/http"
	"sync"
	"time"

	tiff "github.com/chai2010/tiff"
)

const geotiffSourceURL = "https://elevation-tiles-prod.s3.dualstack.us-east-1.amazonaws.com/geotiff/%d/%d/%d.tif"

// cacheEntry represents a cached elevation map with its expiration time
type cacheEntry struct {
	data      *ElevationMap
	expiresAt time.Time
}

// elevationCache is a thread-safe cache for elevation maps
type elevationCache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	done    chan struct{} // Channel to signal cleanup goroutine to stop
}

// newElevationCache creates a new elevation cache and starts the cleanup routine
func newElevationCache() *elevationCache {
	cache := &elevationCache{
		entries: make(map[string]cacheEntry),
		done:    make(chan struct{}),
	}
	go cache.startCleanupRoutine()
	return cache
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
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, entry := range c.entries {
		if now.After(entry.expiresAt) {
			delete(c.entries, key)
		}
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

func GetElevationMapFromGeoTIFF(ctx context.Context, coord TileCoord) (*ElevationMap, error) {
	cacheKey := getCacheKey(coord)

	// Try to get from cache first
	globalCache.mu.RLock()
	if entry, exists := globalCache.entries[cacheKey]; exists && time.Now().Before(entry.expiresAt) {
		globalCache.mu.RUnlock()
		return entry.data, nil
	}
	globalCache.mu.RUnlock()

	// If not in cache or expired, fetch new data
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf(geotiffSourceURL, coord.Z, coord.X, coord.Y), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for GeoTIFF: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download GeoTIFF: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d", res.StatusCode)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read TIFF: %v", err)
	}

	// Create a new TIFF reader
	matrix, tileSize, err := readTIFFToFloat32Matrix(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode TIFF: %v", err)
	}

	// Copy over data to ElevationMap
	em := &ElevationMap{
		Data:     matrix,
		TileSize: tileSize,
	}

	// Store in cache
	globalCache.mu.Lock()
	globalCache.entries[cacheKey] = cacheEntry{
		data:      em,
		expiresAt: time.Now().Add(cacheExpiration),
	}
	globalCache.mu.Unlock()

	return em, nil
}

func readTIFFToFloat32Matrix(data []byte) ([][]float32, int, error) {
	// Decode TIFF
	images, errors, err := tiff.DecodeAll(bytes.NewReader(data))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode TIFF: %v", err)
	}

	// We'll process the first image/band
	if len(images) == 0 || len(images[0]) == 0 {
		return nil, 0, fmt.Errorf("no images found in TIFF")
	}
	if errors[0][0] != nil {
		return nil, 0, fmt.Errorf("error in first image: %v", errors[0][0])
	}

	img := images[0][0]
	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	// Create result matrix
	matrix := make([][]float32, height)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		matrix[y] = make([]float32, width)

		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			height := img.At(x, y)

			if _, ok := height.(color.Gray16); ok {
				matrix[y][x] = float32(int16(height.(color.Gray16).Y))
				continue
			}

			if _, ok := height.(color.Gray); ok {
				// Ignore this, as the data is empty!
				continue
			}

			return nil, 0, fmt.Errorf("unsupported color type: %T", height)
		}
	}

	return matrix, img.Bounds().Max.X, nil
}
