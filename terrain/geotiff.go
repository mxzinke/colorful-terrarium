package terrain

import (
	"bytes"
	"fmt"
	"image/color"
	"io"
	"net/http"

	tiff "github.com/chai2010/tiff"
)

const geotiffSourceURL = "https://elevation-tiles-prod.s3.amazonaws.com/geotiff/%d/%d/%d.tif"

func GetElevationMapFromGeoTIFF(coord TileCoord) (*ElevationMap, error) {
	// GeoTIFF fetching...
	resp, err := http.Get(fmt.Sprintf(geotiffSourceURL, coord.Z, coord.Y, coord.X))
	if err != nil {
		return nil, fmt.Errorf("failed to download GeoTIFF: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read TIFF: %v", err)
	}

	// Create a new TIFF reader
	matrix, tileSize, err := readTIFFToFloat32Matrix(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode TIFF: %v", err)
	}

	// Copy over data to ElevationMap
	em := &ElevationMap{
		Data:     matrix,
		TileSize: tileSize,
	}

	return em, nil
}

func readTIFFToFloat32Matrix(data []byte) ([][]float32, int, error) {
	// Decode TIFF
	images, errors, err := tiff.DecodeAll(bytes.NewReader(data))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode TIFF: %v", err)
	}

	// We'll process the first image/band
	if len(images) == 0 || len(images[0]) == 0 {
		return nil, 0, fmt.Errorf("no images found in TIFF")
	}
	if errors[0][0] != nil {
		return nil, 0, fmt.Errorf("error in first image: %v", errors[0][0])
	}

	img := images[0][0]
	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	// Create result matrix
	matrix := make([][]float32, height)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		matrix[y] = make([]float32, width)

		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			height := img.At(x, y).(color.Gray16)
			matrix[y][x] = float32(int16(height.Y))
		}
	}

	return matrix, img.Bounds().Max.X, nil
}