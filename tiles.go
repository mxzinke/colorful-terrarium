package main

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

const baseURL = "https://terrain.mapstudio.ai/%d/%d/%d.png"
const tileSize = 256 // Standard tile size

type TileCoord struct {
	Z, X, Y uint32
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

func downloadSubTiles(parentZ, parentX, parentY uint32) ([]TileImage, error) {
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
