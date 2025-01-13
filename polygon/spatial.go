package polygon

import (
	"github.com/dhconnelly/rtreego"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/planar"
	"github.com/rclancey/go-earcut"
)

// SpatialIndexer is the interface that wraps the basic spatial index methods
type SpatialIndexer interface {
	// Insert adds a polygon to the spatial index
	Insert(p orb.Polygon) error

	// IsPointInPolygons checks if a point lies within any indexed polygon
	IsPointInPolygons(p orb.Point) bool
}

// triangleWrapper wraps a triangle to implement rtreego.Spatial
type triangleWrapper struct {
	points    [3]orb.Point
	polyIndex int // Index of the original polygon
	bbox      rtreego.Rect
}

// Bounds implements rtreego.Spatial interface
func (tw *triangleWrapper) Bounds() rtreego.Rect {
	return tw.bbox
}

// Index implements the SpatialIndexer interface
type Index struct {
	rtree    *rtreego.Rtree
	polygons []orb.Polygon // Store original polygons for reference
	bounds   orb.Bound
}

// New creates a new spatial index
func New() SpatialIndexer {
	return &Index{
		rtree:    rtreego.NewTree(2, 25, 50),
		polygons: make([]orb.Polygon, 0),
		bounds:   orb.Bound{Min: orb.Point{1e10, 1e10}, Max: orb.Point{-1e10, -1e10}},
	}
}

// convertPolygonToEarcut converts an orb.Polygon to earcut input format
func convertPolygonToEarcut(poly orb.Polygon) ([]float64, []int) {
	// Calculate total number of points
	totalPoints := len(poly[0]) // exterior ring
	holes := make([]int, len(poly)-1)
	currentIndex := totalPoints

	// Calculate hole start indices and total points
	for i := 1; i < len(poly); i++ {
		holes[i-1] = currentIndex
		totalPoints += len(poly[i])
		currentIndex += len(poly[i])
	}

	// Create flat array of vertices
	vertices := make([]float64, 0, totalPoints*2)

	// Add exterior ring
	for _, p := range poly[0] {
		vertices = append(vertices, p[0], p[1])
	}

	// Add holes
	for i := 1; i < len(poly); i++ {
		for _, p := range poly[i] {
			vertices = append(vertices, p[0], p[1])
		}
	}

	return vertices, holes
}

// Insert implements SpatialIndexer
func (idx *Index) Insert(poly orb.Polygon) error {
	// Convert polygon to earcut format
	vertices, holes := convertPolygonToEarcut(poly)

	// Triangulate using earcut
	indices, err := earcut.Earcut(vertices, holes, 2)
	if err != nil {
		return err
	}

	// Store original polygon
	polyIndex := len(idx.polygons)
	idx.polygons = append(idx.polygons, poly)

	// Update bounds
	idx.bounds = idx.bounds.Union(poly.Bound())

	// Create triangles from indices
	for i := 0; i < len(indices); i += 3 {
		// Get vertex indices for this triangle
		i1, i2, i3 := indices[i]*2, indices[i+1]*2, indices[i+2]*2

		// Create triangle points
		points := [3]orb.Point{
			{vertices[i1], vertices[i1+1]},
			{vertices[i2], vertices[i2+1]},
			{vertices[i3], vertices[i3+1]},
		}

		// Calculate triangle bounds
		minX := min3(points[0][0], points[1][0], points[2][0])
		minY := min3(points[0][1], points[1][1], points[2][1])
		maxX := max3(points[0][0], points[1][0], points[2][0])
		maxY := max3(points[0][1], points[1][1], points[2][1])

		// Create R-tree rectangle
		rect, err := rtreego.NewRectFromPoints(
			rtreego.Point{minX - 1e-7, minY - 1e-7},
			rtreego.Point{maxX + 1e-7, maxY + 1e-7},
		)
		if err != nil {
			return err
		}

		// Create and insert triangle wrapper
		tw := &triangleWrapper{
			points:    points,
			polyIndex: polyIndex,
			bbox:      rect,
		}
		idx.rtree.Insert(tw)
	}

	return nil
}

// IsPointInPolygons implements SpatialIndexer
func (idx *Index) IsPointInPolygons(p orb.Point) bool {
	// Quick check if point is in index bounds
	if !idx.bounds.Contains(p) {
		return false
	}

	// Create a point rectangle for R-tree search
	pointRect, err := rtreego.NewRect(
		rtreego.Point{p[0], p[1]},
		[]float64{0.1, 0.1},
	)
	if err != nil {
		return false
	}

	// Search R-tree for potential triangles
	results := idx.rtree.SearchIntersect(pointRect)

	// Check each triangle
	for _, item := range results {
		tw := item.(*triangleWrapper)

		// Convert triangle to orb.Polygon for robust point-in-polygon test
		trianglePoly := orb.Polygon{{
			tw.points[0],
			tw.points[1],
			tw.points[2],
			tw.points[0], // Close the ring
		}}

		if planar.PolygonContains(trianglePoly, p) {
			return true
		}
	}

	return false
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
