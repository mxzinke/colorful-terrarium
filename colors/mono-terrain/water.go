package mono_terrain

import (
	"context"
	"image"
	"io"

	"github.com/mxzinke/colorful-terrarium/colors"
)

type WaterMonoTerrainProfile struct {
}

func NewWaterMonoTerrainProfile() *WaterMonoTerrainProfile {
	return &WaterMonoTerrainProfile{}
}

func (w *WaterMonoTerrainProfile) Name() string {
	return "mono-terrain-water"
}

func (w *WaterMonoTerrainProfile) FileType() string {
	return "png"
}

func (w *WaterMonoTerrainProfile) MaxZoom() uint32 {
	return 14
}

func (w *WaterMonoTerrainProfile) GetImage(ctx context.Context, imgRect image.Rectangle, input colors.ColorInput) (image.Image, error) {
	output := image.NewGray16(imgRect)

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
				output.Set(x, y, EncodeElevationToMonoTerrain(float32(cell.Elevation())))
			}
		}
	}
	return output, nil
}

func (water *WaterMonoTerrainProfile) EncodeImage(w io.Writer, image image.Image) error {
	return noCompressionEncoder().Encode(w, image)
}
