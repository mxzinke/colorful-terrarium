package terrarium

import (
	"context"
	"image"
	"io"

	"github.com/mxzinke/colorful-terrarium/colors"
)

type WaterTerrariumProfile struct {
}

func NewWaterTerrariumProfile() *WaterTerrariumProfile {
	return &WaterTerrariumProfile{}
}

func (w *WaterTerrariumProfile) Name() string {
	return "terrarium-water"
}

func (w *WaterTerrariumProfile) FileType() string {
	return "png"
}

func (w *WaterTerrariumProfile) MaxZoom() uint32 {
	return 14
}

func (w *WaterTerrariumProfile) GetImage(ctx context.Context, imgRect image.Rectangle, input colors.ColorInput) (image.Image, error) {
	output := image.NewNRGBA(imgRect)

	for y, row := range input.DataMap {

		for x, cell := range row {
			if cell.IsLand() {
				output.Set(x, y, ZeroElevation)
				continue
			}

			if cell.IsIce() {
				output.Set(x, y, IceElevation)
			}

			if cell.Elevation() == 0 {
				output.Set(x, y, ZeroElevation)
			} else {
				output.Set(x, y, encodeElevationToTerrarium(float64(cell.Elevation())))
			}
		}
	}
	return output, nil
}

func (water *WaterTerrariumProfile) EncodeImage(w io.Writer, image image.Image) error {
	return noCompressionEncoder().Encode(w, image)
}
