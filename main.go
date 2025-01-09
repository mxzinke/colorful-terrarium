package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
)

func processAndColorize(img image.Image, minLat, maxLat float64) *image.RGBA {
	bounds := img.Bounds()
	output := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		// Calculate precise latitude for this pixel row
		latitude := getLatitudeForPixel(y, minLat, maxLat, bounds.Max.Y)

		log.Printf("%d => latitude: %f\n", y, latitude)

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
	if len(os.Args) != 5 {
		log.Fatal("Usage: ./terrain-downloader z x y output.png")
	}

	var z, y, x uint32
	fmt.Sscanf(os.Args[1], "%d", &z)
	fmt.Sscanf(os.Args[2], "%d", &y)
	fmt.Sscanf(os.Args[3], "%d", &x)
	outputPath := os.Args[4]

	log.Printf("z: %d y: %d x: %d\n", z, y, x)

	// Download subtiles
	tiles, err := downloadSubTiles(z, y, x)
	if err != nil {
		log.Fatal(err)
	}

	// Combine tiles
	combined := compositeImages(tiles)

	// Get latitudes (min and max) for the tile
	minLat, maxLat := getTileLatitudes(z, y, x)

	// Process and colorize
	processed := processAndColorize(combined, minLat, maxLat)

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

// getTileLatitudes calculates the minimum and maximum latitudes for a given tile
func getTileLatitudes(z, y, x uint32) (minLat, maxLat float64) {
	n := math.Pi - 2.0*math.Pi*float64(y)/math.Pow(2.0, float64(z))
	maxLat = math.Atan(math.Sinh(n)) * 180.0 / math.Pi

	n = math.Pi - 2.0*math.Pi*float64(y+1)/math.Pow(2.0, float64(z))
	minLat = math.Atan(math.Sinh(n)) * 180.0 / math.Pi

	if (x%2 != 0 && y%2 == 0) || (x%2 == 0 && y%2 != 0) {
		return minLat, maxLat
	}

	// The other way around for odd tiles
	return maxLat, minLat
}

// getLatitudeForPixel calculates the latitude for a specific pixel y-coordinate within a tile
func getLatitudeForPixel(pixelY int, minLat, maxLat float64, tileSize int) float64 {
	// Convert pixel position to normalized position within the tile (0 to 1)
	normalizedY := float64(pixelY) / float64(tileSize-1)

	// Mercator projection is non-linear, so we need to convert to radians,
	// interpolate in projected space, then convert back
	maxLatRad := maxLat * math.Pi / 180.0
	minLatRad := minLat * math.Pi / 180.0

	// Project to mercator y coordinate
	maxY := math.Log(math.Tan(math.Pi/4 + maxLatRad/2))
	minY := math.Log(math.Tan(math.Pi/4 + minLatRad/2))

	// Interpolate in projected space
	y := minY + normalizedY*(maxY-minY)

	// Convert back to latitude
	return (2*math.Atan(math.Exp(y)) - math.Pi/2) * 180.0 / math.Pi
}
