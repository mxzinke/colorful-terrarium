package polygon

import "math"

func DistanceToPolygon(point [2]float64, poly Polygon) float64 {
	rings := poly.Data()[0]
	if len(rings) < 3 {
		return math.Inf(1)
	}

	minDist := math.Inf(1)

	// Iteriere über alle Kanten des Polygons
	for i := 0; i < len(rings); i++ {
		j := (i + 1) % len(rings)

		dist := distanceToLineSegment(point, rings[i], rings[j])
		if dist < minDist {
			minDist = dist
		}
	}

	return minDist
}

// distanceToLineSegment berechnet den Abstand zwischen einem Punkt und einer Linie
func distanceToLineSegment(p, start, end [2]float64) float64 {
	// Vektor von start zu end
	lineVec := [2]float64{end[0] - start[0], end[1] - start[1]}

	// Vektor von start zu p
	pointVec := [2]float64{p[0] - start[0], p[1] - start[1]}

	// Länge des Liniensegments
	lineLen := lineVec[0]*lineVec[0] + lineVec[1]*lineVec[1]

	// Spezialfall: start == end
	if lineLen == 0 {
		return math.Sqrt(pointVec[0]*pointVec[0] + pointVec[1]*pointVec[1])
	}

	// Projektion von pointVec auf lineVec
	t := (pointVec[0]*lineVec[0] + pointVec[1]*lineVec[1]) / lineLen

	if t < 0 {
		// Punkt ist näher am Startpunkt
		return math.Sqrt(pointVec[0]*pointVec[0] + pointVec[1]*pointVec[1])
	}

	if t > 1 {
		// Punkt ist näher am Endpunkt
		return math.Sqrt(
			(p[0]-end[0])*(p[0]-end[0]) +
				(p[1]-end[1])*(p[1]-end[1]),
		)
	}

	// Nächster Punkt liegt auf der Linie
	projection := [2]float64{
		start[0] + t*lineVec[0],
		start[1] + t*lineVec[1],
	}

	return math.Sqrt(
		(p[0]-projection[0])*(p[0]-projection[0]) +
			(p[1]-projection[1])*(p[1]-projection[1]),
	)
}
