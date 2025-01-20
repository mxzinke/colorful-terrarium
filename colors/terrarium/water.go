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

func (w *WaterTerrariumProfile) GetColor(ctx context.Context, input colors.ColorInput) ([][]colors.Color, error) {
	terrariumColors := make([][]colors.Color, len(input.DataMap))
	for i, row := range input.DataMap {
		terrariumColors[i] = make([]colors.Color, len(row))

		for j, cell := range row {
			if cell.IsLand() {
				terrariumColors[i][j] = ZeroElevation
				continue
			}

			if cell.IsIce() {
				terrariumColors[i][j] = IceElevation
			}

			if cell.Elevation() == 0 {
				terrariumColors[i][j] = ZeroElevation
			} else {
				terrariumColors[i][j] = encodeElevationToTerrarium(float64(cell.Elevation()))
			}
		}
	}
	return terrariumColors, nil
}

func (water *WaterTerrariumProfile) EncodeImage(w io.Writer, image image.Image) error {
	return noCompressionEncoder().Encode(w, image)
}
