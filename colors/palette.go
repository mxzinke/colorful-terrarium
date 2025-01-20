package colors

import (
	img_color "image/color"
)

// Color represents an RGB color
type Color struct {
	R, G, B uint8
	A       uint8
}

func (c Color) RGBA() img_color.RGBA {
	return img_color.RGBA{c.R, c.G, c.B, c.A}
}

// ColorStop defines a color at a specific elevation
type ColorStop struct {
	Elevation float32
	Color     Color
}

// ColorPalette represents a complete set of elevation-based colors
type ColorPalette struct {
	Stops []ColorStop
}
