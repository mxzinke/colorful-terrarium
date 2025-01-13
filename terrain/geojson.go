package terrain

import (
	"log"
	"os"
	"sync"

	"github.com/mxzinke/colorful-terrarium/polygon"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

type GeoCoverage struct {
	ice   polygon.SpatialIndexer
	lakes polygon.SpatialIndexer
}

func LoadGeoCoverage() (*GeoCoverage, error) {
	var wg sync.WaitGroup

	var ice polygon.SpatialIndexer
	var lakes polygon.SpatialIndexer

	wg.Add(2)
	go func() {
		val, err := loadIndexer("./glaciers.geojson")
		if err != nil {
			log.Fatal(err)
		}
		ice = val
		wg.Done()
	}()

	go func() {
		val, err := loadIndexer("./lakes.geojson")
		if err != nil {
			log.Fatal(err)
		}
		lakes = val
		wg.Done()
	}()

	wg.Wait()

	return &GeoCoverage{
		ice:   ice,
		lakes: lakes,
	}, nil
}

func loadIndexer(path string) (polygon.SpatialIndexer, error) {
	// Read the GeoJSON file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse the GeoJSON
	iceFeatures, err := geojson.UnmarshalFeatureCollection(data)
	if err != nil {
		return nil, err
	}

	polys := polygon.New()
	log.Printf("Loading polygon index with %d features", len(iceFeatures.Features))
	for _, feature := range iceFeatures.Features {
		if poly, ok := feature.Geometry.(orb.Polygon); ok {
			polys.Insert(poly)
		}

		if multiPoly, ok := feature.Geometry.(orb.MultiPolygon); ok {
			for _, poly := range multiPoly {
				polys.Insert(poly)
			}
		}
	}

	return polys, nil
}

func (gc *GeoCoverage) IsPointInIce(lon, lat float64) bool {
	return gc.ice.IsPointInPolygons(orb.Point{lon, lat})
}

func (gc *GeoCoverage) IsPointInLakes(lon, lat float64) bool {
	return gc.lakes.IsPointInPolygons(orb.Point{lon, lat})
}
