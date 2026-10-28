// SPDX-License-Identifier: GPL-2.0-or-later
package utils

import (
	"testing"

	"github.com/klahr/300sf/server/models"
	"github.com/tidwall/geojson"
	"github.com/tidwall/geojson/geometry"
)

// Central Stockholm (Gamla stan and around).
const stockholmRing = `[[[18.072986057402943,59.32811225957224],[18.068345175322833,59.32682199593981],[18.065815894589605,59.3268930209002],[18.064493243196836,59.32625379091354],[18.066651253364086,59.324714854592486],[18.065885507820667,59.32361388046536],[18.067416998907532,59.32220505510449],[18.068437992964505,59.322702294252935],[18.0699694840514,59.32237080229547],[18.071245726622863,59.32232344603784],[18.074471139668503,59.32126975223525],[18.075979426344702,59.32285620013397],[18.076791580708544,59.3253185994561],[18.07528329403229,59.32699955806231],[18.072986057402943,59.32811225957224]]]`

func point(lat, lon float64) *geojson.Point {
	return geojson.NewPoint(geometry.Point{X: lon, Y: lat})
}

func TestZoneContains(t *testing.T) {
	inside := point(59.32455958040549, 18.067933630966028)
	outside := point(59.3293, 18.0686) // Norrmalm, north of the polygon

	shapes := map[string]string{
		"feature collection": `{"type":"FeatureCollection","features":[{"type":"Feature","properties":{},"geometry":{"type":"Polygon","coordinates":` + stockholmRing + `}}]}`,
		"multipolygon":       `{"type":"FeatureCollection","features":[{"type":"Feature","properties":{},"geometry":{"type":"MultiPolygon","coordinates":[` + stockholmRing + `]}}]}`,
		"bare feature":       `{"type":"Feature","properties":{},"geometry":{"type":"Polygon","coordinates":` + stockholmRing + `}}`,
		"bare polygon":       `{"type":"Polygon","coordinates":` + stockholmRing + `}`,
	}
	for name, shape := range shapes {
		zone := models.Zone{Polygon: []byte(shape)}
		if !zoneContains(zone, inside) {
			t.Errorf("%s: inside point not contained", name)
		}
		if zoneContains(zone, outside) {
			t.Errorf("%s: outside point contained", name)
		}
	}
}

func TestZoneContainsInvalid(t *testing.T) {
	for _, shape := range []string{"", "not json", `{"type":"Point","coordinates":[18.07,59.32]}`} {
		if zoneContains(models.Zone{Polygon: []byte(shape)}, point(59.3245, 18.0679)) {
			t.Errorf("%q: should not contain anything", shape)
		}
	}
}
