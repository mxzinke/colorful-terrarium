package color_v2

import (
	"context"
	"fmt"
	"image"
	"io"

	colors "github.com/mxzinke/colorful-terrarium/colors"
)

var (
	waterPalette = colors.ColorPalette{
		Stops: []colors.ColorStop{
			{Elevation: -7000, Color: colors.Color{R: 69, G: 121, B: 180, A: 255}},
			{Elevation: -2500, Color: colors.Color{R: 133, G: 185, B: 228, A: 255}},
			{Elevation: -2000, Color: colors.Color{R: 141, G: 193, B: 234, A: 255}},
			{Elevation: -1500, Color: colors.Color{R: 149, G: 201, B: 240, A: 255}},
			{Elevation: -1000, Color: colors.Color{R: 161, G: 210, B: 247, A: 255}},
			{Elevation: -500, Color: colors.Color{R: 171, G: 219, B: 252, A: 255}},
			{Elevation: -200, Color: colors.Color{R: 185, G: 227, B: 255, A: 255}},
			{Elevation: -50, Color: colors.Color{R: 200, G: 234, B: 255, A: 255}},
			{Elevation: 0, Color: colors.Color{R: 216, G: 242, B: 254, A: 255}}, // Coastal water
		},
	}

	normalPalette = colors.ColorPalette{
		Stops: []colors.ColorStop{
			{Elevation: 0, Color: colors.Color{R: 172, G: 208, B: 165, A: 255}},
			{Elevation: 100, Color: colors.Color{R: 148, G: 191, B: 139, A: 255}},
			{Elevation: 250, Color: colors.Color{R: 168, G: 198, B: 143, A: 255}},
			{Elevation: 500, Color: colors.Color{R: 189, G: 204, B: 150, A: 255}},
			{Elevation: 750, Color: colors.Color{R: 209, G: 215, B: 171, A: 255}},
			{Elevation: 1250, Color: colors.Color{R: 239, G: 235, B: 192, A: 255}},
			{Elevation: 1500, Color: colors.Color{R: 222, G: 214, B: 163, A: 255}},
			{Elevation: 2000, Color: colors.Color{R: 211, G: 202, B: 157, A: 255}},
			{Elevation: 2500, Color: colors.Color{R: 202, G: 185, B: 130, A: 255}},
			{Elevation: 3000, Color: colors.Color{R: 192, G: 154, B: 83, A: 255}},
			{Elevation: 5000, Color: colors.Color{R: 168, G: 120, B: 62, A: 255}},
			{Elevation: 6500, Color: colors.Color{R: 133, G: 100, B: 50, A: 255}},
			{Elevation: 8000, Color: colors.Color{R: 100, G: 70, B: 30, A: 255}},
		},
	}

	polarPalette = colors.ColorPalette{
		Stops: []colors.ColorStop{
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
)

type ColorV2Provider struct {
}

func NewColorV2Provider() *ColorV2Provider {
	return &ColorV2Provider{}
}

func (p *ColorV2Provider) Name() string {
	return "color-v2"
}

func (p *ColorV2Provider) FileType() string {
	return "png"
}

func (p *ColorV2Provider) MaxZoom() uint32 {
	return 13
}

func (p *ColorV2Provider) GetImage(ctx context.Context, imgRect image.Rectangle, input colors.ColorInput) (image.Image, error) {
	if input.Zoom > p.MaxZoom() {
		return nil, fmt.Errorf("zoom level %d is not supported", input.Zoom)
	}

	if len(input.DataMap) == 0 {
		return nil, fmt.Errorf("data map is empty")
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	output := image.NewNRGBA(imgRect)
	for y, row := range input.DataMap {
		for x, cell := range row {
			elevation := cell.Elevation()

			if !cell.IsLand() {
				if !cell.IsIce() {
					output.Set(x, y, colors.GetColorFromPalette(elevation, waterPalette).RGBA())
				} else {
					output.Set(x, y, colors.GetColorFromPalette(elevation, polarPalette).RGBA())
				}
				continue
			}

			if cell.IsIce() {
				output.Set(x, y, colors.GetColorFromPalette(elevation, polarPalette).RGBA())
				continue
			}

			output.Set(x, y, colors.GetColorFromPalette(elevation, normalPalette).RGBA())
		}
	}

	return output, nil
}

func (p *ColorV2Provider) EncodeImage(w io.Writer, img image.Image) error {
	return colors.EncodePNGOptimized(w, img)
}
