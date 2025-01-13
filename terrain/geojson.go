package terrain

import (
	"log"
	"os"

	"github.com/mxzinke/colorful-terrarium/polygon"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

type GeoCoverage struct {
	ice polygon.SpatialIndexer
}

func LoadGeoCoverage() (*GeoCoverage, error) {
	// Read the GeoJSON file
	data, err := os.ReadFile("./glaciers.geojson")
	if err != nil {
		return nil, err
	}

	// Parse the GeoJSON
	iceFeatures, err := geojson.UnmarshalFeatureCollection(data)
	if err != nil {
		return nil, err
	}

	ice := polygon.New()
	log.Printf("Loading %d ice features", len(iceFeatures.Features))
	for _, feature := range iceFeatures.Features {
		if poly, ok := feature.Geometry.(orb.Polygon); ok {
			ice.Insert(poly)
		}

		if multiPoly, ok := feature.Geometry.(orb.MultiPolygon); ok {
			for _, poly := range multiPoly {
				ice.Insert(poly)
			}
		}
	}

	return &GeoCoverage{
		ice: ice,
	}, nil
}

func (gc *GeoCoverage) HasIceInBounds(bounds orb.Bound) bool {
	return gc.ice.IsPointInPolygons(bounds.Min) || gc.ice.IsPointInPolygons(bounds.Max)
}

func (gc *GeoCoverage) IsPointInIce(lon, lat float64) bool {
	return gc.ice.IsPointInPolygons(orb.Point{lon, lat})
}
