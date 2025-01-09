package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
)

func processAndColorize(img image.Image) *image.RGBA {
	bounds := img.Bounds()
	output := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()

			// Convert 16-bit color values to 8-bit
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)

			// Calculate elevation using Terrarium formula
			elevation := float64(r8)*256 + float64(g8) + float64(b8)/256 - 32768

			// Get color for elevation
			newColor := getColorForElevation(elevation)
			output.Set(x, y, color.RGBA{newColor.R, newColor.G, newColor.B, 255})
		}
	}

	return output
}

func main() {
	if len(os.Args) != 5 {
		log.Fatal("Usage: ./terrain-downloader z x y output.png")
	}

	var z, x, y int
	fmt.Sscanf(os.Args[1], "%d", &z)
	fmt.Sscanf(os.Args[2], "%d", &x)
	fmt.Sscanf(os.Args[3], "%d", &y)
	outputPath := os.Args[4]

	// Download subtiles
	tiles, err := downloadSubTiles(z, x, y)
	if err != nil {
		log.Fatal(err)
	}

	// Combine tiles
	combined := compositeImages(tiles)

	// Process and colorize
	processed := processAndColorize(combined)

	// Save output
	outputFile, err := os.Create(outputPath)
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	if err := png.Encode(outputFile, processed); err != nil {
		log.Fatal(err)
	}
}
