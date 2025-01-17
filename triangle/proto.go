package triangle

import (
	"fmt"

	internal "github.com/mxzinke/colorful-terrarium/triangle/proto"
	"github.com/paulmach/orb"
	"google.golang.org/protobuf/proto"
)

func Unmarshal(data []byte) ([]Triangle, error) {
	tri := internal.TriangleCollection{}
	err := proto.Unmarshal(data, &tri)
	if err != nil {
		return nil, err
	}

	triangles := make([]Triangle, len(tri.Triangles))
	for i, t := range tri.Triangles {
		triangles[i] = NewTriangle(
			fmt.Sprintf("%d", i),
			[3]orb.Point{
				{float64(t.P1.X), float64(t.P1.Y)},
				{float64(t.P2.X), float64(t.P2.Y)},
				{float64(t.P3.X), float64(t.P3.Y)},
			},
		)
	}

	return triangles, nil
}

func Marshal(tri []Triangle) ([]byte, error) {
	msg := make([]*internal.Triangle, len(tri))
	for i, t := range tri {
		msg[i] = &internal.Triangle{
			P1: &internal.Point{X: float32(t.points[0][0]), Y: float32(t.points[0][1])},
			P2: &internal.Point{X: float32(t.points[1][0]), Y: float32(t.points[1][1])},
			P3: &internal.Point{X: float32(t.points[2][0]), Y: float32(t.points[2][1])},
		}
	}
	return proto.Marshal(&internal.TriangleCollection{Triangles: msg})
}
