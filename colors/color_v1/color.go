package color_v1

import (
	"context"
	"fmt"
	"image"
	"io"
	"math"

	colors "github.com/mxzinke/colorful-terrarium/colors"
)

var (
	waterPalette = colors.ColorPalette{
		Stops: []colors.ColorStop{
			{Elevation: -10000, Color: colors.Color{R: 65, G: 146, B: 208, A: 255}}, // Shallow ocean
			{Elevation: -1000, Color: colors.Color{R: 87, G: 172, B: 230, A: 255}},  // Deep ocean
			{Elevation: -500, Color: colors.Color{R: 96, G: 178, B: 235, A: 255}},   // Medium depth ocean
			{Elevation: -200, Color: colors.Color{R: 109, G: 187, B: 239, A: 255}},  // Shallow ocean
			{Elevation: -80, Color: colors.Color{R: 125, G: 197, B: 245, A: 255}},   // Very Shallow ocean
			{Elevation: -40, Color: colors.Color{R: 170, G: 218, B: 252, A: 255}},   // Shallow water
			{Elevation: -20, Color: colors.Color{R: 173, G: 216, B: 247, A: 255}},   // Very shallow water
			{Elevation: 0, Color: colors.Color{R: 191, G: 228, B: 252, A: 255}},     // Coastal water
		},
	}

	normalPalette = colors.ColorPalette{
		Stops: []colors.ColorStop{
			{Elevation: 0, Color: colors.Color{R: 172, G: 208, B: 165, A: 255}},    // Coastline
			{Elevation: 50, Color: colors.Color{R: 148, G: 191, B: 139, A: 255}},   // Coastal plains
			{Elevation: 100, Color: colors.Color{R: 148, G: 191, B: 139, A: 255}},  // Coastal plains
			{Elevation: 300, Color: colors.Color{R: 168, G: 198, B: 143, A: 255}},  // Lowlands
			{Elevation: 600, Color: colors.Color{R: 189, G: 204, B: 150, A: 255}},  // Hills
			{Elevation: 1000, Color: colors.Color{R: 195, G: 182, B: 157, A: 255}}, // Low mountains
			{Elevation: 1500, Color: colors.Color{R: 168, G: 154, B: 134, A: 255}}, // Medium mountains
			{Elevation: 2000, Color: colors.Color{R: 148, G: 144, B: 139, A: 255}}, // High mountains
			{Elevation: 2500, Color: colors.Color{R: 130, G: 115, B: 95, A: 255}},  // Very high mountains
			{Elevation: 3000, Color: colors.Color{R: 240, G: 240, B: 240, A: 255}}, // Alpine/Snow transition
			{Elevation: 4000, Color: colors.Color{R: 255, G: 255, B: 255, A: 255}}, // Permanent snow
		},
	}

	polarPalette = colors.ColorPalette{
		Stops: []colors.ColorStop{
			{Elevation: -500, Color: colors.Color{R: 242, G: 248, B: 250, A: 255}}, // Iced Water
			{Elevation: 0, Color: colors.Color{R: 235, G: 246, B: 250, A: 255}},    // Iced Coastline
			{Elevation: 50, Color: colors.Color{R: 228, G: 240, B: 245, A: 255}},   // Snow plains
			{Elevation: 200, Color: colors.Color{R: 225, G: 234, B: 237, A: 255}},  // Snow lowlands
			{Elevation: 400, Color: colors.Color{R: 211, G: 221, B: 222, A: 255}},  // Snow hills
			{Elevation: 700, Color: colors.Color{R: 218, G: 228, B: 230, A: 255}},  // Snow mountains
			{Elevation: 1000, Color: colors.Color{R: 217, G: 221, B: 222, A: 255}}, // Deep snow mountains
			{Elevation: 1500, Color: colors.Color{R: 227, G: 231, B: 232, A: 255}}, // High snow
			{Elevation: 2000, Color: colors.Color{R: 233, G: 238, B: 240, A: 255}}, // Alpine snow
			{Elevation: 2500, Color: colors.Color{R: 237, G: 243, B: 245, A: 255}}, // Permanent snow
			{Elevation: 3000, Color: colors.Color{R: 245, G: 251, B: 252, A: 255}}, // High permanent snow
		},
	}

	desertPalette = colors.ColorPalette{
		Stops: []colors.ColorStop{
			{Elevation: 0, Color: colors.Color{R: 235, G: 230, B: 185, A: 255}},    // Beach
			{Elevation: 300, Color: colors.Color{R: 209, G: 199, B: 159, A: 255}},  // Lowlands
			{Elevation: 600, Color: colors.Color{R: 189, G: 170, B: 134, A: 255}},  // Hills
			{Elevation: 1500, Color: colors.Color{R: 168, G: 154, B: 134, A: 255}}, // Medium mountains
			{Elevation: 2000, Color: colors.Color{R: 148, G: 144, B: 139, A: 255}}, // High mountains
			{Elevation: 2500, Color: colors.Color{R: 130, G: 115, B: 95, A: 255}},  // Very high mountains
			{Elevation: 3000, Color: colors.Color{R: 240, G: 240, B: 240, A: 255}}, // Alpine/Snow transition
			{Elevation: 4000, Color: colors.Color{R: 255, G: 255, B: 255, A: 255}}, // Permanent snow
		},
	}
)

type ColorV1Provider struct {
}

func NewColorV1Provider() *ColorV1Provider {
	return &ColorV1Provider{}
}

func (p *ColorV1Provider) Name() string {
	return "color-v1"
}

func (p *ColorV1Provider) FileType() string {
	return "png"
}

func (p *ColorV1Provider) MaxZoom() uint32 {
	return 13
}

func (p *ColorV1Provider) GetColor(ctx context.Context, input colors.ColorInput) ([][]colors.Color, error) {
	if input.Zoom > p.MaxZoom() {
		return nil, fmt.Errorf("zoom level %d is not supported", input.Zoom)
	}

	if len(input.DataMap) == 0 {
		return nil, fmt.Errorf("data map is empty")
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	colorsOutput := make([][]colors.Color, len(input.DataMap))
	for i, row := range input.DataMap {
		colorsOutput[i] = make([]colors.Color, len(row))

		for j, cell := range row {
			elevation := cell.Elevation()

			if !cell.IsLand() {
				if !cell.IsIce() {
					colorsOutput[i][j] = colors.GetColorFromPalette(elevation, waterPalette)
				} else {
					colorsOutput[i][j] = colors.GetColorFromPalette(elevation, polarPalette)
				}
				continue
			}

			snowThreadsholdFactor := math.Max(0.05, math.Pow(cell.AquatorFactor()/0.7, 1.5))
			elevation = elevation * float32(snowThreadsholdFactor)

			polarFactor := cell.PolarFactor()
			if polarFactor == 1 {
				colorsOutput[i][j] = colors.GetColorFromPalette(elevation, polarPalette)
				continue
			} else if polarFactor > 0 {
				polarColor := colors.GetColorFromPalette(elevation, polarPalette)

				if cell.IsIce() {
					colorsOutput[i][j] = polarColor
					continue
				}

				normalColor := colors.GetColorFromPalette(elevation, normalPalette)

				// Interpolate between normal and polar colors
				colorsOutput[i][j] = colors.Color{
					R: uint8(math.Round(float64(normalColor.R)*(1-float64(polarFactor)) + float64(polarColor.R)*float64(polarFactor))),
					G: uint8(math.Round(float64(normalColor.G)*(1-float64(polarFactor)) + float64(polarColor.G)*float64(polarFactor))),
					B: uint8(math.Round(float64(normalColor.B)*(1-float64(polarFactor)) + float64(polarColor.B)*float64(polarFactor))),
					A: uint8(math.Round(float64(normalColor.A)*(1-float64(polarFactor)) + float64(polarColor.A)*float64(polarFactor))),
				}
				continue
			}

			desertFactor := cell.DesertFactor()
			if desertFactor == 1 {
				colorsOutput[i][j] = colors.GetColorFromPalette(elevation, desertPalette)
				continue
			} else if desertFactor > 0 {
				normalColor := colors.GetColorFromPalette(elevation, normalPalette)
				desertColor := colors.GetColorFromPalette(elevation, desertPalette)

				colorsOutput[i][j] = colors.Color{
					R: uint8(math.Round(float64(normalColor.R)*(1-float64(desertFactor)) + float64(desertColor.R)*float64(desertFactor))),
					G: uint8(math.Round(float64(normalColor.G)*(1-float64(desertFactor)) + float64(desertColor.G)*float64(desertFactor))),
					B: uint8(math.Round(float64(normalColor.B)*(1-float64(desertFactor)) + float64(desertColor.B)*float64(desertFactor))),
					A: uint8(math.Round(float64(normalColor.A)*(1-float64(desertFactor)) + float64(desertColor.A)*float64(desertFactor))),
				}
				continue
			}

			if cell.IsIce() {
				colorsOutput[i][j] = colors.GetColorFromPalette(elevation, polarPalette)
				continue
			}

			colorsOutput[i][j] = colors.GetColorFromPalette(elevation, normalPalette)
		}
	}

	return colorsOutput, nil
}

func (p *ColorV1Provider) EncodeImage(w io.Writer, img image.Image) error {
	return colors.EncodePNGOptimized(w, img)
}
