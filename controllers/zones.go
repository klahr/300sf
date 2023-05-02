package controllers

import (
	"fmt"
	"net/http"
	"time"

	"git.sr.ht/~klahr/300sf/models"
	"git.sr.ht/~klahr/300sf/table"
	"git.sr.ht/~klahr/300sf/utils"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/geojson"
	"gopkg.in/guregu/null.v4/zero"
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
		fmt.Println(err)
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
		fmt.Println(err)
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
	if err := models.DB.Table(table.Zones).Preload("Name").Preload("ParentZone").Find(&zones).Error; err != nil {
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
