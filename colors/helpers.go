package colors

import (
	"image"
	"image/png"
	"io"
	"math"
)

// GetColorFromPalette returns a color from a palette based on the elevation. This is a simple linear interpolation between the two closest color stops.
// It also applies a smoothstep function to the interpolation factor to create a smoother transition between colors.
// This function can be used as a helper function for any color palette.
func GetColorFromPalette(elevation float32, palette ColorPalette) Color {
	if elevation <= palette.Stops[0].Elevation {
		return palette.Stops[0].Color
	}
	if elevation >= palette.Stops[len(palette.Stops)-1].Elevation {
		return palette.Stops[len(palette.Stops)-1].Color
	}

	// Find color stops for interpolation
	var lowStop, highStop ColorStop

	// Find appropriate color stops
	for i := 0; i < len(palette.Stops)-1; i++ {
		if elevation >= palette.Stops[i].Elevation && elevation < palette.Stops[i+1].Elevation {
			lowStop = palette.Stops[i]
			highStop = palette.Stops[i+1]
			break
		}
	}

	// Calculate base interpolation factor
	factor := (elevation - lowStop.Elevation) / (highStop.Elevation - lowStop.Elevation)

	// Apply smoothstep function for smoother transitions
	factor = factor * factor * (3 - 2*factor)

	// Ensure factor stays within bounds
	factor = float32(math.Max(0, math.Min(1, float64(factor))))

	// Calculate each color component separately with gamma correction
	gamma := 2.2
	r := math.Pow(float64(factor)*math.Pow(float64(highStop.Color.R)/255, gamma)+
		(1-float64(factor))*math.Pow(float64(lowStop.Color.R)/255, gamma), 1/gamma) * 255
	g := math.Pow(float64(factor)*math.Pow(float64(highStop.Color.G)/255, gamma)+
		(1-float64(factor))*math.Pow(float64(lowStop.Color.G)/255, gamma), 1/gamma) * 255
	b := math.Pow(float64(factor)*math.Pow(float64(highStop.Color.B)/255, gamma)+
		(1-float64(factor))*math.Pow(float64(lowStop.Color.B)/255, gamma), 1/gamma) * 255
	alpha := math.Pow(float64(factor)*math.Pow(float64(highStop.Color.A)/255, gamma)+
		(1-float64(factor))*math.Pow(float64(lowStop.Color.A)/255, gamma), 1/gamma) * 255

	return Color{
		R: uint8(math.Round(math.Max(0, math.Min(255, r)))),
		G: uint8(math.Round(math.Max(0, math.Min(255, g)))),
		B: uint8(math.Round(math.Max(0, math.Min(255, b)))),
		A: uint8(math.Round(math.Max(0, math.Min(255, alpha)))),
	}
}

// EncodePNGOptimized creates a PNG encoder with optimal compression settings
func EncodePNGOptimized(w io.Writer, img image.Image) error {
	encoder := &png.Encoder{
		CompressionLevel: png.NoCompression, // No compression => fast encoding
		BufferPool:       nil,               // Use default buffer pool
	}
	return encoder.Encode(w, img)
}
