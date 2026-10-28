package controllers

import (
	"errors"

	"github.com/klahr/300sf/server/models"
)

// SiteStats are the community numbers shown on the landing page.
type SiteStats struct {
	Birders   int64
	Sightings int64
	Species   int64
}

func GetSiteStats() (*SiteStats, error) {
	if models.DB == nil {
		return nil, errors.New("no database")
	}
	var s SiteStats
	if err := models.DB.Table("users").Where("deleted_at IS NULL AND deactivated_at IS NULL").Count(&s.Birders).Error; err != nil {
		return nil, err
	}
	if err := models.DB.Table("reports").Where("archived IS NULL").Count(&s.Sightings).Error; err != nil {
		return nil, err
	}
	if err := models.DB.Table("reports").Where("archived IS NULL").Distinct("species_id").Count(&s.Species).Error; err != nil {
		return nil, err
	}
	return &s, nil
}
