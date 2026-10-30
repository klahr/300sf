// SPDX-License-Identifier: AGPL-3.0-or-later
package utils

import (
	"sync"

	"github.com/klahr/300sf/server/log"
	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/types"
	"github.com/tidwall/geojson"
	"github.com/tidwall/geojson/geometry"
)

// GetZonesFromPosition returns the zones containing position, from the
// country inwards: usually country, county and municipality, but a country
// without counties goes straight to its municipalities.
func GetZonesFromPosition(position types.LatLon) []models.Zone {
	var zones []models.Zone
	p := geojson.NewPoint(geometry.Point{
		X: position.Longitude,
		Y: position.Latitude,
	})
	if err := getZonesFromPosition(p, &zones, types.AdminLevelCountry, nil); err != nil {
		log.Error(nil, 0, err.Error())
	}
	return zones
}

// Parsed outlines by zone id. A zone's outline never changes, and parsing a
// country's for every sighting would be slow.
var outlines sync.Map

func outline(zone models.Zone) (geojson.Object, error) {
	if obj, ok := outlines.Load(zone.Id); ok {
		return obj.(geojson.Object), nil
	}
	obj, err := geojson.Parse(string(zone.Polygon), geojson.DefaultParseOptions)
	if err != nil {
		return nil, err
	}
	if zone.Id != 0 {
		outlines.Store(zone.Id, obj)
	}
	return obj, nil
}

func zoneContains(zone models.Zone, position *geojson.Point) bool {
	obj, err := outline(zone)
	if err != nil {
		log.Error(nil, 0, "zone "+zone.Uuid+": "+err.Error())
		return false
	}
	var children []geojson.Object
	if fc, ok := obj.(*geojson.FeatureCollection); ok {
		children = fc.Children()
	} else {
		children = []geojson.Object{obj}
	}
	for _, child := range children {
		if f, ok := child.(*geojson.Feature); ok {
			child = f.Base()
		}
		switch g := child.(type) {
		case *geojson.MultiPolygon:
			for _, polygon := range g.Children() {
				if polygon.Contains(position) {
					return true
				}
			}
		case *geojson.Polygon:
			if g.Contains(position) {
				return true
			}
		}
	}
	return false
}

func getZonesFromPosition(position *geojson.Point, zones *[]models.Zone, level int, parentZone *models.Zone) error {
	var q = models.DB
	if parentZone != nil {
		q = q.Where("parent_zone_id=?", parentZone.Id)
	} else {
		q = q.Where("admin_level=?", level).Where("parent_zone_id IS NULL")
	}

	var candidates []models.Zone
	if err := q.Preload("Name").Find(&candidates).Error; err != nil {
		return err
	}

	var found *models.Zone
	for i := range candidates {
		if zoneContains(candidates[i], position) {
			found = &candidates[i]
			break
		}
	}
	if found == nil {
		return nil
	}
	*zones = append(*zones, *found)
	if found.AdminLevel >= types.AdminLevelMunicipality {
		return nil
	}
	return getZonesFromPosition(position, zones, 0, found)
}

// ZoneBounds is the box around a zone's outline, for finding the zones in
// view on a map.
func ZoneBounds(zone models.Zone) (geometry.Rect, bool) {
	obj, err := outline(zone)
	if err != nil {
		return geometry.Rect{}, false
	}
	return obj.Rect(), true
}
