package main

import (
	"math"
)

// Color represents an RGB color
type Color struct {
	R, G, B uint8
}

// ColorStop defines a color at a specific elevation
type ColorStop struct {
	Elevation float64
	Color     Color
}

// ColorPalette represents a complete set of elevation-based colors
type ColorPalette struct {
	stops []ColorStop
}

var (
	normalPalette = ColorPalette{
		stops: []ColorStop{
			{-10000, Color{65, 146, 208}}, // Shallow ocean
			{-1000, Color{87, 172, 230}},  // Deep ocean
			{-500, Color{96, 178, 235}},   // Medium depth ocean
			{-200, Color{109, 187, 239}},  // Shallow ocean
			{-50, Color{170, 218, 252}},   // Very shallow water
			{0, Color{191, 228, 252}},     // Coastal water
			{0.1, Color{235, 230, 185}},   // Beach
			{3, Color{172, 208, 165}},     // Coastline
			{50, Color{148, 191, 139}},    // Coastal plains
			{100, Color{148, 191, 139}},   // Coastal plains
			{300, Color{168, 198, 143}},   // Lowlands
			{600, Color{189, 204, 150}},   // Hills
			{1000, Color{195, 182, 157}},  // Low mountains
			{1500, Color{168, 154, 134}},  // Medium mountains
			{2000, Color{137, 125, 107}},  // High mountains
			{2500, Color{130, 115, 95}},   // Very high mountains
			{3000, Color{210, 200, 190}},  // Alpine/Snow transition
			{4000, Color{255, 255, 255}},  // Permanent snow
		},
	}

	polarPalette = ColorPalette{
		stops: []ColorStop{
			{-10000, Color{65, 146, 208}}, // Shallow ocean
			{-1000, Color{87, 172, 230}},  // Deep ocean
			{-500, Color{96, 178, 235}},   // Medium depth ocean
			{-200, Color{109, 187, 239}},  // Shallow ocean
			{-50, Color{170, 218, 252}},   // Very shallow water
			{1, Color{191, 228, 252}},     // Coastal water
			{3, Color{172, 208, 165}},     // Coastline
			{50, Color{250, 250, 250}},    // Snow plains
			{200, Color{245, 245, 245}},   // Snow lowlands
			{400, Color{240, 240, 240}},   // Snow hills
			{700, Color{235, 235, 235}},   // Snow mountains
			{1000, Color{230, 230, 230}},  // Deep snow mountains
			{1500, Color{225, 225, 225}},  // High snow
			{2000, Color{220, 220, 220}},  // Alpine snow
			{2500, Color{215, 215, 215}},  // Permanent snow
			{3000, Color{210, 210, 210}},  // High permanent snow
		},
	}
)

// getColorForElevationAndLatitude returns interpolated color based on elevation and latitude
func getColorForElevationAndLatitude(elevation, baseLatitude float64, isInIce bool) Color {
	if elevation <= 0 && !isInIce {
		return getColorFromPalette(elevation, normalPalette)
	}

	if elevation <= 100 && isInIce {
		return Color{255, 255, 255}
	}

	latitude := math.Abs(baseLatitude)

	// Calculate how "polar" the location is (0 = arctic circle, 1 = poles)
	polarFactor := math.Min(math.Max((latitude-66)/(80-66), 0), 1)
	snowThresholdFactor := math.Min(math.Max(latitude/66, 0.1), 1) / 0.65

	if baseLatitude < -66 {
		polarFactor = 1
	}

	if isInIce {
		snowThresholdFactor += 0.2
	}

	// Get colors from both palette
	normalColor := getColorFromPalette(elevation*snowThresholdFactor, normalPalette)
	polarColor := getColorFromPalette(elevation*snowThresholdFactor, polarPalette)

	// Interpolate between normal and polar colors
	return Color{
		R: uint8(float64(normalColor.R)*(1-polarFactor) + float64(polarColor.R)*polarFactor),
		G: uint8(float64(normalColor.G)*(1-polarFactor) + float64(polarColor.G)*polarFactor),
		B: uint8(float64(normalColor.B)*(1-polarFactor) + float64(polarColor.B)*polarFactor),
	}
}

func getColorFromPalette(elevation float64, palette ColorPalette) Color {
	// Find color stops for interpolation
	var lowStop, highStop ColorStop

	if elevation <= palette.stops[0].Elevation {
		return palette.stops[0].Color
	}
	if elevation >= palette.stops[len(palette.stops)-1].Elevation {
		return palette.stops[len(palette.stops)-1].Color
	}

	// Find appropriate color stops
	for i := 0; i < len(palette.stops)-1; i++ {
		if elevation >= palette.stops[i].Elevation && elevation < palette.stops[i+1].Elevation {
			lowStop = palette.stops[i]
			highStop = palette.stops[i+1]
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
