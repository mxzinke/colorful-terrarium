package main

import (
	"compress/gzip"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/mxzinke/colorful-terrarium/terrain"
)

const (
	addr = ":8080"
)

func main() {
	geoCoverage, err := terrain.LoadGeoCoverage()
	if err != nil {
		log.Fatalf("Failed to load geo coverage: %v", err)
	}

	log.Printf("Starting terrain tile server on %s", addr)
	log.Printf("Tiles Server Format: http://127.0.0.1%s/{theme}/{z}/{y}/{x}.{fileType}", addr)

	if err := http.ListenAndServe(
		addr,
		handlers.CompressHandlerLevel(
			handlers.CORS(
				handlers.AllowedOrigins([]string{"*"}),
				handlers.AllowedMethods([]string{"GET"}),
			)(MainHandler(geoCoverage)),
			gzip.BestCompression),
	); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
