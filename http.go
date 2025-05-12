package main

import (
	"context"
	"fmt"
	"image"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/mxzinke/colorful-terrarium/colors"
	"github.com/mxzinke/colorful-terrarium/colors/color_v1"
	"github.com/mxzinke/colorful-terrarium/colors/color_v2"
	mono_terrain "github.com/mxzinke/colorful-terrarium/colors/mono-terrain"
	"github.com/mxzinke/colorful-terrarium/colors/terrarium"
	"github.com/mxzinke/colorful-terrarium/terrain"
)

func MainHandler(geoCoverage *terrain.GeoCoverage) http.Handler {
	mux := mux.NewRouter()

	providers := []colors.ColorProvider{
		color_v1.NewColorV1Provider(),
		color_v2.NewColorV2Provider(),
		terrarium.NewLandTerrariumProfile(),
		terrarium.NewWaterTerrariumProfile(),
		mono_terrain.NewLandMonoTerrainProfile(),
		mono_terrain.NewWaterMonoTerrainProfile(),
	}

	for _, provider := range providers {
		handler := configureHandler(provider, geoCoverage)
		mux.HandleFunc(fmt.Sprintf("/%s/{z:[1-2]?[0-9]}/{y:[0-9]+}/{x:[0-9]+}.%s", provider.Name(), provider.FileType()), handler)
	}

	terrariumLandHandler := configureHandler(terrarium.NewLandTerrariumProfile(), geoCoverage)
	mux.HandleFunc("/terrarium-land/{z:[1-2]?[0-9]}/{y:[0-9]+}/{x:[0-9]+}.png", terrariumLandHandler)

	terrariumWaterHandler := configureHandler(terrarium.NewWaterTerrariumProfile(), geoCoverage)
	mux.HandleFunc("/terrarium-water/{z:[1-2]?[0-9]}/{y:[0-9]+}/{x:[0-9]+}.png", terrariumWaterHandler)

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
		elevationMap, err := terrain.GetElevationMapForTerrarium(ctx, terrain.TileCoord{Z: uint32(z), Y: uint32(y), X: uint32(x)})
		if err != nil {
			http.Error(w, "Failed to get source data for tile", http.StatusInternalServerError)
			return
		}

		// Create tile bounds (calculation for pixel lat/lng mapping)
		tile := CreateTileBounds(uint32(z), uint32(y), uint32(x), elevationMap.TileSize)

		// Fixing the elevation data on some parts of the world
		fixElevationMap(elevationMap, tile, geoCoverage)

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
		imgRect := image.Rect(0, 0, elevationMap.TileSize, elevationMap.TileSize)

		img, err := provider.GetImage(ctx, imgRect, colors.ColorInput{
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

		w.WriteHeader(http.StatusOK)

		err = provider.EncodeImage(w, img)
		if err != nil {
			http.Error(w, "Failed to encode image", http.StatusInternalServerError)
			return
		}
	}
}
