package mono_terrain

import (
	img_color "image/color"
	"image/png"
	"math"
)

var (
	ZeroElevation = img_color.Gray16{30000}
	IceElevation  = img_color.Gray16{30012}
)

func noCompressionEncoder() *png.Encoder {
	return &png.Encoder{
		CompressionLevel: png.NoCompression, // Maximum compression
		BufferPool:       nil,               // Use default buffer pool
	}
}

const ElevationShift = 7500

func EncodeElevationToMonoTerrain(elevation float32) img_color.Gray16 {
	v := (elevation + ElevationShift) * 4
	return img_color.Gray16{uint16(math.Round(float64(v)))}
}

func DecodeElevationFromMonoTerrain(pixel img_color.Gray16) float32 {
	return (float32(pixel.Y) / 4.0) - ElevationShift
}
