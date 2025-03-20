package colors

import (
	"context"
	"image"
	"io"
)

type DataCell interface {
	// Elevation is the elevation of the cell in meters -12000 to 9000 can be expected
	Elevation() float32
	// IsLand is true if the cell is part of the landmass
	IsLand() bool
	// IsIce is true if the cell has ice on it
	IsIce() bool
	// DesertFactor is a value between 0 and 1, where 1 is desert and 0 is normal land
	DesertFactor() float64
	// PolarFactor is a value between 0 and 1, where 1 is polar and 0 is non-polar
	PolarFactor() float64
	// AquatorFactor is a value between 0 and 1, where 1 is aquator and 0 is polar
	AquatorFactor() float64
}

// ColorInput is the input for the GetColor handler
type ColorInput struct {
	Zoom uint32
	// DataMap is a 2D slice of DataCell where each element is a cell
	DataMap [][]DataCell
}

// ColorProvider is the interface for to handle color generation
type ColorProvider interface {
	// Name is the name of the provider
	Name() string
	// MaxZoom is the maximum zoom level that this provider can handle
	MaxZoom() uint32
	// GetImage returns the color image for each cell (for a given input)
	GetImage(ctx context.Context, imgRect image.Rectangle, input ColorInput) (image.Image, error)
	// FileType returns the file type of the image (e.g. "png", "jpg", "webp", etc.)
	FileType() string
	// EncodeImage encodes the final image (e.g. to PNG or else) and returns the file type or error
	EncodeImage(w io.Writer, img image.Image) error
}
