package terrain

// ElevationMap holds preprocessed elevation data for efficient access
type ElevationMap struct {
	Data     [][]float32
	TileSize int
}

// GetElevation returns the elevation at the given coordinates
// Returns 0 (sea level) for out of bounds coordinates
func (em *ElevationMap) GetElevation(x, y int) float32 {
	if x < 0 || y < 0 || x >= em.TileSize || y >= em.TileSize {
		return 0
	}
	return em.Data[y][x]
}

// ModifyElevation modifies the elevation at the given coordinates (pixels)
func (em *ElevationMap) ModifyElevation(x, y int, elevation float32) {
	em.Data[y][x] = elevation
}

// IsLand returns true if the elevation indicates land
func (em *ElevationMap) IsLand(elevation float32) bool {
	return elevation > 0
}

// GetNeighborhood returns elevation values in a square neighborhood
func (em *ElevationMap) GetNeighborhood(x, y, radius int) []float32 {
	size := 2*radius + 1
	result := make([]float32, 0, size*size)

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
			if newX < 0 || newY < 0 || newX >= em.TileSize || newY >= em.TileSize {
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
