// SPDX-License-Identifier: AGPL-3.0-or-later
package utils

import (
	"encoding/json"
	"errors"
	"math"
	"sync"

	"github.com/klahr/300sf/server/models"
)

// Outlines for maps by zone id: zones added by hand can be megabytes, far
// more than a map of a whole county needs.
var mapOutlines sync.Map

// mapTolerance drops detail smaller than about 500 m.
const mapTolerance = 0.005

// ZoneOutlineForMap is the zone's outline as a simplified GeoJSON
// MultiPolygon, for drawing.
func ZoneOutlineForMap(zone models.Zone) (json.RawMessage, error) {
	if o, ok := mapOutlines.Load(zone.Id); ok {
		return o.(json.RawMessage), nil
	}
	polygons, err := polygonsOf(zone.Polygon)
	if err != nil {
		return nil, err
	}
	var out [][][][2]float64
	for _, polygon := range polygons {
		if len(polygon) == 0 {
			continue
		}
		outer := simplifyRing(polygon[0], mapTolerance)
		if len(outer) < 4 {
			// An islet smaller than the tolerance.
			continue
		}
		rings := [][][2]float64{outer}
		for _, hole := range polygon[1:] {
			if h := simplifyRing(hole, mapTolerance); len(h) >= 4 {
				rings = append(rings, h)
			}
		}
		out = append(out, rings)
	}
	raw, err := json.Marshal(struct {
		Type        string           `json:"type"`
		Coordinates [][][][2]float64 `json:"coordinates"`
	}{"MultiPolygon", out})
	if err != nil {
		return nil, err
	}
	if zone.Id != 0 {
		mapOutlines.Store(zone.Id, json.RawMessage(raw))
	}
	return raw, nil
}

type geoJSON struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"`
	Geometry    *geoJSON        `json:"geometry"`
	Features    []geoJSON       `json:"features"`
	Geometries  []geoJSON       `json:"geometries"`
}

// polygonsOf collects the polygons of any GeoJSON: a geometry, a feature or
// a collection of them.
func polygonsOf(raw []byte) ([][][][2]float64, error) {
	var g geoJSON
	if err := json.Unmarshal(raw, &g); err != nil {
		return nil, err
	}
	var polygons [][][][2]float64
	var walk func(g geoJSON) error
	walk = func(g geoJSON) error {
		switch g.Type {
		case "Polygon":
			var p [][][2]float64
			if err := json.Unmarshal(g.Coordinates, &p); err != nil {
				return err
			}
			polygons = append(polygons, p)
		case "MultiPolygon":
			var mp [][][][2]float64
			if err := json.Unmarshal(g.Coordinates, &mp); err != nil {
				return err
			}
			polygons = append(polygons, mp...)
		case "Feature":
			if g.Geometry != nil {
				return walk(*g.Geometry)
			}
		case "FeatureCollection":
			for _, f := range g.Features {
				if err := walk(f); err != nil {
					return err
				}
			}
		case "GeometryCollection":
			for _, c := range g.Geometries {
				if err := walk(c); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := walk(g); err != nil {
		return nil, err
	}
	if len(polygons) == 0 {
		return nil, errors.New("no polygons")
	}
	return polygons, nil
}

// simplifyRing drops the points of a closed ring that stray less than
// tolerance from the line between their neighbours (Douglas–Peucker).
func simplifyRing(ring [][2]float64, tolerance float64) [][2]float64 {
	if len(ring) < 4 {
		return ring
	}
	keep := make([]bool, len(ring))
	keep[0], keep[len(ring)-1] = true, true
	// A closed ring starts and ends on one point; split it at the point
	// farthest from there so both halves have a line to measure against.
	far, farthest := 0, -1.0
	for i, p := range ring {
		if d := math.Hypot(p[0]-ring[0][0], p[1]-ring[0][1]); d > farthest {
			far, farthest = i, d
		}
	}
	keep[far] = true
	var mark func(a, b int)
	mark = func(a, b int) {
		index, most := -1, tolerance
		for i := a + 1; i < b; i++ {
			if d := distanceToSegment(ring[i], ring[a], ring[b]); d > most {
				index, most = i, d
			}
		}
		if index >= 0 {
			keep[index] = true
			mark(a, index)
			mark(index, b)
		}
	}
	mark(0, far)
	mark(far, len(ring)-1)
	var out [][2]float64
	for i, p := range ring {
		if keep[i] {
			out = append(out, p)
		}
	}
	return out
}

func distanceToSegment(p, a, b [2]float64) float64 {
	dx, dy := b[0]-a[0], b[1]-a[1]
	if dx == 0 && dy == 0 {
		return math.Hypot(p[0]-a[0], p[1]-a[1])
	}
	t := math.Max(0, math.Min(1, ((p[0]-a[0])*dx+(p[1]-a[1])*dy)/(dx*dx+dy*dy)))
	return math.Hypot(p[0]-a[0]-t*dx, p[1]-a[1]-t*dy)
}
