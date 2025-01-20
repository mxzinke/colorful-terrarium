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

func (l *LandTerrariumProfile) GetColor(ctx context.Context, input colors.ColorInput) ([][]colors.Color, error) {
	terrariumColors := make([][]colors.Color, len(input.DataMap))
	for i, row := range input.DataMap {
		terrariumColors[i] = make([]colors.Color, len(row))

		for j, cell := range row {
			if !cell.IsLand() {
				if cell.IsIce() {
					terrariumColors[i][j] = IceElevation
				} else {
					terrariumColors[i][j] = ZeroElevation
				}
				continue
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

func (l *LandTerrariumProfile) EncodeImage(w io.Writer, image image.Image) error {
	return noCompressionEncoder().Encode(w, image)
}
