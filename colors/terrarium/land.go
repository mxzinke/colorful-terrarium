package terrarium

import (
	"context"
	"image"
	"io"

	"github.com/mxzinke/colorful-terrarium/colors"
)

type LandTerrariumProfile struct {
}

func NewLandTerrariumProfile() *LandTerrariumProfile {
	return &LandTerrariumProfile{}
}

func (l *LandTerrariumProfile) Name() string {
	return "terrarium-land"
}

func (l *LandTerrariumProfile) FileType() string {
	return "png"
}

func (l *LandTerrariumProfile) MaxZoom() uint32 {
	return 14
}

func (l *LandTerrariumProfile) GetImage(ctx context.Context, imgRect image.Rectangle, input colors.ColorInput) (image.Image, error) {
	output := image.NewNRGBA(imgRect)

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
				output.Set(x, y, encodeElevationToTerrarium(float64(cell.Elevation())))
			}
		}
	}
	return output, nil
}

func (l *LandTerrariumProfile) EncodeImage(w io.Writer, image image.Image) error {
	return noCompressionEncoder().Encode(w, image)
}
