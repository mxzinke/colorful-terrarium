package main

import (
	"math"

	"github.com/mxzinke/colorful-terrarium/terrain"
)

const (
	polarStartLatitude    = 64
	polarAbsoluteLatitude = 70
	polLatitude           = 85.05
	warmZoneLatitude      = 28
	earthTilt             = 6
)

type PixelCell struct {
	elevation   float32
	latitude    float64
	longitude   float64
	geoCoverage *terrain.GeoCoverage
}

func (c *PixelCell) Elevation() float32 {
	return c.elevation
}

func (c *PixelCell) IsLand() bool {
	if c.elevation < -420 {
		return false
	}
	return c.elevation > 100 || c.geoCoverage.IsPointInLand(c.longitude, c.latitude)
}

func (c *PixelCell) Latitude() float64 {
	return c.latitude
}

func (c *PixelCell) Longitude() float64 {
	return c.longitude
}

func (c *PixelCell) IsIce() bool {
	if c.Latitude() < 23 && c.Latitude() > -35 {
		return false
	}
	return c.geoCoverage.IsPointInIce(c.longitude, c.latitude)
}

func (c *PixelCell) DesertFactor() float64 {
	return c.geoCoverage.DesertFactorForPoint(c.longitude, c.latitude)
}

func (c *PixelCell) PolarFactor() float64 {
	polarFactor := 0.0
	latitude := math.Abs(c.Latitude())
	if c.Latitude() < -1*(polarStartLatitude-earthTilt) {
		polarFactor = math.Min(math.Max((latitude-(polarStartLatitude-earthTilt))/((polarAbsoluteLatitude-earthTilt)-(polarStartLatitude-earthTilt)), 0), 1)
	} else if c.Latitude() > (polarStartLatitude + earthTilt) {
		polarFactor = math.Min(math.Max((latitude-(polarStartLatitude+earthTilt))/((polarAbsoluteLatitude+earthTilt)-(polarStartLatitude+earthTilt)), 0), 1)
	}
	return polarFactor
}

func (c *PixelCell) AquatorFactor() float64 {
	return math.Max(0, math.Min(1, (math.Abs(c.Latitude())/polarAbsoluteLatitude)))
}

func GetCellsForTile(elevationMap *terrain.ElevationMap, tile *TileBounds, geoCoverage *terrain.GeoCoverage) ([][]*PixelCell, error) {
	cells := make([][]*PixelCell, elevationMap.TileSize)
	for y := 0; y < elevationMap.TileSize; y++ {
		cells[y] = make([]*PixelCell, elevationMap.TileSize)
		for x := 0; x < elevationMap.TileSize; x++ {
			cells[y][x] = &PixelCell{
				elevation:   elevationMap.GetElevation(x, y),
				latitude:    tile.GetPixelLat(y),
				longitude:   tile.GetPixelLng(x),
				geoCoverage: geoCoverage,
			}
		}
	}
	return cells, nil
}
