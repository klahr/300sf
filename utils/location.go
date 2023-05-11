package utils

import (
	"git.sr.ht/~klahr/300sf/models"
	"git.sr.ht/~klahr/300sf/types"
	"github.com/tidwall/geojson"
	"github.com/tidwall/geojson/geometry"
	"gorm.io/gorm"
)

func GetZonesFromPosition(position types.LatLon) []models.Zone {
	tx := models.DB.Begin()
	var zones []models.Zone
	p := geojson.NewPoint(geometry.Point{
		X: position.Longitude,
		Y: position.Latitude,
	})
	getZonesFromPosition(tx, p, &zones, types.AdminLevelCountry, nil)
	return zones
}

func getZonesFromPosition(tx *gorm.DB, position *geojson.Point, zones *[]models.Zone, level int, parentZone *models.Zone) error {
	q := tx.Where("admin_level=?", level)

	if parentZone != nil {
		q.Where("parent_zone_id=?", parentZone.Id)
	} else {
		q.Where("parent_zone_id IS NULL")
	}

	var z []models.Zone
	if err := q.Preload("Name").Find(&z).Error; err != nil {
		return err
	}

	var newParentZone *models.Zone
	for _, zone := range z {
		fc, _ := geojson.Parse(string(zone.Polygon), geojson.DefaultParseOptions)
		realFc := fc.(*geojson.FeatureCollection)
		for _, feature := range realFc.Children() {
			realFeature := feature.(*geojson.Feature)
			if mp, ok := realFeature.Base().(*geojson.MultiPolygon); ok {
				for _, polygon := range mp.Children() {
					p := polygon.(*geojson.Polygon)
					if p.Contains(position) {
						newParentZone = &zone
						*zones = append(*zones, zone)
						break
					}
				}
				if newParentZone != nil {
					break
				}
			} else if p, ok := realFeature.Base().(*geojson.Polygon); ok {
				if p.Contains(position) {
					newParentZone = &zone
					*zones = append(*zones, zone)
					break
				}
			}
		}
		if newParentZone != nil {
			break
		}
	}

	if newParentZone == nil {
		return nil
	}

	switch level {
	case types.AdminLevelCountry:
		level = types.AdminLevelCounty
	case types.AdminLevelCounty:
		level = types.AdminLevelMunicipality
	default:
		return nil
	}

	return getZonesFromPosition(tx, position, zones, level, newParentZone)
}
