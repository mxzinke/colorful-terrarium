package terrarium

import (
	img_color "image/color"
	"image/png"
	"math"
)

var (
	ZeroElevation = img_color.NRGBA{128, 0, 0, 255}
	IceElevation  = img_color.NRGBA{128, 3, 0, 255}
)

func noCompressionEncoder() *png.Encoder {
	return &png.Encoder{
		CompressionLevel: png.NoCompression, // Maximum compression
		BufferPool:       nil,               // Use default buffer pool
	}
}

func encodeElevationToTerrarium(elevation float64) img_color.NRGBA {
	v := elevation + 32768
	return img_color.NRGBA{
		R: uint8(math.Floor(v / 256.0)),
		G: uint8(math.Floor(math.Mod(v, 256.0))),
		B: uint8(math.Floor((v - math.Floor(v)) * 256)),
		A: 255,
	}
}
