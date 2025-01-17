package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mxzinke/colorful-terrarium/triangle"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: geojson-to-tri <input.geojson> <output.tri.pbf>")
		os.Exit(1)
	}

	inputPath := os.Args[1]
	outputPath := os.Args[2]

	// GeoJSON-Datei einlesen
	data, err := os.ReadFile(inputPath)
	if err != nil {
		log.Fatal(err)
	}

	// GeoJSON parsen
	fc, err := geojson.UnmarshalFeatureCollection(data)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Loaded %d features", len(fc.Features))

	triangles := make([]triangle.Triangle, 0)
	for _, feature := range fc.Features {
		if poly, ok := feature.Geometry.(orb.Polygon); ok {
			newTriangles, err := triangle.FromPolygon(poly)
			if err != nil {
				log.Fatal(err)
			}
			triangles = append(triangles, newTriangles...)
		}

		if multiPoly, ok := feature.Geometry.(orb.MultiPolygon); ok {
			for _, poly := range multiPoly {
				newTriangles, err := triangle.FromPolygon(poly)
				if err != nil {
					log.Fatal(err)
				}
				triangles = append(triangles, newTriangles...)
			}
		}
	}
	log.Printf("Collected %d triangles", len(triangles))

	trianglesBytes, err := triangle.Marshal(triangles)
	if err != nil {
		log.Fatal(err)
	}

	os.WriteFile(outputPath, trianglesBytes, 0644)
}
