package terrain

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

const baseURL = "https://elevation-tiles-prod.s3.amazonaws.com/terrarium/%d/%d/%d.png"
const tileSize = 256 // Standard tile size

type TileCoord struct {
	Z, X, Y uint32
}

func GetElevationMapForTile(coord TileCoord) (*ElevationMap, error) {
	tiles, err := downloadSubTiles(coord.Z, coord.Y, coord.X)
	if err != nil {
		log.Printf("Error downloading subtiles: %v", err)
		return nil, err
	}

	img := compositeImages(tiles)

	return NewElevationMap(img), nil
}

func downloadTile(coord TileCoord) (image.Image, error) {
	maxRetries := 3
	retryDelay := 200 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryDelay)
			log.Printf("Retry attempt %d for tile %d/%d/%d", attempt, coord.Z, coord.Y, coord.X)
		}

		url := fmt.Sprintf(baseURL, coord.Z, coord.Y, coord.X)
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Download error (attempt %d): %v", attempt+1, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("HTTP error %d (attempt %d)", resp.StatusCode, attempt+1)
			continue
		}

		// Read response body
		body, err := io.ReadAll(resp.Body)
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

func downloadSubTiles(parentZ, parentY, parentX uint32) ([]TileImage, error) {
	childZ := parentZ + 1
	baseChildX := parentX * 2
	baseChildY := parentY * 2

	var wg sync.WaitGroup
	tiles := make([]TileImage, 4)
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
				img, err := downloadTile(coord)
				if err != nil {
					errors <- err
					return
				}
				tiles[index] = TileImage{Coord: coord, Image: img}
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

type TileImage struct {
	Coord TileCoord
	Image image.Image
}

func compositeImages(tiles []TileImage) *image.RGBA {
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