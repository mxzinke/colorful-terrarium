package main

import (
	"image"
)

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
