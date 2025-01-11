package terrain

import (
	"image"
)

// ElevationMap holds preprocessed elevation data for efficient access
type ElevationMap struct {
	Data   [][]float64
	Bounds image.Rectangle
}

// NewElevationMap creates a new ElevationMap from an image
func NewElevationMap(img image.Image) *ElevationMap {
	bounds := img.Bounds()
	data := make([][]float64, bounds.Dy())

	for y := 0; y < bounds.Dy(); y++ {
		data[y] = make([]float64, bounds.Dx())
		for x := 0; x < bounds.Dx(); x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)
			data[y][x] = float64(r8)*256 + float64(g8) + float64(b8)/256 - 32768
		}
	}

	return &ElevationMap{
		Data:   data,
		Bounds: bounds,
	}
}

// GetElevation returns the elevation at the given coordinates
// Returns 0 (sea level) for out of bounds coordinates
func (em *ElevationMap) GetElevation(x, y int) float64 {
	if x < 0 || y < 0 || x >= em.Bounds.Dx() || y >= em.Bounds.Dy() {
		return 0
	}
	return em.Data[y][x]
}

func (em *ElevationMap) ModifyElevation(x, y int, elevation float64) {
	em.Data[y][x] = elevation
}

// IsLand returns true if the elevation indicates land
func (em *ElevationMap) IsLand(elevation float64) bool {
	return elevation > 0
}

// GetNeighborhood returns elevation values in a square neighborhood
func (em *ElevationMap) GetNeighborhood(x, y, radius int) []float64 {
	size := 2*radius + 1
	result := make([]float64, 0, size*size)

	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			result = append(result, em.GetElevation(x+dx, y+dy))
		}
	}

	return result
}

// GetNeighborhoodStats returns the count of land and water pixels in a neighborhood
// and a flag indicating if the neighborhood includes tile edges
func (em *ElevationMap) GetNeighborhoodStats(x, y, radius int) (landCount, waterCount int, hasEdge bool) {
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			newX, newY := x+dx, y+dy

			// Check if we're at the tile edge
			if newX < 0 || newY < 0 || newX >= em.Bounds.Dx() || newY >= em.Bounds.Dy() {
				hasEdge = true
				continue
			}

			elev := em.Data[newY][newX]
			if em.IsLand(elev) {
				landCount++
			} else {
				waterCount++
			}
		}
	}
	return landCount, waterCount, hasEdge
}
