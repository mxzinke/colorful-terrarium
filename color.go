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
	{-10000, Color{38, 116, 184}}, // Shallow ocean
	{-1000, Color{65, 146, 208}},  // Deep ocean - helleres Blau
	{-500, Color{89, 171, 227}},   // Medium depth ocean - noch heller
	{-200, Color{109, 187, 239}},  // Shallow ocean - sehr hell
	{-50, Color{170, 218, 252}},   // Very shallow water - fast weiß-blau
	{-1, Color{191, 228, 252}},    // Coastal water - hellster Übergang
	{0, Color{172, 208, 165}},     // Coastline
	{100, Color{148, 191, 139}},   // Coastal plains
	{300, Color{168, 198, 143}},   // Lowlands
	{600, Color{189, 204, 150}},   // Hills
	{1000, Color{195, 182, 157}},  // Low mountains
	{1500, Color{168, 154, 134}},  // Medium mountains
	{2000, Color{137, 125, 107}},  // High mountains
	{2500, Color{130, 115, 95}},   // Very high mountains
	{3000, Color{210, 200, 190}},  // Alpine/Snow transition
	{4000, Color{255, 255, 255}},  // Permanent snow
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

	// Calculate base interpolation factor
	factor := (elevation - lowStop.Elevation) / (highStop.Elevation - lowStop.Elevation)

	// Apply smoothstep function for smoother transitions
	factor = factor * factor * (3 - 2*factor)

	// Ensure factor stays within bounds
	factor = math.Max(0, math.Min(1, factor))

	// Calculate each color component separately with gamma correction
	gamma := 2.2
	r := math.Pow(factor*math.Pow(float64(highStop.Color.R)/255, gamma)+
		(1-factor)*math.Pow(float64(lowStop.Color.R)/255, gamma), 1/gamma) * 255
	g := math.Pow(factor*math.Pow(float64(highStop.Color.G)/255, gamma)+
		(1-factor)*math.Pow(float64(lowStop.Color.G)/255, gamma), 1/gamma) * 255
	b := math.Pow(factor*math.Pow(float64(highStop.Color.B)/255, gamma)+
		(1-factor)*math.Pow(float64(lowStop.Color.B)/255, gamma), 1/gamma) * 255

	return Color{
		R: uint8(math.Round(math.Max(0, math.Min(255, r)))),
		G: uint8(math.Round(math.Max(0, math.Min(255, g)))),
		B: uint8(math.Round(math.Max(0, math.Min(255, b)))),
	}
}
