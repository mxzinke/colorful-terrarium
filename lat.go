package main

import (
	"math"
)

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
