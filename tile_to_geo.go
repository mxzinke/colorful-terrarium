package main

import (
	"math"

	"github.com/paulmach/orb"
)

type TileBounds struct {
	Zoom    uint32
	TileX   uint32
	TileY   uint32
	MinLat  float64
	MaxLat  float64
	MinLon  float64
	MaxLon  float64
	xLookup []float64
	yLookup []float64
}

func CreateTileBounds(zoom, tileY, tileX uint32, tileSize int) *TileBounds {
	minLat, maxLat := getTileLatitudes(zoom, tileY)
	minLon, maxLon := getTileLongitudes(zoom, tileX)

	// Mercator projection is non-linear, so we need to convert to radians,
	// interpolate in projected space, then convert back
	maxLatRad := maxLat * math.Pi / 180.0
	minLatRad := minLat * math.Pi / 180.0

	// Project to mercator y coordinate
	maxY := math.Log(math.Tan(math.Pi/4 + maxLatRad/2))
	minY := math.Log(math.Tan(math.Pi/4 + minLatRad/2))
	deltaY := maxY - minY

	// Create the latitude lookup table
	var yLookup []float64 = make([]float64, tileSize)
	for pixelY := 0; pixelY < tileSize; pixelY++ {
		// Convert pixel position to normalized position within the tile (0 to 1)
		normalizedY := float64(pixelY) / float64(tileSize-1)

		// Interpolate in projected space
		y := minY + normalizedY*deltaY

		// Convert back to latitude
		yLookup[pixelY] = (2*math.Atan(math.Exp(y)) - math.Pi/2) * 180.0 / math.Pi
	}

	// Calculate the longitude delta
	deltaLon := maxLon - minLon

	// Create the longitude lookup table
	var xLookup []float64 = make([]float64, tileSize)
	for pixelX := 0; pixelX < tileSize; pixelX++ {
		// Convert pixel position to normalized position within the tile (0 to 1)
		normalizedX := float64(pixelX) / float64(tileSize)

		// Linear interpolation between minLon and maxLon
		xLookup[pixelX] = minLon + normalizedX*deltaLon
	}

	return &TileBounds{
		MinLat:  minLat,
		MaxLat:  maxLat,
		MinLon:  minLon,
		MaxLon:  maxLon,
		xLookup: xLookup,
		yLookup: yLookup,
	}
}

func (tb *TileBounds) GetPixelLat(y int) float64 {
	return tb.yLookup[y]
}

func (tb *TileBounds) GetPixelLng(x int) float64 {
	return tb.xLookup[x]
}

func (tb *TileBounds) Bound() orb.Bound {
	return orb.Bound{
		Min: orb.Point{tb.MinLon, tb.MinLat},
		Max: orb.Point{tb.MaxLon, tb.MaxLat},
	}
}

// getTileLatitudes calculates the minimum and maximum latitudes for a given tile
func getTileLatitudes(z, y uint32) (minLat, maxLat float64) {
	n := math.Pi - 2.0*math.Pi*float64(y)/math.Pow(2.0, float64(z))
	maxLat = math.Atan(math.Sinh(n)) * 180.0 / math.Pi

	n = math.Pi - 2.0*math.Pi*float64(y+1)/math.Pow(2.0, float64(z))
	minLat = math.Atan(math.Sinh(n)) * 180.0 / math.Pi

	return maxLat, minLat
}

// getTileLongitudes calculates the minimum and maximum longitudes for a given tile
func getTileLongitudes(z, x uint32) (minLon, maxLon float64) {
	// Each tile covers 360/2^zoom degrees
	tileWidth := 360.0 / math.Pow(2, float64(z))

	// Calculate longitude of western edge of tile
	minLon = float64(x)*tileWidth - 180.0

	// Calculate longitude of eastern edge of tile
	maxLon = minLon + tileWidth

	return minLon, maxLon
}
