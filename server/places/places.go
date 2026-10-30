// SPDX-License-Identifier: AGPL-3.0-or-later

// Package places gives sightings their zones anywhere in the world. A
// sighting inside known zones gets them right away (utils.GetZonesFromPosition);
// one in a new place is looked up with the geocoder after it is saved, and
// the country, region and municipality found become zones, so the next
// sighting there needs no lookup.
package places

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/klahr/300sf/server/geocoder"
	"github.com/klahr/300sf/server/log"
	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/table"
	"github.com/klahr/300sf/server/types"
	"github.com/klahr/300sf/server/utils"
	"gopkg.in/guregu/null.v4/zero"
	"gorm.io/gorm"
)

var (
	client *geocoder.Client
	jobs   = make(chan int64, 1000)
	// Zones are made one at a time, so the same area is never made twice.
	resolving sync.Mutex
	// Places looked up lately without finding a municipality (the sea, a
	// country without one), so they aren't asked about again and again.
	tried sync.Map
	// Sightings the backfill has done since the server started.
	backfilled sync.Map
)

const (
	retryAfter       = 24 * time.Hour
	backfillInterval = 15 * time.Minute
	backfillBatch    = 20
)

// Start looks up new places with c; without a call, sightings in new places
// get no zone.
func Start(c *geocoder.Client) {
	client = c
	go func() {
		for id := range jobs {
			locate(context.Background(), id, client.Lookup)
		}
	}()
	go backfill()
}

// Locate gives a just-saved sighting its zone in the background, when no
// zone knew its place. It never blocks; a full queue leaves it to the backfill.
func Locate(reportId int64) {
	if client == nil {
		return
	}
	select {
	case jobs <- reportId:
	default:
	}
}

// Complete says whether zones (from GetZonesFromPosition) reach down to a
// municipality, so there is nothing to look up.
func Complete(zones []models.Zone) bool {
	return len(zones) > 0 && zones[len(zones)-1].AdminLevel >= types.AdminLevelMunicipality
}

type lookupFunc func(ctx context.Context, lat, lon float64) (geocoder.Areas, error)

func locate(ctx context.Context, reportId int64, lookup lookupFunc) {
	var report models.Report
	if err := models.DB.Select("id", "latitude", "longitude", "zone_id").First(&report, reportId).Error; err != nil {
		return
	}
	if !report.Latitude.Valid || !report.Longitude.Valid || (report.Latitude.Float64 == 0 && report.Longitude.Float64 == 0) {
		return
	}
	lat, lon := report.Latitude.Float64, report.Longitude.Float64

	// Another sighting may have made the zones since this one was saved.
	zones := utils.GetZonesFromPosition(types.LatLon{Latitude: lat, Longitude: lon})
	known := func() {
		if len(zones) > 0 {
			setZone(report.Id, zones[len(zones)-1].Id)
		}
	}
	if Complete(zones) {
		known()
		return
	}
	key := placeKey(lat, lon)
	if at, ok := tried.Load(key); ok && time.Since(at.(time.Time)) < retryAfter {
		known()
		return
	}

	areas, err := lookup(ctx, lat, lon)
	if err != nil {
		// Left without a zone; the backfill tries again.
		log.Error(nil, 0, "geocoder: "+err.Error())
		return
	}
	if areas.Municipality == nil {
		tried.Store(key, time.Now())
	}
	resolving.Lock()
	zone, err := resolve(areas)
	resolving.Unlock()
	if err != nil {
		log.Error(nil, 0, "places: "+err.Error())
		return
	}
	if zone == nil {
		known()
		return
	}
	setZone(report.Id, zone.Id)
}

// About a kilometre: one lookup that found nothing does for the area around it.
func placeKey(lat, lon float64) string {
	return fmt.Sprintf("%.2f,%.2f", math.Round(lat*100)/100, math.Round(lon*100)/100)
}

func setZone(reportId, zoneId int64) {
	models.DB.Table(table.Reports).Where("id=?", reportId).Update("zone_id", zoneId)
}

// resolve makes zones of the areas, from the country in, and returns the most
// local one. Nil without a country: zones hang from their country.
func resolve(areas geocoder.Areas) (*models.Zone, error) {
	if areas.Country == nil {
		return nil, nil
	}
	var parent *models.Zone
	for _, step := range []struct {
		area *geocoder.Area
		tier int64
	}{
		{areas.Country, types.AdminLevelCountry},
		{areas.Region, types.AdminLevelCounty},
		{areas.Municipality, types.AdminLevelMunicipality},
	} {
		if step.area == nil {
			continue
		}
		zone, err := zoneFor(*step.area, step.tier, parent)
		if err != nil {
			return nil, err
		}
		parent = zone
	}
	return parent, nil
}

// zoneFor finds the area's zone or makes it. A zone added by hand before,
// with the same name and level, becomes this area's rather than a twin.
func zoneFor(area geocoder.Area, tier int64, parent *models.Zone) (*models.Zone, error) {
	var zone models.Zone
	if err := models.DB.Where("osm_id=?", area.OsmId).Limit(1).Find(&zone).Error; err != nil {
		return nil, err
	}
	if zone.Id != 0 {
		return &zone, nil
	}

	names := names(area)
	q := models.DB.Table(table.Zones).Select(table.Zones+".*").
		Joins("JOIN "+table.Translations+" t ON t.id="+table.Zones+".name_id").
		Where(table.Zones+".osm_id IS NULL").Where(table.Zones+".admin_level=?", tier).
		Where("t.sv IN ? OR t.en IN ? OR t.de IN ?", names, names, names)
	if parent != nil {
		q = q.Where(table.Zones+".parent_zone_id=?", parent.Id)
	} else {
		q = q.Where(table.Zones + ".parent_zone_id IS NULL")
	}
	if err := q.Limit(1).Find(&zone).Error; err != nil {
		return nil, err
	}
	if zone.Id != 0 {
		err := models.DB.Model(&zone).Update("osm_id", area.OsmId).Error
		return &zone, err
	}

	err := models.DB.Transaction(func(tx *gorm.DB) error {
		name := models.Translation{
			Model: newModel(),
			Sv:    zero.StringFrom(area.LocalName("sv")),
			En:    zero.StringFrom(area.LocalName("en")),
			De:    zero.StringFrom(area.LocalName("de")),
		}
		if err := tx.Create(&name).Error; err != nil {
			return err
		}
		zone = models.Zone{
			Model:      newModel(),
			NameId:     name.Id,
			Polygon:    area.Geometry,
			AdminLevel: tier,
			OsmId:      zero.StringFrom(area.OsmId),
		}
		if parent != nil {
			zone.ParentZoneId = zero.IntFrom(parent.Id)
		}
		return tx.Omit("Name", "ParentZone").Create(&zone).Error
	})
	return &zone, err
}

func names(area geocoder.Area) []string {
	seen := map[string]bool{}
	var list []string
	for _, n := range []string{area.Name, area.LocalName("sv"), area.LocalName("en"), area.LocalName("de")} {
		if n != "" && !seen[n] {
			seen[n] = true
			list = append(list, n)
		}
	}
	return list
}

func newModel() models.Model {
	return models.Model{Timestamp: time.Now(), Uuid: utils.GenerateShortUuid()}
}

// backfill gives zones to sightings that have a position but none: saved
// while the geocoder was down, or before this existed. It runs at intervals,
// so it keeps to the slower pace the geocoder's policy asks of such work.
func backfill() {
	time.Sleep(time.Minute)
	for {
		var ids []int64
		models.DB.Table(table.Reports).Where("latitude IS NOT NULL AND longitude IS NOT NULL AND zone_id IS NULL").
			Order("id DESC").Limit(backfillBatch*5).Pluck("id", &ids)
		done := 0
		for _, id := range ids {
			if _, ok := backfilled.LoadOrStore(id, true); ok {
				continue
			}
			locate(context.Background(), id, client.Slowly)
			if done++; done >= backfillBatch {
				break
			}
		}
		time.Sleep(backfillInterval)
	}
}

// Enabled says whether places can be looked up: geocoder-endpoint is set.
func Enabled() bool {
	return client != nil
}

// SearchRegions finds regions by name for someone choosing them, such as
// the counties a competition counts in.
func SearchRegions(ctx context.Context, text string) ([]geocoder.Candidate, error) {
	if client == nil {
		return nil, ErrDisabled
	}
	return client.Search(ctx, text)
}

// RegionZone returns the zone of a region chosen from SearchRegions, making
// it (and its country) when no sighting has yet.
func RegionZone(ctx context.Context, osmId string) (*models.Zone, error) {
	var zone models.Zone
	if err := models.DB.Where("osm_id=?", osmId).Limit(1).Find(&zone).Error; err != nil {
		return nil, err
	}
	if zone.Id != 0 {
		return &zone, nil
	}
	if client == nil {
		return nil, ErrDisabled
	}
	areas, err := client.Region(ctx, osmId)
	if err != nil {
		return nil, err
	}
	resolving.Lock()
	defer resolving.Unlock()
	made, err := resolve(areas)
	if err == nil && made == nil {
		err = fmt.Errorf("places: no country around %s", osmId)
	}
	return made, err
}

var ErrDisabled = errors.New("places: no geocoder-endpoint")
