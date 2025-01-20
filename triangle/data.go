package triangle

import "github.com/paulmach/orb"

type Triangle struct {
	id     string
	points [3]orb.Point
	bounds orb.Bound
}

func NewTriangle(id string, points [3]orb.Point) Triangle {
	// Calculate triangle bounds
	minX := min3(points[0][0], points[1][0], points[2][0])
	minY := min3(points[0][1], points[1][1], points[2][1])
	maxX := max3(points[0][0], points[1][0], points[2][0])
	maxY := max3(points[0][1], points[1][1], points[2][1])

	return Triangle{id: id, points: points, bounds: orb.Bound{
		Min: orb.Point{minX, minY},
		Max: orb.Point{maxX, maxY},
	}}
}

func (t Triangle) ID() string {
	return t.id
}

func (t Triangle) Data() []orb.Ring {
	return []orb.Ring{{
		{t.points[0][0], t.points[0][1]},
		{t.points[1][0], t.points[1][1]},
		{t.points[2][0], t.points[2][1]},
		{t.points[0][0], t.points[0][1]},
	}}
}

func (t Triangle) Points() [3]orb.Point {
	return t.points
}

func (t Triangle) Bound() orb.Bound {
	return t.bounds
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
