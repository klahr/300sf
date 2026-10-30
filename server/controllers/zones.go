// SPDX-License-Identifier: AGPL-3.0-or-later
package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/server/log"
	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/places"
	"github.com/klahr/300sf/server/table"
	"github.com/klahr/300sf/server/utils"
	"github.com/tidwall/geojson"
	"github.com/tidwall/geojson/geometry"
	"gopkg.in/guregu/null.v4/zero"
	"gorm.io/gorm"
)

func Zones(c *gin.Context) {
	type request struct {
		Name           models.Translation `json:"name"`
		Polygon        string             `json:"polygon"`
		AdminLevel     int64              `json:"admin_level"`
		ParentZoneUuid string             `json:"parent_zone_uuid"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		log.Error(c, 0, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "could not parse request",
			"code":    ErrorParse,
		})
		return
	}

	if r.AdminLevel >= models.ZoneOutOfRange || r.AdminLevel <= models.ZoneTypeUnknown {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "zone type is out of range",
		})
		return
	}

	tx := models.DB.Begin()
	defer tx.Rollback()

	var user models.User
	if err := tx.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	if _, err := geojson.Parse(r.Polygon, nil); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": "invalid polygon",
		})
		tx.Rollback()
		return
	}

	name := r.Name
	name.Timestamp = time.Now()
	name.Uuid = utils.GenerateShortUuid()
	if err := tx.Create(&name).Error; err != nil {
		log.Error(c, 0, err.Error())
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "could not insert translation",
			"code":    ErrorInsert,
		})
		return
	}

	zone := models.Zone{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateShortUuid(),
		},
		NameId:     name.Id,
		Polygon:    []byte(r.Polygon),
		AdminLevel: r.AdminLevel,
	}

	if r.ParentZoneUuid != "" {
		var parentZone models.Zone
		if err := tx.Where("uuid=?", r.ParentZoneUuid).First(&parentZone).Error; err == nil {
			zone.ParentZoneId = zero.IntFrom(parentZone.Id)
		}
	}

	if err := tx.Create(&zone).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, zone)
}

func ZonesList(c *gin.Context) {
	var zones []models.Zone
	// Without the polygons, which can be megabytes each.
	if err := models.DB.Table(table.Zones).Omit("polygon").Preload("Name").Preload("ParentZone", func(db *gorm.DB) *gorm.DB { return db.Omit("polygon") }).Preload("ParentZone.Name").Find(&zones).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	type response struct {
		Result []models.Zone `json:"result"`
	}

	c.JSON(http.StatusOK, response{
		Result: zones,
	})
}

// A county or region as competitions choose them: on a map, or found by name.
type regionZone struct {
	Uuid    string             `json:"uuid,omitempty"`
	OsmId   string             `json:"osm_id,omitempty"`
	Name    models.Translation `json:"name"`
	Country string             `json:"country,omitempty"`
	Outline json.RawMessage    `json:"outline,omitempty"`
}

func regionOf(zone models.Zone) regionZone {
	r := regionZone{Uuid: zone.Uuid, Name: zone.Name}
	if outline, err := utils.ZoneOutlineForMap(zone); err == nil {
		r.Outline = outline
	}
	return r
}

// Boxes around the regions' outlines by zone id: regions never move.
var regionBounds sync.Map

const maxRegionsInView = 200

// ZonesRegions returns the regions (zone level 4) in view on a map,
// bbox=west,south,east,north, with outlines simplified for drawing.
func ZonesRegions(c *gin.Context) {
	var box [4]float64
	parts := strings.Split(c.Query("bbox"), ",")
	if len(parts) != 4 {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrorParse, "message": "bbox=west,south,east,north"})
		return
	}
	for i, p := range parts {
		v, err := strconv.ParseFloat(p, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": ErrorParse, "message": "bbox=west,south,east,north"})
			return
		}
		box[i] = v
	}
	view := geometry.Rect{Min: geometry.Point{X: box[0], Y: box[1]}, Max: geometry.Point{X: box[2], Y: box[3]}}

	var ids []int64
	if err := models.DB.Table(table.Zones).Where("admin_level=?", models.ZoneTypeCounty).Pluck("id", &ids).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	var missing []int64
	for _, id := range ids {
		if _, ok := regionBounds.Load(id); !ok {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		var zones []models.Zone
		models.DB.Select("id", "uuid", "polygon").Where("id IN ?", missing).Find(&zones)
		for _, z := range zones {
			if rect, ok := utils.ZoneBounds(z); ok {
				regionBounds.Store(z.Id, rect)
			}
		}
	}
	var inView []int64
	for _, id := range ids {
		if rect, ok := regionBounds.Load(id); ok && rect.(geometry.Rect).IntersectsRect(view) {
			inView = append(inView, id)
			if len(inView) == maxRegionsInView {
				break
			}
		}
	}

	result := []regionZone{}
	if len(inView) > 0 {
		var zones []models.Zone
		if err := models.DB.Where("id IN ?", inView).Preload("Name").Find(&zones).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		for _, z := range zones {
			result = append(result, regionOf(z))
		}
	}
	c.JSON(http.StatusOK, gin.H{"result": result})
}

// ZonesSearch finds regions by name on OpenStreetMap, for choosing ones no
// sighting has made a zone of yet. People search on purpose, never as they
// type, and each may search a few times a minute: the whole site shares the
// geocoder's one request a second.
func ZonesSearch(c *gin.Context) {
	text := strings.TrimSpace(c.Query("q"))
	if len(text) < 2 || len(text) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrorParse, "message": "q must be 2 to 100 characters"})
		return
	}
	if !places.Enabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"message": "place search is off"})
		return
	}
	if Throttled("zonesearch/"+c.GetString("uuid"), 10, time.Minute) {
		c.JSON(http.StatusTooManyRequests, gin.H{"message": "too many searches, try again in a minute"})
		return
	}
	found, err := places.SearchRegions(c.Request.Context(), text)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"message": "place search failed"})
		return
	}
	result := []regionZone{}
	for _, f := range found {
		r := regionZone{OsmId: f.OsmId, Country: f.Country, Name: models.Translation{
			Sv: zero.StringFrom(f.LocalName("sv")),
			En: zero.StringFrom(f.LocalName("en")),
			De: zero.StringFrom(f.LocalName("de")),
		}}
		var zone models.Zone
		if models.DB.Select("uuid").Where("osm_id=?", f.OsmId).Limit(1).Find(&zone); zone.Uuid != "" {
			r.Uuid = zone.Uuid
		}
		result = append(result, r)
	}
	c.JSON(http.StatusOK, gin.H{"result": result})
}

// ZonesFromOsm returns the zone of a region found with ZonesSearch, making
// it first if needed.
func ZonesFromOsm(c *gin.Context) {
	var r struct {
		OsmId string `json:"osm_id"`
	}
	if err := c.ShouldBindJSON(&r); err != nil || r.OsmId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": ErrorParse, "message": "osm_id is missing"})
		return
	}
	if Throttled("zonesearch/"+c.GetString("uuid"), 10, time.Minute) {
		c.JSON(http.StatusTooManyRequests, gin.H{"message": "too many searches, try again in a minute"})
		return
	}
	zone, err := places.RegionZone(c.Request.Context(), r.OsmId)
	if errors.Is(err, places.ErrDisabled) {
		c.JSON(http.StatusServiceUnavailable, gin.H{"message": "place search is off"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"message": "could not fetch the region"})
		return
	}
	if err := models.DB.Preload("Name").First(zone, zone.Id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, regionOf(*zone))
}
