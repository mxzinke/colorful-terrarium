package main

import (
	"math"

	"github.com/mxzinke/colorful-terrarium/terrain"
)

// smoothCoastlines applies intelligent smoothing to coastline areas
func smoothCoastlines(elevation float64, x, y int, elevMap *terrain.ElevationMap, zoom uint32) float64 {
	// Skip smoothing for low zoom levels
	if zoom < 7 {
		return elevation
	}

	// Only smooth on coastlines
	if math.Abs(elevation) > 200 {
		return elevation
	}

	// Calculate pattern size based on zoom level
	patternSize := int(math.Pow(2, float64(zoom-7)))

	// Quick check for coastline using neighborhood stats
	landCount, waterCount, hasEdge := elevMap.GetNeighborhoodStats(x, y, patternSize)

	// Skip smoothing at tile edges to prevent artifacts
	if hasEdge {
		return elevation
	}

	totalChecked := landCount + waterCount
	minCount := int(float64(totalChecked) * 0.2) // At least 20% should be either land or water

	// If not a coastline area, return original elevation
	if landCount < minCount || waterCount < minCount {
		return elevation
	}

	// Initialize weighted sum for smoothing
	var weightedSum float64
	var totalWeight float64
	centerIsLand := elevMap.IsLand(elevation)

	// Sample radius increases with zoom level
	radius := patternSize

	// Collect and weight neighboring elevations
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}

			neighborElev := elevMap.GetElevation(x+dx, y+dy)
			neighborIsLand := elevMap.IsLand(neighborElev)

			// Calculate base weight based on distance
			distance := math.Sqrt(float64(dx*dx + dy*dy))
			weight := 1.0 / (1.0 + distance)

			// Adjust weight based on land/water relationship
			if neighborIsLand == centerIsLand {
				weight *= 2.0
			} else {
				weight *= 1.5
			}

			weightedSum += neighborElev * weight
			totalWeight += weight
		}
	}

	// Calculate smoothed elevation
	smoothedElevation := weightedSum / totalWeight

	// Mix original and smoothed elevations
	mixFactor := 0.5

	// Calculate the final elevation with bounds
	if centerIsLand {
		result := elevation*(1-mixFactor) + smoothedElevation*mixFactor
		return math.Max(0.1, result) // Ensure we stay above sea level
	} else {
		result := elevation*(1-mixFactor) + smoothedElevation*mixFactor
		return math.Min(0, result) // Ensure we stay at or below sea level
	}
}