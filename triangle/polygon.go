package triangle

import (
	"strconv"

	"github.com/paulmach/orb"
	"github.com/rclancey/go-earcut"
)

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

func FromPolygon(p orb.Polygon) ([]Triangle, error) {
	vertices, holes := convertPolygonToEarcut(p)

	indices, err := earcut.Earcut(vertices, holes, 2)
	if err != nil {
		return nil, err
	}

	triangles := make([]Triangle, len(indices)/3)
	for i := 0; i < len(indices); i += 3 {
		// Get vertex indices for this triangle
		i1, i2, i3 := indices[i]*2, indices[i+1]*2, indices[i+2]*2

		// Create triangle points
		points := [3]orb.Point{
			{vertices[i1], vertices[i1+1]},
			{vertices[i2], vertices[i2+1]},
			{vertices[i3], vertices[i3+1]},
		}

		triangles[i/3] = NewTriangle(strconv.Itoa(i), points)
	}

	return triangles, nil
}
