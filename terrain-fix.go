package main

import (
	"github.com/mxzinke/colorful-terrarium/terrain"
)

const fixedElevation = -220
const minHeight = -24

func fixElevationMap(elevationMap *terrain.ElevationMap, tileBounds *TileBounds, geoCoverage *terrain.GeoCoverage) {
	if tileBounds.Zoom > 10 {
		return
	}

	hasFixFactors := geoCoverage.HasBoundsAnyFixFactors(tileBounds.Bound())
	if !hasFixFactors {
		return
	}

	for y, row := range elevationMap.Data {
		lat := tileBounds.GetPixelLat(y)
		for x, cell := range row {
			lon := tileBounds.GetPixelLng(x)

			// Skip if point is in land
			if cell > 20 || geoCoverage.IsPointInLand(lon, lat) {
				continue
			}

			// Skip if factor is 0
			factor := geoCoverage.HighFixFactorForPoint(lon, lat)
			if factor == 0 {
				continue
			}

			targetElevation := minHeight + (fixedElevation-minHeight)*float32(factor)

			// When factor is small than 0.05 we know the edge is comming and we should adjust to the outer edge elevation
			if factor <= 0.1 {
				untilEdgeFactor := float32(factor / 0.1)
				edgeElevation := targetElevation*untilEdgeFactor + cell*(1-untilEdgeFactor)
				elevationMap.ModifyElevation(x, y, edgeElevation)
			} else {
				elevationMap.ModifyElevation(x, y, targetElevation)
			}
		}
	}
}
