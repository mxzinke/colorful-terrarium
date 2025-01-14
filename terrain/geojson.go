package terrain

import (
	"fmt"
	"log"
	"math"
	"os"
	"sync"

	"github.com/mxzinke/colorful-terrarium/polygon"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

const maxDesertInnerOuterDistance = 0.56 // approx 60km

type GeoCoverage struct {
	ice          polygon.SpatialIndexer
	lakes        polygon.SpatialIndexer
	innerDeserts polygon.SpatialIndexer
	outerDeserts polygon.SpatialIndexer
}

type internalPolygon struct {
	orb.Polygon
	id string
}

func (p internalPolygon) ID() string {
	return p.id
}

func (p internalPolygon) Data() []orb.Ring {
	return p.Polygon
}

func (p internalPolygon) Clone() polygon.Polygon {
	return internalPolygon{p.Polygon.Clone(), p.id}
}

func (p internalPolygon) Equal(polygon polygon.Polygon) bool {
	return p.Polygon.Equal(polygon.Data())
}

func LoadGeoCoverage() (*GeoCoverage, error) {
	var wg sync.WaitGroup

	var ice polygon.SpatialIndexer
	var lakes polygon.SpatialIndexer
	var innerDeserts polygon.SpatialIndexer
	var outerDeserts polygon.SpatialIndexer

	wg.Add(4)
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

	go func() {
		val, err := loadIndexer("./inner-deserts.geojson")
		if err != nil {
			log.Fatal(err)
		}
		innerDeserts = val
		wg.Done()
	}()

	go func() {
		val, err := loadIndexer("./outer-deserts.geojson")
		if err != nil {
			log.Fatal(err)
		}
		outerDeserts = val
		wg.Done()
	}()

	wg.Wait()

	return &GeoCoverage{
		ice:          ice,
		lakes:        lakes,
		innerDeserts: innerDeserts,
		outerDeserts: outerDeserts,
	}, nil
}

func loadIndexer(path string) (polygon.SpatialIndexer, error) {
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
			polys.Insert(internalPolygon{poly, fmt.Sprintf("%d", featureIdx)})
		}

		if multiPoly, ok := feature.Geometry.(orb.MultiPolygon); ok {
			for multiPolyIdx, poly := range multiPoly {
				polys.Insert(internalPolygon{poly, fmt.Sprintf("%d-%d", featureIdx, multiPolyIdx)})
			}
		}
	}

	return polys, nil
}

func (gc *GeoCoverage) IsPointInIce(lon, lat float64) bool {
	return gc.ice.PointInAnyPolygon(orb.Point{lon, lat})
}

func (gc *GeoCoverage) IsPointInLakes(lon, lat float64) bool {
	return gc.lakes.PointInAnyPolygon(orb.Point{lon, lat})
}

func (gc *GeoCoverage) DesertFactorForPoint(lon, lat float64) float64 {
	outerDesertPolys := gc.outerDeserts.PointInPolygons(orb.Point{lon, lat})
	// When not somewhere in the desert zone
	if len(outerDesertPolys) == 0 {
		return 0.0
	}

	innerDesertPolys := gc.innerDeserts.PointInPolygons(orb.Point{lon, lat})
	if len(innerDesertPolys) > 0 {
		return 1.0
	}

	distancePolygons := make([]*polygon.Polygon, len(outerDesertPolys))
	for i, poly := range outerDesertPolys {
		distancePolygons[i] = gc.innerDeserts.PolygonByID((*poly).ID())
	}

	if len(distancePolygons) == 0 {
		log.Fatalf("no distance polygons found for point %f, %f", lon, lat)
	}

	// Important, to use the same polygon for both distance calculations
	idWithDistanceToInner := (*distancePolygons[0]).ID()

	// Find the closest inner polygon
	distanceToInner := polygon.DistanceToPolygon(orb.Point{lon, lat}, *distancePolygons[0])
	if len(distancePolygons) > 1 {
		for i, poly := range distancePolygons {
			if i == 0 {
				continue
			}

			distance := polygon.DistanceToPolygon(orb.Point{lon, lat}, *poly)
			if distance < distanceToInner {
				distanceToInner = distance
				idWithDistanceToInner = (*poly).ID()
			}
		}
	}

	distanceToOuter := polygon.DistanceToPolygon(orb.Point{lon, lat}, *gc.outerDeserts.PolygonByID(idWithDistanceToInner))

	return 1 - math.Max(0.0, math.Min(distanceToInner/(distanceToInner+distanceToOuter), 1.0))
}
