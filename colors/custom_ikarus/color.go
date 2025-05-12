package custom_ikarus

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
			{Elevation: -2500, Color: colors.Color{R: 219, G: 229, B: 233, A: 255}},
			{Elevation: -2000, Color: colors.Color{R: 218, G: 232, B: 234, A: 255}},
			{Elevation: -1000, Color: colors.Color{R: 239, G: 240, B: 241, A: 255}},
			{Elevation: -200, Color: colors.Color{R: 235, G: 241, B: 242, A: 255}},
			{Elevation: 0, Color: colors.Color{R: 246, G: 251, B: 253, A: 255}}, // Coastal water
		},
	}

	normalPalette = colors.ColorPalette{
		Stops: []colors.ColorStop{
			{Elevation: -100, Color: colors.Color{R: 242, G: 215, B: 158, A: 255}},
			{Elevation: 0, Color: colors.Color{R: 246, G: 240, B: 157, A: 255}},
			{Elevation: 500, Color: colors.Color{R: 251, G: 223, B: 175, A: 255}},
			{Elevation: 700, Color: colors.Color{R: 249, G: 238, B: 166, A: 255}},
			{Elevation: 1700, Color: colors.Color{R: 233, G: 188, B: 126, A: 255}},
			{Elevation: 3000, Color: colors.Color{R: 205, G: 193, B: 126, A: 255}},
			{Elevation: 4200, Color: colors.Color{R: 222, G: 203, B: 173, A: 255}},
		},
	}

	polarPalette = colors.ColorPalette{
		Stops: []colors.ColorStop{
			{Elevation: 0, Color: colors.Color{R: 235, G: 246, B: 250, A: 255}},   // Iced Coastline
			{Elevation: 50, Color: colors.Color{R: 228, G: 240, B: 245, A: 255}},  // Snow plains
			{Elevation: 100, Color: colors.Color{R: 250, G: 250, B: 250, A: 255}}, // Snow lowlands
		},
	}
)

type CustomerProvider struct {
}

func NewCustomerProvider() *CustomerProvider {
	return &CustomerProvider{}
}

func (p *CustomerProvider) Name() string {
	return "custom-ikarus"
}

func (p *CustomerProvider) FileType() string {
	return "png"
}

func (p *CustomerProvider) MaxZoom() uint32 {
	return 13
}

func (p *CustomerProvider) GetImage(ctx context.Context, imgRect image.Rectangle, input colors.ColorInput) (image.Image, error) {
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

func (p *CustomerProvider) EncodeImage(w io.Writer, img image.Image) error {
	return colors.EncodePNGOptimized(w, img)
}
