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
	client *http.Client
}

func newTileServer() *TileServer {
	return &TileServer{
		client: &http.Client{},
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

	// Download and process tile
	processedTile, err := s.processTile(uint32(z), uint32(y), uint32(x))
	if err != nil {
		log.Printf("Error processing tile %d/%d/%d: %v", z, y, x, err)
		http.Error(w, "Error processing tile", http.StatusInternalServerError)
		return
	}

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

	log.Printf("z: %d y: %d x: %d\n", z, y, x)

	// Download subtiles
	tiles, err := downloadSubTiles(z, y, x)
	if err != nil {
		log.Printf("Error downloading subtiles: %v", err)
		return nil, err
	}

	// Combine tiles
	combined := compositeImages(tiles)

	// Get latitudes (min and max) for the tile
	minLat, maxLat := getTileLatitudes(z, y, x)

	// Process and colorize
	processed := processAndColorize(combined, minLat, maxLat)

	// Encode processed tile
	var buf bytes.Buffer
	if err := encodePNGOptimized(&buf, processed); err != nil {
		return []byte{}, fmt.Errorf("failed to encode processed tile: %w", err)
	}

	return buf.Bytes(), nil
}

func processAndColorize(img image.Image, minLat, maxLat float64) *image.RGBA {
	bounds := img.Bounds()
	output := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		// Calculate precise latitude for this pixel row
		latitude := math.Abs(getLatitudeForPixel(y, minLat, maxLat, bounds.Max.Y))

		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()

			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)

			// Calculate base elevation and apply latitude-based offset
			elevation := float64(r8)*256 + float64(g8) + float64(b8)/256 - 32768

			newColor := getColorForElevationAndLatitude(elevation, latitude)
			output.Set(x, y, color.RGBA{newColor.R, newColor.G, newColor.B, 255})
		}
	}

	return output
}

func main() {
	server := newTileServer()
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
