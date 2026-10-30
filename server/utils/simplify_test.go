// SPDX-License-Identifier: AGPL-3.0-or-later
package utils

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/klahr/300sf/server/models"
)

// A circle of n points with radius 1 degree, as a closed ring.
func circle(n int, cx float64) string {
	var pts []string
	for i := 0; i <= n; i++ {
		a := 2 * math.Pi * float64(i%n) / float64(n)
		pts = append(pts, fmt.Sprintf("[%f,%f]", cx+math.Cos(a), 60+math.Sin(a)))
	}
	return "[" + strings.Join(pts, ",") + "]"
}

func TestZoneOutlineForMap(t *testing.T) {
	// A detailed county and an islet far smaller than the tolerance, as a feature collection.
	islet := `[[30,60],[30.001,60],[30.001,60.001],[30,60.001],[30,60]]`
	raw := `{"type":"FeatureCollection","features":[
		{"type":"Feature","properties":{},"geometry":{"type":"Polygon","coordinates":[` + circle(5000, 15) + `]}},
		{"type":"Feature","properties":{},"geometry":{"type":"Polygon","coordinates":[` + islet + `]}}]}`
	out, err := ZoneOutlineForMap(models.Zone{Polygon: []byte(raw)})
	if err != nil {
		t.Fatal(err)
	}
	var mp struct {
		Type        string
		Coordinates [][][][2]float64
	}
	if err := json.Unmarshal(out, &mp); err != nil {
		t.Fatal(err)
	}
	if mp.Type != "MultiPolygon" || len(mp.Coordinates) != 1 {
		t.Fatalf("type %s with %d polygons, want the county alone", mp.Type, len(mp.Coordinates))
	}
	ring := mp.Coordinates[0][0]
	if len(ring) < 8 || len(ring) > 200 {
		t.Errorf("5000 points became %d", len(ring))
	}
	if ring[0] != ring[len(ring)-1] {
		t.Error("the ring should stay closed")
	}

	if _, err := ZoneOutlineForMap(models.Zone{Polygon: []byte(`{"type":"Point","coordinates":[1,2]}`)}); err == nil {
		t.Error("a point has no outline")
	}
}
