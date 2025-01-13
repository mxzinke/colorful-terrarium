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

const (
	polarStartLatitude    = 64
	polarAbsoluteLatitude = 70
	earthTilt             = 6
	snowBaseFactor        = 0.65
	lowestSnowFactor      = 0.1
)

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
			{2000, Color{148, 144, 139}},  // High mountains
			{2500, Color{130, 115, 95}},   // Very high mountains
			{3000, Color{240, 240, 240}},  // Alpine/Snow transition
			{4000, Color{255, 255, 255}},  // Permanent snow
		},
	}

	polarPalette = ColorPalette{
		stops: []ColorStop{
			{500, Color{245, 251, 255}},  // Iced Water
			{0, Color{245, 248, 250}},    // Iced Coastline
			{50, Color{250, 250, 250}},   // Snow plains
			{200, Color{245, 245, 245}},  // Snow lowlands
			{400, Color{240, 240, 240}},  // Snow hills
			{700, Color{242, 242, 245}},  // Snow mountains
			{1000, Color{227, 227, 227}}, // Deep snow mountains
			{1500, Color{235, 235, 235}}, // High snow
			{2000, Color{238, 238, 238}}, // Alpine snow
			{2500, Color{242, 242, 242}}, // Permanent snow
			{3000, Color{255, 255, 255}}, // High permanent snow
		},
	}
)

// getColorForElevationAndTerrain returns interpolated color based on elevation and latitude
func getColorForElevationAndTerrain(elevation, polarFactor float64, hasIce bool) Color {
	if elevation <= 0 && !hasIce {
		return getColorFromPalette(elevation, normalPalette)
	}

	// If the elevation is below 100 and the location is in ice, use the polar palette
	if hasIce {
		polarColor := getColorFromPalette(elevation, polarPalette)
		return polarColor
	}

	// Get colors from both palette
	normalColor := getColorFromPalette(elevation, normalPalette)
	polarColor := getColorFromPalette(elevation, polarPalette)

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
