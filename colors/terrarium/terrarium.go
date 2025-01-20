package terrarium

import (
	"image/png"
	"math"

	"github.com/mxzinke/colorful-terrarium/colors"
)

var (
	ZeroElevation = colors.Color{R: 128, G: 0, B: 0, A: 255}
	IceElevation  = colors.Color{R: 128, G: 3, B: 0, A: 255}
)

func noCompressionEncoder() *png.Encoder {
	return &png.Encoder{
		CompressionLevel: png.NoCompression, // Maximum compression
		BufferPool:       nil,               // Use default buffer pool
	}
}

func encodeElevationToTerrarium(elevation float64) colors.Color {
	v := elevation + 32768
	return colors.Color{
		R: uint8(math.Floor(v / 256.0)),
		G: uint8(math.Floor(math.Mod(v, 256.0))),
		B: uint8(math.Floor((v - math.Floor(v)) * 256)),
		A: 255,
	}
}
