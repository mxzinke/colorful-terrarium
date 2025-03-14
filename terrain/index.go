package terrain

import (
	"log"
	"math"
	"sync"

	"github.com/mxzinke/colorful-terrarium/polygon"
	"github.com/paulmach/orb"
)

const maxDesertInnerOuterDistance = 0.56 // approx 60km

type GeoCoverage struct {
	ice          polygon.SpatialIndexer
	innerDeserts polygon.SpatialIndexer
	outerDeserts polygon.SpatialIndexer
	land         polygon.SpatialIndexer
	highFixInner polygon.SpatialIndexer
	highFixOuter polygon.SpatialIndexer
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

func LoadGeoCoverage() (*GeoCoverage, error) {
	var wg sync.WaitGroup

	var ice polygon.SpatialIndexer
	var innerDeserts polygon.SpatialIndexer
	var outerDeserts polygon.SpatialIndexer
	var land polygon.SpatialIndexer
	var highFixInner polygon.SpatialIndexer
	var highFixOuter polygon.SpatialIndexer

	wg.Add(6)
	go func() {
		val, err := loadIndexerFromTrianglePkg("./data/glaciers.tri.pbf")
		if err != nil {
			log.Fatal(err)
		}
		ice = val
		wg.Done()
	}()

	go func() {
		val, err := loadIndexerFromTrianglePkg("./data/osm_land_simplified.tri.pbf")
		if err != nil {
			log.Fatal(err)
		}
		land = val
		wg.Done()
	}()

	go func() {
		val, err := loadIndexerFromGeojson("./data/inner-deserts.geojson")
		if err != nil {
			log.Fatal(err)
		}
		innerDeserts = val
		wg.Done()
	}()

	go func() {
		val, err := loadIndexerFromGeojson("./data/outer-deserts.geojson")
		if err != nil {
			log.Fatal(err)
		}
		outerDeserts = val
		wg.Done()
	}()

	go func() {
		val, err := loadIndexerFromGeojson("./data/high-fix-inner.geojson")
		if err != nil {
			log.Fatal(err)
		}
		highFixInner = val
		wg.Done()
	}()

	go func() {
		val, err := loadIndexerFromGeojson("./data/high-fix-outer.geojson")
		if err != nil {
			log.Fatal(err)
		}
		highFixOuter = val
		wg.Done()
	}()

	wg.Wait()

	return &GeoCoverage{
		ice:          ice,
		innerDeserts: innerDeserts,
		outerDeserts: outerDeserts,
		land:         land,
		highFixInner: highFixInner,
		highFixOuter: highFixOuter,
	}, nil
}

func (gc *GeoCoverage) IsPointInLand(lon, lat float64) bool {
	return gc.land.PointInAnyPolygon(orb.Point{lon, lat})
}

func (gc *GeoCoverage) IsPointInIce(lon, lat float64) bool {
	return gc.ice.PointInAnyPolygon(orb.Point{lon, lat})
}

func (gc *GeoCoverage) DesertFactorForPoint(lon, lat float64) float64 {
	return distanceFactorForPoint(orb.Point{lon, lat}, gc.innerDeserts, gc.outerDeserts)
}

func (gc *GeoCoverage) HasBoundsAnyFixFactors(b orb.Bound) bool {
	return gc.highFixOuter.BoundsInAnyPolygon(b)
}

func (gc *GeoCoverage) HighFixFactorForPoint(lon, lat float64) float64 {
	return distanceFactorForPoint(orb.Point{lon, lat}, gc.highFixInner, gc.highFixOuter)
}

func distanceFactorForPoint(point orb.Point, inner, outer polygon.SpatialIndexer) float64 {
	outerPolys := outer.PointInPolygons(point)
	if len(outerPolys) == 0 {
		return 0.0
	}

	innerPolys := inner.PointInPolygons(point)
	if len(innerPolys) > 0 {
		return 1.0
	}

	distancePolygons := make([]*polygon.Polygon, len(outerPolys))
	for i, poly := range outerPolys {
		inner := inner.PolygonByID((*poly).ID())
		if inner == nil {
			log.Printf("WARNING: no inner polygon found for %s", (*poly).ID())
			distancePolygons[i] = poly
		}
		distancePolygons[i] = inner
	}

	if len(distancePolygons) == 0 {
		log.Fatalf("no distance polygons found for point %f, %f", point.Lon(), point.Lat())
	}

	// Important, to use the same polygon for both distance calculations
	idWithDistanceToInner := (*distancePolygons[0]).ID()

	// Find the closest inner polygon
	distanceToInner := polygon.DistanceToPolygon(point, *distancePolygons[0])
	if len(distancePolygons) > 1 {
		for i, poly := range distancePolygons {
			if i == 0 {
				continue
			}

			distance := polygon.DistanceToPolygon(point, *poly)
			if distance < distanceToInner {
				distanceToInner = distance
				idWithDistanceToInner = (*poly).ID()
			}
		}
	}

	outerDistancePoly := outer.PolygonByID(idWithDistanceToInner)
	if outerDistancePoly == nil {
		log.Fatalf("Fatal Error: no same as outer polygon found for %s", idWithDistanceToInner)
		outerDistancePoly = outerPolys[0]
	}
	distanceToOuter := polygon.DistanceToPolygon(point, *outerDistancePoly)

	return 1 - math.Max(0.0, math.Min(distanceToInner/(distanceToInner+distanceToOuter), 1.0))
}
