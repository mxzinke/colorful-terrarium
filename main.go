package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mxzinke/colorful-terrarium/terrain"
)

const (
	port    = 8080
	maxZoom = 14
)

type compressionWriter struct {
	io.Writer
	http.ResponseWriter
}

func (cw *compressionWriter) Write(b []byte) (int, error) {
	return cw.Writer.Write(b)
}

func (cw *compressionWriter) Header() http.Header {
	return cw.ResponseWriter.Header()
}

var (
	// URL pattern: /{z}/{y}/{x}.png
	tilePattern = regexp.MustCompile(`^/(\d{1,2})/(\d+)/(\d+)\.png$`)
)

type TileServer struct {
	client      *http.Client
	geoCoverage *terrain.GeoCoverage
}

func newTileServer(geoCoverage *terrain.GeoCoverage) *TileServer {
	return &TileServer{
		client:      &http.Client{},
		geoCoverage: geoCoverage,
	}
}

func (s *TileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only handle GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Match URL pattern
	matches := tilePattern.FindStringSubmatch(r.URL.Path)
	if matches == nil {
		http.Error(w, "Invalid tile URL", http.StatusBadRequest)
		return
	}

	// Parse tile coordinates
	z, _ := strconv.ParseUint(matches[1], 10, 32)
	y, _ := strconv.ParseUint(matches[2], 10, 32)
	x, _ := strconv.ParseUint(matches[3], 10, 32)

	log.Printf("Requesting tile %d/%d/%d\n", z, y, x)

	// Download and process tile
	startTime := time.Now()
	processedTile, err := s.processTile(uint32(z), uint32(y), uint32(x))
	if err != nil {
		log.Printf("Error processing tile %d/%d/%d: %v", z, y, x, err)
		http.Error(w, "Error processing tile", http.StatusInternalServerError)
		return
	}
	elapsed := time.Since(startTime)

	log.Printf("Processed tile %d/%d/%d in %s", z, y, x, elapsed)

	if len(processedTile) == 0 {
		http.Error(w, "Tile not found", http.StatusNotFound)
		return
	}

	// Set content type and write response
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400") // 24h cache

	w.Write(processedTile)
}

// enableCompression wraps a handler with compression support
func enableCompression(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts compression
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Set compression headers
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")

		// Create gzip writer
		gz := gzip.NewWriter(w)
		defer gz.Close()

		// Wrap response writer
		wrapped := &compressionWriter{Writer: gz, ResponseWriter: w}

		// Call the original handler with our wrapped writer
		next.ServeHTTP(wrapped, r)
	})
}

func (s *TileServer) processTile(z, y, x uint32) ([]byte, error) {
	if z > maxZoom {
		return []byte{}, nil
	}

	maxScale := uint32(math.Pow(2.0, float64(z)))
	if y >= maxScale || x >= maxScale {
		return []byte{}, nil
	}

	// Download subtiles
	tiles, err := downloadSubTiles(z, y, x)
	if err != nil {
		log.Printf("Error downloading subtiles: %v", err)
		return nil, err
	}

	// Combine tiles
	combined := compositeImages(tiles)

	// Get latitudes (min and max) for the tile
	bounds := CreateTileBounds(z, y, x, combined.Bounds().Max.X)

	// Process and colorize
	processed := processAndColorize(s.geoCoverage, combined, bounds, z)

	// Encode processed tile
	var buf bytes.Buffer
	if err := encodePNGOptimized(&buf, processed); err != nil {
		return []byte{}, fmt.Errorf("failed to encode processed tile: %w", err)
	}

	return buf.Bytes(), nil
}

func processAndColorize(geoCoverage *terrain.GeoCoverage, img image.Image, tileBounds *TileBounds, z uint32) *image.RGBA {
	bounds := img.Bounds()
	output := image.NewRGBA(bounds)

	elevationMap := terrain.NewElevationMap(img)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		// Calculate precise latitude for this pixel row
		baseLatitude := tileBounds.GetPixelLat(y)
		latitude := math.Abs(baseLatitude)

		// Calculate polar factor (whether the snow is in the polar region), respecting the earth's tilt
		polarFactor := 0.0
		if baseLatitude < -1*(polarStartLatitude-earthTilt) {
			polarFactor = math.Min(math.Max((latitude-(polarStartLatitude-earthTilt))/((polarAbsoluteLatitude-earthTilt)-(polarStartLatitude-earthTilt)), 0), 1)
		} else if baseLatitude > (polarStartLatitude + earthTilt) {
			polarFactor = math.Min(math.Max((latitude-(polarStartLatitude+earthTilt))/((polarAbsoluteLatitude+earthTilt)-(polarStartLatitude+earthTilt)), 0), 1)
		}

		// Calculate snow threshold factor (bringing down the elevation of the snow)
		snowThresholdFactor := math.Min(math.Max(latitude/polarStartLatitude, lowestSnowFactor), 1) / snowBaseFactor

		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			longitude := tileBounds.GetPixelLng(x)

			isInIce := geoCoverage.IsPointInIce(longitude, baseLatitude)

			elevation := elevationMap.GetElevation(x, y)
			smoothedElev := smoothCoastlines(elevation, x, y, elevationMap, z)

			newColor := getColorForElevationAndTerrain(
				smoothedElev*snowThresholdFactor,
				polarFactor,
				isInIce,
			)
			output.Set(x, y, color.RGBA{newColor.R, newColor.G, newColor.B, 255})
		}
	}

	return output
}

func main() {
	geoCoverage, err := terrain.LoadGeoCoverage()
	if err != nil {
		log.Fatalf("Failed to load geo coverage: %v", err)
	}

	server := newTileServer(geoCoverage)
	addr := fmt.Sprintf(":%d", port)

	// Wrap the TileServer with compression
	compressedHandler := enableCompression(server)

	log.Printf("Starting terrain tile server on %s", addr)
	log.Printf("Tiles Server Format: http://localhost:%d/{z}/{y}/{x}.png", port)

	if err := http.ListenAndServe(addr, compressedHandler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// optimizedPNGEncoder creates a PNG encoder with optimal compression settings
func encodePNGOptimized(w io.Writer, img image.Image) error {
	encoder := &png.Encoder{
		CompressionLevel: png.BestCompression, // Maximum compression
		BufferPool:       nil,                 // Use default buffer pool
	}
	return encoder.Encode(w, img)
}
