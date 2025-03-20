package mono_terrain

import (
	"context"
	"image"
	"io"

	"github.com/mxzinke/colorful-terrarium/colors"
)

type LandMonoTerrainProfile struct {
}

func NewLandMonoTerrainProfile() *LandMonoTerrainProfile {
	return &LandMonoTerrainProfile{}
}

func (l *LandMonoTerrainProfile) Name() string {
	return "mono-terrain-land"
}

func (l *LandMonoTerrainProfile) FileType() string {
	return "png"
}

func (l *LandMonoTerrainProfile) MaxZoom() uint32 {
	return 14
}

func (l *LandMonoTerrainProfile) GetImage(ctx context.Context, imgRect image.Rectangle, input colors.ColorInput) (image.Image, error) {
	output := image.NewGray16(imgRect)

	for y, row := range input.DataMap {
		for x, cell := range row {
			if !cell.IsLand() {
				if cell.IsIce() {
					output.Set(x, y, IceElevation)
				} else {
					output.Set(x, y, ZeroElevation)
				}
				continue
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

func (l *LandMonoTerrainProfile) EncodeImage(w io.Writer, image image.Image) error {
	return noCompressionEncoder().Encode(w, image)
}
