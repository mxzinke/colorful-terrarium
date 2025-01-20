package terrain

import (
	"log"
	"os"

	"github.com/mxzinke/colorful-terrarium/polygon"
	"github.com/mxzinke/colorful-terrarium/triangle"
)

func loadIndexerFromTrianglePkg(path string) (polygon.SpatialIndexer, error) {
	// Read the Triangle file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Unmarshal the Triangle file
	triangles, err := triangle.Unmarshal(data)
	if err != nil {
		return nil, err
	}

	log.Printf("Loading polygon index with %d triangles", len(triangles))

	// Insert the triangles into the indexer
	indexer, err := polygon.CreateIndexFromTriangles(triangles)
	if err != nil {
		return nil, err
	}

	return indexer, nil
}
