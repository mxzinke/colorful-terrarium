package terrain

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

const terrariumSourceURL = "https://elevation-tiles-prod.s3.amazonaws.com/terrarium/%d/%d/%d.png"
const tileSize = 256 // Standard tile size

func GetElevationMapForTerrarium(ctx context.Context, coord TileCoord) (*ElevationMap, error) {
	tiles, err := downloadSubTiles(ctx, coord.Z, coord.X, coord.Y)
	if err != nil {
		log.Printf("Error downloading subtiles: %v", err)
		return nil, err
	}

	img := compositeImages(tiles)

	return newElevationMapFromTerrarium(img), nil
}

func downloadTile(ctx context.Context, coord TileCoord) (image.Image, error) {
	maxRetries := 3
	retryDelay := 200 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryDelay)
			log.Printf("Retry attempt %d for tile %d/%d/%d", attempt, coord.Z, coord.Y, coord.X)
		}

		url := fmt.Sprintf(terrariumSourceURL, coord.Z, coord.Y, coord.X)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			log.Printf("Download error (creating request) (attempt %d): %v", attempt+1, err)
			continue
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Download error (attempt %d): %v", attempt+1, err)
			continue
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			log.Printf("HTTP error %d (attempt %d)", res.StatusCode, attempt+1)
			continue
		}

		// Read response body
		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Printf("Read error (attempt %d): %v", attempt+1, err)
			continue
		}

		// Decode PNG
		img, err := png.Decode(bytes.NewReader(body))
		if err != nil {
			log.Printf("Decode error (attempt %d): %v", attempt+1, err)
			continue
		}

		return img, nil
	}

	return nil, fmt.Errorf("failed to download tile after %d attempts", maxRetries)
}

func downloadSubTiles(ctx context.Context, parentZ, parentY, parentX uint32) ([]tileImage, error) {
	childZ := parentZ + 1
	baseChildX := parentX * 2
	baseChildY := parentY * 2

	var wg sync.WaitGroup
	tiles := make([]tileImage, 4)
	errors := make(chan error, 4)

	// Download all 4 subtiles concurrently
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			wg.Add(1)
			go func(index, offsetX, offsetY uint32) {
				defer wg.Done()
				coord := TileCoord{
					Z: childZ,
					X: baseChildX + offsetX,
					Y: baseChildY + offsetY,
				}
				img, err := downloadTile(ctx, coord)
				if err != nil {
					errors <- err
					return
				}
				tiles[index] = tileImage{Coord: coord, Image: img}
			}(uint32(i*2+j), uint32(j), uint32(i))
		}
	}

	// Wait for all downloads to complete
	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		if err != nil {
			return nil, err
		}
	}

	return tiles, nil
}

type tileImage struct {
	Coord TileCoord
	Image image.Image
}

func compositeImages(tiles []tileImage) *image.RGBA {
	// Create a new image with double the dimensions
	combined := image.NewRGBA(image.Rect(0, 0, tileSize*2, tileSize*2))

	// Copy each tile into the correct position with corrected orientation
	for _, tile := range tiles {
		bounds := tile.Image.Bounds()
		offsetX := (tile.Coord.Y % 2) * tileSize
		offsetY := (tile.Coord.X % 2) * tileSize

		for y := 0; y < bounds.Dy(); y++ {
			for x := 0; x < bounds.Dx(); x++ {
				combined.Set(x+int(offsetX), y+int(offsetY), tile.Image.At(x, y))
			}
		}
	}

	return combined
}

// newElevationMapFromTerrarium creates a new ElevationMap from a terrarium image
func newElevationMapFromTerrarium(img image.Image) *ElevationMap {
	bounds := img.Bounds()
	data := make([][]float32, bounds.Dy())

	for y := 0; y < bounds.Dy(); y++ {
		data[y] = make([]float32, bounds.Dx())
		for x := 0; x < bounds.Dx(); x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)
			data[y][x] = float32(r8)*256.0 + float32(g8) + float32(b8)/256.0 - 32768.0
		}
	}

	return &ElevationMap{
		Data:     data,
		TileSize: bounds.Max.X,
	}
}
