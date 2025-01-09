package main

import "math"

// Color represents an RGB color
type Color struct {
	R, G, B uint8
}

// ColorStop defines a color at a specific elevation
type ColorStop struct {
	Elevation float64
	Color     Color
}

// Define color palette with more natural colors and smoother transitions
var colorPalette = []ColorStop{
	{-1000, Color{21, 42, 103}},  // Deep ocean
	{-500, Color{25, 68, 132}},   // Medium depth ocean
	{-200, Color{29, 99, 171}},   // Shallow ocean
	{-50, Color{34, 139, 204}},   // Very shallow water
	{-1, Color{158, 220, 233}},   // Coastal water
	{0, Color{172, 208, 165}},    // Coastline
	{50, Color{148, 191, 139}},   // Coastal plains
	{200, Color{168, 198, 143}},  // Lowlands
	{400, Color{189, 204, 150}},  // Hills
	{700, Color{199, 193, 154}},  // Low mountains
	{1000, Color{195, 182, 157}}, // Medium mountains
	{1500, Color{168, 154, 134}}, // High mountains
	{2000, Color{145, 123, 111}}, // Alpine
	{2500, Color{215, 210, 203}}, // Snow line
	{3000, Color{250, 250, 250}}, // Permanent snow
}

// getColorForElevation returns interpolated color for given elevation
func getColorForElevation(elevation float64) Color {
	// Find color stops for interpolation
	var lowStop, highStop ColorStop

	// Handle elevation below or above range with smooth transitions
	if elevation <= colorPalette[0].Elevation {
		factor := math.Min(1.0, (colorPalette[0].Elevation-elevation)/1000.0)
		return Color{
			R: uint8(float64(colorPalette[0].Color.R) * (1 - factor*0.5)),
			G: uint8(float64(colorPalette[0].Color.G) * (1 - factor*0.5)),
			B: uint8(float64(colorPalette[0].Color.B) * (1 - factor*0.5)),
		}
	}
	if elevation >= colorPalette[len(colorPalette)-1].Elevation {
		excess := elevation - colorPalette[len(colorPalette)-1].Elevation
		lastColor := colorPalette[len(colorPalette)-1].Color
		factor := math.Min(1.0, excess/1000.0)
		return Color{
			R: uint8(math.Min(255, float64(lastColor.R)+(255-float64(lastColor.R))*factor)),
			G: uint8(math.Min(255, float64(lastColor.G)+(255-float64(lastColor.G))*factor)),
			B: uint8(math.Min(255, float64(lastColor.B)+(255-float64(lastColor.B))*factor)),
		}
	}

	// Find appropriate color stops
	for i := 0; i < len(colorPalette)-1; i++ {
		if elevation >= colorPalette[i].Elevation && elevation < colorPalette[i+1].Elevation {
			lowStop = colorPalette[i]
			highStop = colorPalette[i+1]
			break
		}
	}

	// Calculate interpolation factor with smoothing
	factor := (elevation - lowStop.Elevation) / (highStop.Elevation - lowStop.Elevation)
	// Apply subtle smoothing using sine curve
	factor = (1 - math.Cos(factor*math.Pi)) / 2

	// Interpolate colors
	return Color{
		R: uint8(math.Round(float64(lowStop.Color.R) + factor*float64(highStop.Color.R-lowStop.Color.R))),
		G: uint8(math.Round(float64(lowStop.Color.G) + factor*float64(highStop.Color.G-lowStop.Color.G))),
		B: uint8(math.Round(float64(lowStop.Color.B) + factor*float64(highStop.Color.B-lowStop.Color.B))),
	}
}
