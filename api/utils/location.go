// SPDX-License-Identifier: GPL-2.0-or-later
package utils

import (
	"github.com/klahr/300sf/api/log"
	"github.com/klahr/300sf/api/models"
	"github.com/klahr/300sf/api/types"
	"github.com/tidwall/geojson"
	"github.com/tidwall/geojson/geometry"
)

// GetZonesFromPosition returns the country, county and municipality zones
// containing position, from the outermost level inwards.
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

func zoneContains(zone models.Zone, position *geojson.Point) bool {
	obj, err := geojson.Parse(string(zone.Polygon), geojson.DefaultParseOptions)
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
	q := models.DB.Where("admin_level=?", level)
	if parentZone != nil {
		q = q.Where("parent_zone_id=?", parentZone.Id)
	} else {
		q = q.Where("parent_zone_id IS NULL")
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

	switch level {
	case types.AdminLevelCountry:
		level = types.AdminLevelCounty
	case types.AdminLevelCounty:
		level = types.AdminLevelMunicipality
	default:
		return nil
	}

	return getZonesFromPosition(position, zones, level, found)
}
