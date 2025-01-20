package main

import (
	"context"
	"image"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/mxzinke/colorful-terrarium/colors"
	"github.com/mxzinke/colorful-terrarium/colors/color_v1"
	"github.com/mxzinke/colorful-terrarium/terrain"
)

func MainHandler(geoCoverage *terrain.GeoCoverage) http.Handler {
	mux := mux.NewRouter()

	colorHandler := configureHandler(color_v1.NewColorV1Provider(), geoCoverage)
	mux.HandleFunc("/color-v1/{z:[1-2]?[0-9]}/{y:[0-9]+}/{x:[0-9]+}.png", colorHandler)

	return mux
}

func configureHandler(provider colors.ColorProvider, geoCoverage *terrain.GeoCoverage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		vars := mux.Vars(r)

		z, err := strconv.ParseUint(vars["z"], 10, 8)
		if err != nil {
			http.Error(w, "Invalid zoom level", http.StatusBadRequest)
			return
		}

		if uint32(z) > provider.MaxZoom() {
			http.Error(w, "Zoom level too high", http.StatusBadRequest)
			return
		}

		maxScale := uint64(math.Pow(2, float64(z)))

		y, err := strconv.ParseUint(vars["y"], 10, 32)
		if err != nil || y >= maxScale {
			http.Error(w, "Invalid y coordinate", http.StatusBadRequest)
			return
		}

		x, err := strconv.ParseUint(vars["x"], 10, 32)
		if err != nil || x >= maxScale {
			http.Error(w, "Invalid x coordinate", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()

		log.Printf("/%s/%d/%d/%d.%s", provider.Name(), z, y, x, provider.FileType())
		defer func() {
			log.Printf("/%s/%d/%d/%d.%s - %s", provider.Name(), z, y, x, provider.FileType(), time.Since(start).Round(time.Millisecond))
		}()

		// Download subtiles
		elevationMap, err := terrain.GetElevationMapFromGeoTIFF(ctx, terrain.TileCoord{Z: uint32(z), Y: uint32(y), X: uint32(x)})
		if err != nil {
			http.Error(w, "Failed to get source data for tile", http.StatusInternalServerError)
			return
		}

		// Create tile bounds (calculation for pixel lat/lng mapping)
		tile := CreateTileBounds(uint32(z), uint32(y), uint32(x), elevationMap.TileSize)

		cells, err := GetCellsForTile(elevationMap, tile, geoCoverage)
		if err != nil {
			http.Error(w, "Failed to get cells for tile", http.StatusInternalServerError)
			return
		}

		// Convert cells ([][]*PixelCell) to [][]colors.DataCell
		dataMap := make([][]colors.DataCell, len(cells))
		for i, row := range cells {
			dataMap[i] = make([]colors.DataCell, len(row))
			for j, cell := range row {
				dataMap[i][j] = cell
			}
		}

		// Get color for each cell
		colorMap, err := provider.GetColor(ctx, colors.ColorInput{
			Zoom:    uint32(z),
			DataMap: dataMap,
		})
		if err != nil {
			http.Error(w, "Failed to get color pixels on tile", http.StatusInternalServerError)
			return
		}

		if ctx.Err() != nil {
			return
		}

		output := image.NewRGBA(image.Rect(0, 0, elevationMap.TileSize, elevationMap.TileSize))
		for y, row := range colorMap {
			for x, colorPoint := range row {
				output.Set(x, y, colorPoint.RGBA())
			}
		}

		if ctx.Err() != nil {
			return
		}

		w.WriteHeader(http.StatusOK)

		err = provider.EncodeImage(w, output)
		if err != nil {
			http.Error(w, "Failed to encode image", http.StatusInternalServerError)
			return
		}
	}
}
