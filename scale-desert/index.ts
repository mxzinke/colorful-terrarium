import { readFileSync, writeFileSync } from "fs";
import * as turf from "@turf/turf";
import * as geojson from "geojson";

const SCALE_DISTANCE = 30; // in kms
const INPUT_FILE = "deserts.geojson";

const INNER_FILE = "inner-deserts.geojson";
const OUTER_FILE = "outer-deserts.geojson";

function scaleGeometry(
  geometry: geojson.Polygon | geojson.MultiPolygon,
  distance: number,
): geojson.Polygon | geojson.MultiPolygon {
  try {
    // buffer the geometry by the given distance
    const buffered = turf.buffer(geometry, distance, {
      units: "kilometers",
      steps: 8, // higher number = smoother edges
    });

    // turf.buffer returns a Feature<Polygon|MultiPolygon>
    // we extract only the geometry
    return buffered.geometry as geojson.Polygon | geojson.MultiPolygon;
  } catch (error) {
    console.error("Error processing geometry:", error);
    return geometry;
  }
}

try {
  // GeoJSON einlesen
  const geojsonString = readFileSync(INPUT_FILE, "utf-8");
  const geojson: geojson.FeatureCollection = JSON.parse(geojsonString);

  // Features verarbeiten
  const outerFeatures = geojson.features.map((feature, idx) => {
    if (
      feature.geometry.type === "Polygon" ||
      feature.geometry.type === "MultiPolygon"
    ) {
      return {
        ...feature,
        id: idx.toString(),
        properties: {
          ...feature.properties,
          id: idx.toString(),
        },
        geometry: scaleGeometry(
          feature.geometry as geojson.Polygon | geojson.MultiPolygon,
          SCALE_DISTANCE,
        ),
      };
    }
    return {
      ...feature,
      id: idx.toString(),
      properties: {
        ...feature.properties,
        id: idx.toString(),
      },
    };
  });
  const innerFeatures = geojson.features.map((feature, idx) => {
    if (
      feature.geometry.type === "Polygon" ||
      feature.geometry.type === "MultiPolygon"
    ) {
      return {
        ...feature,
        id: idx.toString(),
        properties: {
          ...feature.properties,
          id: idx.toString(),
        },
        geometry: scaleGeometry(
          feature.geometry as geojson.Polygon | geojson.MultiPolygon,
          -SCALE_DISTANCE,
        ),
      };
    }
    return {
      ...feature,
      id: idx.toString(),
      properties: {
        ...feature.properties,
        id: idx.toString(),
      },
    };
  });

  writeFileSync(
    OUTER_FILE,
    JSON.stringify(
      {
        type: "FeatureCollection",
        features: outerFeatures,
      },
      null,
      2,
    ),
  );
  writeFileSync(
    INNER_FILE,
    JSON.stringify(
      {
        type: "FeatureCollection",
        features: innerFeatures,
      },
      null,
      2,
    ),
  );
  console.log(`Desert scaled by ${SCALE_DISTANCE}km.`);
} catch (error) {
  console.error("Error processing GeoJSON file:", error);
  process.exit(1);
}
