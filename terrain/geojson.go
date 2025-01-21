package terrain

import (
	"fmt"
	"log"
	"os"

	"github.com/mxzinke/colorful-terrarium/polygon"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

func loadIndexerFromGeojson(path string) (polygon.SpatialIndexer, error) {
	// Read the GeoJSON file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse the GeoJSON
	fc, err := geojson.UnmarshalFeatureCollection(data)
	if err != nil {
		return nil, err
	}

	polys := polygon.New()
	log.Printf("Loading polygon index with %d features", len(fc.Features))
	for featureIdx, feature := range fc.Features {
		if poly, ok := feature.Geometry.(orb.Polygon); ok {
			id := fmt.Sprint(feature.Properties["id"])
			if feature.Properties["id"] == nil {
				id = fmt.Sprintf("%d", featureIdx)
			}

			polys.Insert(internalPolygon{poly, id})
		}

		if multiPoly, ok := feature.Geometry.(orb.MultiPolygon); ok {
			baseId := fmt.Sprint(feature.Properties["id"])
			if feature.Properties["id"] == nil {
				baseId = fmt.Sprintf("%d", featureIdx)
			}

			for multiPolyIdx, poly := range multiPoly {
				id := fmt.Sprintf("%s-m%d", baseId, multiPolyIdx)
				polys.Insert(internalPolygon{poly, id})
			}
		}
	}

	log.Printf("Loaded %d triangles (from %d features) into index", polys.Size(), len(fc.Features))

	return polys, nil
}
