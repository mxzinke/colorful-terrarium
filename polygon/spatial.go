package polygon

import (
	"errors"
	"log"
	"math"

	"github.com/dhconnelly/rtreego"
	"github.com/mxzinke/colorful-terrarium/triangle"
	"github.com/paulmach/orb"
)

type Polygon interface {
	ID() string
	Bound() orb.Bound
	Data() []orb.Ring
}

// SpatialIndexer is the interface that wraps the basic spatial index methods
type SpatialIndexer interface {
	// Insert adds a polygon to the spatial index
	Insert(p Polygon) error

	// PointInPolygons checks if a point lies within any indexed polygons, returns list of polygons
	PointInPolygons(p orb.Point) []*Polygon

	// PointInAnyPolygon checks if a point lies within any indexed polygons, returns true if it does
	PointInAnyPolygon(p orb.Point) bool

	// BoundsInAnyPolygon checks if a given bounds lies within any indexed polygons, returns true if it does
	BoundsInAnyPolygon(b orb.Bound) bool

	// PolygonByID returns a polygon by its ID
	PolygonByID(id string) *Polygon
}

// triangleWrapper wraps a triangle to implement rtreego.Spatial
type triangleWrapper struct {
	points   [3]orb.Point
	original *Polygon
	bbox     rtreego.Rect
}

// Bounds implements rtreego.Spatial interface
func (tw *triangleWrapper) Bounds() rtreego.Rect {
	return tw.bbox
}

// Index implements the SpatialIndexer interface
type Index struct {
	rtree  *rtreego.Rtree
	polys  map[string]*Polygon
	bounds orb.Bound
}

// New creates a new spatial index
func New() *Index {
	return &Index{
		rtree:  rtreego.NewTree(2, 25, 50),
		polys:  make(map[string]*Polygon),
		bounds: orb.Bound{Min: orb.Point{1e10, 1e10}, Max: orb.Point{-1e10, -1e10}},
	}
}

func (idx *Index) Size() int {
	return idx.rtree.Size()
}

func CreateIndexFromTriangles(triangles []triangle.Triangle) (*Index, error) {
	if len(triangles) == 0 {
		return nil, errors.New("no triangles provided")
	}

	lookupMap := make(map[string]*Polygon, len(triangles))
	spacials := make([]rtreego.Spatial, len(triangles))
	bounds := triangles[0].Bound()

	for i, tri := range triangles {
		rectBounds := tri.Bound()
		bounds = bounds.Union(rectBounds)

		rect, err := rtreego.NewRectFromPoints(
			rtreego.Point{rectBounds.Min[0] - 1e-7, rectBounds.Min[1] - 1e-7},
			rtreego.Point{rectBounds.Max[0] + 1e-7, rectBounds.Max[1] + 1e-7},
		)
		if err != nil {
			return nil, err
		}

		poly := Polygon(tri)
		polyPointer := &poly
		// Create and insert triangle wrapper
		tw := &triangleWrapper{
			points:   tri.Points(),
			original: polyPointer,
			bbox:     rect,
		}

		lookupMap[poly.ID()] = polyPointer
		spacials[i] = tw
	}

	rtree := rtreego.NewTree(2, 25, 50, spacials...)

	return &Index{
		rtree:  rtree,
		polys:  lookupMap,
		bounds: bounds,
	}, nil
}

// Insert implements SpatialIndexer
func (idx *Index) Insert(p Polygon) error {
	poly := orb.Polygon(p.Data())

	triangles, err := triangle.FromPolygon(poly)
	if err != nil {
		return err
	}

	// Update bounds
	idx.bounds = idx.bounds.Union(poly.Bound())

	// Create a pointer to the original polygon
	polyPointer := &p

	// Create triangles from indices
	for i := 0; i < len(triangles); i++ {
		// Create R-tree rectangle
		rect, err := rtreego.NewRectFromPoints(
			rtreego.Point{triangles[i].Bound().Min[0] - 1e-7, triangles[i].Bound().Min[1] - 1e-7},
			rtreego.Point{triangles[i].Bound().Max[0] + 1e-7, triangles[i].Bound().Max[1] + 1e-7},
		)
		if err != nil {
			return err
		}

		// Create and insert triangle wrapper
		tw := &triangleWrapper{
			points:   triangles[i].Points(),
			original: polyPointer,
			bbox:     rect,
		}
		idx.rtree.Insert(tw)
	}
	idx.polys[p.ID()] = polyPointer

	return nil
}

// PointInPolygons implements SpatialIndexer
func (idx *Index) PointInPolygons(p orb.Point) []*Polygon {
	results := idx.getIntersectingTriangles(p)

	// Check each triangle
	polys := make([]*Polygon, 0, len(results))
	for _, item := range results {
		tw := item.(*triangleWrapper)

		found := false
		for _, poly := range polys {
			if (*poly).ID() == (*tw.original).ID() {
				found = true
				break
			}
		}

		if found {
			continue
		}

		if PointInTriangle(p, tw.points[0], tw.points[1], tw.points[2]) {
			polys = append(polys, tw.original)
		}
	}

	return polys
}

// PointInAnyPolygon implements SpatialIndexer
func (idx *Index) PointInAnyPolygon(p orb.Point) bool {
	for _, tri := range idx.getIntersectingTriangles(p) {
		tw := tri.(*triangleWrapper)
		if PointInTriangle(p, tw.points[0], tw.points[1], tw.points[2]) {
			return true
		}
	}
	return false
}

func (idx *Index) BoundsInAnyPolygon(b orb.Bound) bool {
	rect, err := rtreego.NewRectFromPoints(
		rtreego.Point{b.Min[0], b.Min[1]},
		rtreego.Point{b.Max[0], b.Max[1]},
	)
	if err != nil {
		log.Printf("error creating rect for bounds in any polygon search: %v", err)
		return false
	}

	results := idx.rtree.SearchIntersectWithLimit(1, rect)
	return len(results) > 0
}

func (idx *Index) PolygonByID(id string) *Polygon {
	return idx.polys[id]
}

func (idx *Index) getIntersectingTriangles(p orb.Point) []rtreego.Spatial {
	// Quick check if point is in index bounds
	if !idx.bounds.Contains(p) {
		return []rtreego.Spatial{}
	}

	// Create a point rectangle for R-tree search
	pointRect, err := rtreego.NewRect(
		rtreego.Point{p[0], p[1]},
		[]float64{0.01, 0.01},
	)
	if err != nil {
		log.Printf("error creating point rect for point in polygon search: %v", err)
		return []rtreego.Spatial{}
	}

	// Search R-tree for potential triangles
	return idx.rtree.SearchIntersect(pointRect)
}

func PointInTriangle(pt, v1, v2, v3 [2]float64) bool {
	d1, d2, d3 := sign(pt, v1, v2), sign(pt, v2, v3), sign(pt, v3, v1)

	hasNeg := d1 < 0 || d2 < 0 || d3 < 0
	hasPos := d1 > 0 || d2 > 0 || d3 > 0

	return !(hasNeg && hasPos)
}

// Helper functions for min/max of 3 values
func min3(a, b, c float64) float64 {
	return min2(min2(a, b), c)
}

func max3(a, b, c float64) float64 {
	return max2(max2(a, b), c)
}

func min2(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max2(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func sign(p1, p2, p3 [2]float64) float64 {
	return (p1[0]-p3[0])*(p2[1]-p3[1]) - (p2[0]-p3[0])*(p1[1]-p3[1])
}

func PointDistanceToTriangle(p [2]float64, t [3][2]float64) float64 {
	d1, d2, d3 := distanceToPoint(p, t[0]), distanceToPoint(p, t[1]), distanceToPoint(p, t[2])

	if d1 <= d2 && d2 <= d3 {
		return distanceToLine(p, t[0], t[1])
	} else if d1 <= d3 && d3 <= d2 {
		return distanceToLine(p, t[0], t[2])
	} else if d2 <= d1 && d1 <= d3 {
		return distanceToLine(p, t[1], t[0])
	} else if d2 <= d3 && d3 <= d1 {
		return distanceToLine(p, t[1], t[2])
	} else if d3 <= d1 && d1 <= d2 {
		return distanceToLine(p, t[2], t[0])
	} else {
		return distanceToLine(p, t[2], t[1])
	}
}

// distanceToLine calculates the distance from point p to line segment ab
func distanceToLine(p, a, b [2]float64) float64 {
	// Vector from a to b
	ab := [2]float64{
		b[0] - a[0],
		b[1] - a[1],
	}

	// Vector from a to p
	ap := [2]float64{
		p[0] - a[0],
		p[1] - a[1],
	}

	// Length of line segment squared
	lenSq := ab[0]*ab[0] + ab[1]*ab[1]

	// Handle degenerate case of zero-length line
	if lenSq == 0 {
		return math.Sqrt(ap[0]*ap[0] + ap[1]*ap[1])
	}

	// Calculate projection of ap onto ab
	// t is the normalized distance along the line
	t := (ap[0]*ab[0] + ap[1]*ab[1]) / lenSq

	// Clamp t to [0,1] to handle points beyond segment ends
	t = math.Max(0, math.Min(1, t))

	// Calculate closest point on line segment
	closest := [2]float64{
		a[0] + t*ab[0],
		a[1] + t*ab[1],
	}

	// Return distance to closest point
	dx := p[0] - closest[0]
	dy := p[1] - closest[1]
	return math.Sqrt(dx*dx + dy*dy)
}

func distanceToPoint(a, b [2]float64) float64 {
	dx := a[0] - b[0]
	dy := a[1] - b[1]
	return math.Sqrt(dx*dx + dy*dy)
}
