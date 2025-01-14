package main

import (
	"math"

	"github.com/mxzinke/colorful-terrarium/terrain"
)

func buildDesertFactorMatrix(geoCoverage *terrain.GeoCoverage, tileBounds *TileBounds, tileMap *terrain.ElevationMap, zoom uint32) [][]float64 {
	desertFactorMatrix := make([][]float64, tileMap.TileSize)
	for y := range desertFactorMatrix {
		latitude := tileBounds.GetPixelLat(y)

		desertFactorMatrix[y] = make([]float64, tileMap.TileSize)

		// Skip latitudes above 60 degrees (non desert zone)
		if math.Abs(latitude) > 60 {
			continue
		}

		for x := range desertFactorMatrix[y] {
			if !tileMap.IsLand(x, y) {
				desertFactorMatrix[y][x] = 0
				continue
			}

			longitude := tileBounds.GetPixelLng(x)
			desertFactorMatrix[y][x] = geoCoverage.DesertFactorForPoint(longitude, latitude)
		}
	}

	return desertFactorMatrix
}
