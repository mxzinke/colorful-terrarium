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

	// Create a new polygon indexer
	indexer := polygon.New()

	log.Printf("Loading polygon index with %d triangles", len(triangles))

	// Insert the triangles into the indexer
	for _, triangle := range triangles {
		err := indexer.InsertTriangle(triangle.ID(), triangle)
		if err != nil {
			log.Fatal(err)
		}
	}

	return indexer, nil
}
