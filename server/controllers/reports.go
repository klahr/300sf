// SPDX-License-Identifier: AGPL-3.0-or-later
package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/places"
	"github.com/klahr/300sf/server/table"
	"github.com/klahr/300sf/server/types"
	"github.com/klahr/300sf/server/utils"
	"gopkg.in/guregu/null.v4/zero"
)

func Reports(c *gin.Context) {
	type request struct {
		// Optional, from apps that report offline: the same sighting sent again is stored once.
		Uuid                string      `json:"uuid"`
		SpeciesUuid         string      `json:"species_uuid"`
		From                zero.Time   `json:"from" time_format:"2006-01-02T15:04Z"`
		Until               zero.Time   `json:"until" time_format:"2006-01-02T15:04Z"`
		Latitude            zero.Float  `json:"latitude"`
		Longitude           zero.Float  `json:"longitude"`
		Quantity            zero.Int    `json:"quantity"`
		AgeUuid             zero.String `json:"age_uuid"`
		ActivityUuid        zero.String `json:"activity_uuid"`
		GenderUuid          zero.String `json:"gender_uuid"`
		MethodUuid          zero.String `json:"method_uuid"`
		Comment             zero.String `json:"comment"`
		Site                zero.String `json:"site"`
		ListUuids           []string    `json:"list_uuids"`
		TagUuids            []string    `json:"tag_uuids"`
		Silent              bool        `json:"silent"`
		NotificationComment zero.String `json:"notification_comment"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
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

	uuid := utils.GenerateShortUuid()
	if r.Uuid != "" {
		if !utils.IsClientUuid(r.Uuid) {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    -10001,
				"message": "invalid uuid",
			})
			return
		}
		// Sent before, but the answer was lost on the way: it is already stored.
		var existing models.Report
		if err := tx.Where("uuid=?", r.Uuid).Limit(1).Find(&existing).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}
		if existing.Id != 0 {
			if existing.UserId != user.Id {
				c.JSON(http.StatusConflict, gin.H{
					"code":    -10001,
					"message": "uuid already in use",
				})
				return
			}
			c.JSON(http.StatusOK, existing)
			return
		}
		uuid = r.Uuid
	}

	var species models.Species
	if err := tx.Where("uuid=?", r.SpeciesUuid).First(&species).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": "species not found",
		})
		tx.Rollback()
		return
	}

	report := models.Report{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      uuid,
		},
		UserId:    user.Id,
		SpeciesId: species.Id,
		From:      r.From,
		Until:     r.Until,
		Latitude:  r.Latitude,
		Longitude: r.Longitude,
		Quantity:  r.Quantity,
		Site:      r.Site,
		Comment: &models.Comment{
			Model: models.Model{
				Timestamp: time.Now(),
				Uuid:      utils.GenerateShortUuid(),
			},
			UserId: user.Id,
			Text:   r.Comment.String,
		},
	}

	// Known places get their zone now; a new place is looked up once the
	// sighting is saved (places.Locate).
	locateLater := false
	if r.Latitude.Valid && r.Longitude.Valid {
		zones := utils.GetZonesFromPosition(types.LatLon{
			Latitude:  r.Latitude.Float64,
			Longitude: r.Longitude.Float64,
		})
		if len(zones) > 0 {
			report.ZoneId = zero.IntFrom(zones[len(zones)-1].Id)
		}
		locateLater = !places.Complete(zones)
	}

	if r.AgeUuid.Valid {
		var age models.Age
		if err := tx.Where("uuid=?", r.AgeUuid.String).First(&age).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    -10001,
				"message": err.Error(),
			})
			tx.Rollback()
			return
		}
		report.AgeId = zero.IntFrom(age.Id)
	}

	if r.ActivityUuid.Valid {
		var activity models.Activity
		if err := tx.Where("uuid=?", r.ActivityUuid.String).First(&activity).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    -10001,
				"message": err.Error(),
			})
			tx.Rollback()
			return
		}
		report.ActivityId = zero.IntFrom(activity.Id)
	}

	if r.GenderUuid.Valid {
		var gender models.Gender
		if err := tx.Where("uuid=?", r.GenderUuid.String).First(&gender).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    -10001,
				"message": err.Error(),
			})
			tx.Rollback()
			return
		}
		report.GenderId = zero.IntFrom(gender.Id)
	}

	if r.MethodUuid.Valid {
		var method models.Method
		if err := tx.Where("uuid=?", r.MethodUuid.String).First(&method).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    -10001,
				"message": err.Error(),
			})
			tx.Rollback()
			return
		}
		report.MethodId = zero.IntFrom(method.Id)
	}

	var lists []models.List
	if len(r.ListUuids) > 0 {
		if err := tx.Where("uuid IN ?", r.ListUuids).Where("user_id=?", user.Id).Find(&lists).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    -10001,
				"message": err.Error(),
			})
			tx.Rollback()
			return
		}
	}

	report.Lists = lists

	var tags []models.Tag
	if len(r.TagUuids) > 0 {
		if err := tx.Where("uuid IN ?", r.TagUuids).Where("user_id=?", user.Id).Where("archived IS NULL").Find(&tags).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    -10001,
				"message": err.Error(),
			})
			tx.Rollback()
			return
		}
	}

	report.Tags = tags

	if err := tx.Create(&report).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	if !r.Silent {
		notification := models.Notification{
			Model: models.Model{
				Timestamp: time.Now(),
				Uuid:      utils.GenerateShortUuid(),
			},
			ReportId: report.Id,
			Comment: models.Comment{
				Model: models.Model{
					Timestamp: time.Now(),
					Uuid:      utils.GenerateShortUuid(),
				},
				UserId: user.Id,
				Text:   r.NotificationComment.String,
			},
		}

		if err := tx.Create(&notification).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			tx.Rollback()
			return
		}
	}

	tx.Commit()
	// The geocoder takes one place a second for the whole site, so one user
	// can't fill its queue; places not looked up now are
	// left to the backfill.
	if locateLater && !Throttled("geocode/"+c.GetString("uuid"), 100, time.Hour) {
		places.Locate(report.Id)
	}

	c.JSON(http.StatusOK, report)
}

func ReportsList(c *gin.Context) {
	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var reports []models.Report
	if err := models.DB.Table(table.Reports).Where("user_id=?", user.Id).Where("archived IS NULL").Preload("Tags", "archived IS NULL").Preload("Zone.Name").Preload("Lists").Preload("Species.Name").Preload("Age.Name").Preload("Gender.Name").Preload("Activity.Name").Preload("Method.Name").Preload("Lists.Name").Preload("Comment").Order("timestamp DESC").Find(&reports).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	type response struct {
		Result []models.Report `json:"result"`
	}

	c.JSON(http.StatusOK, response{
		Result: reports,
	})
}

func ReportsListByList(c *gin.Context) {
	type request struct {
		ListUuid string `json:"list_uuid"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var list models.List
	if err := models.DB.Where("uuid=?", r.ListUuid).Where("user_id=?", user.Id).First(&list).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var reports []models.Report
	if err := models.DB.Table(table.ListReports).Select("reports.*").Where("reports.archived IS NULL").Where("list_reports.list_id=?", list.Id).Joins("JOIN reports ON list_reports.report_id=reports.id").Preload("Species.Name").Preload("Zone.Name").Find(&reports).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	type response struct {
		Result []models.Report `json:"result"`
	}

	c.JSON(http.StatusOK, response{
		Result: reports,
	})
}

// ReportsDelete removes a sighting for good with its comment; its links to lists, labels and notifications go with it.
func ReportsDelete(c *gin.Context) {
	tx := models.DB.Begin()
	defer tx.Rollback()

	var user models.User
	if err := tx.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var report models.Report
	if err := tx.Where("user_id=?", user.Id).Where("uuid=?", c.Param("uuid")).First(&report).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := tx.Delete(&report).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if report.CommentId.Valid {
		if err := tx.Where("id=?", report.CommentId.Int64).Delete(&models.Comment{}).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func ReportsArchive(c *gin.Context) {
	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := models.DB.Table(table.Reports).Where("user_id=?", user.Id).Where("uuid=?", c.Param("uuid")).Where("archived IS NULL").UpdateColumn("archived", time.Now()).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func ReportsDetails(c *gin.Context) {
	var result models.Report

	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	// TODO: Preload("Lists.Name") is not working and returns an empty list.
	if err := models.DB.Where("uuid=?", c.Param("uuid")).Where("user_id=?", user.Id).Preload("Zone.Name").Preload("Species.Name").Preload("Age.Name").Preload("Method.Name").Preload("Activity.Name").Preload("Lists.Name").Preload("Tags", "archived IS NULL").Preload("Comment").Preload("Notification.Comment").Preload("Notification.Comments").First(&result).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "report not found",
		})
		return
	}

	type response struct {
		Result models.Report `json:"result"`
	}

	c.JSON(http.StatusOK, response{
		Result: result,
	})
}

func ReportsAddTag(c *gin.Context) {
	type request struct {
		TagUuid string `json:"tag_uuid"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	var report models.Report
	if err := models.DB.Where("user_id=?", user.Id).Where("uuid=?", c.Param("uuid")).Preload("Tags", "archived IS NULL").First(&report).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var tag models.Tag
	if err := models.DB.Where("uuid=?", r.TagUuid).Where("user_id=?", user.Id).First(&tag).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := models.DB.Table("report_tags").Create(&models.ReportTag{
		ReportId: report.Id,
		TagId:    tag.Id,
	}).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func ReportsRemoveTag(c *gin.Context) {
	type request struct {
		TagUuid string `json:"tag_uuid"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	var report models.Report
	if err := models.DB.Where("user_id=?", user.Id).Where("uuid=?", c.Param("uuid")).Preload("Tags", "archived IS NULL").First(&report).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var tag models.Tag
	if err := models.DB.Where("uuid=?", r.TagUuid).Where("user_id=?", user.Id).First(&tag).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := models.DB.Where("report_id=?", report.Id).Where("tag_id=?", tag.Id).Delete(&models.ReportTag{}).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func ReportsExport(c *gin.Context) {
	type request struct {
		ListUuid       string `json:"list_uuid"`
		Format         string `json:"format"`
		TimezoneOffset int64  `json:"timezone_offset"`
		TimeZone       string `json:"time_zone"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).Preload("Language").First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	var list models.List
	if err := models.DB.Where("uuid=?", r.ListUuid).Where("user_id=?", user.Id).First(&list).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	// Oldest first, so the row kept for each species is the sighting that ticked it off the list.
	var reports []models.Report
	if err := models.DB.Table("list_reports").Select("reports.*").Preload("Comment").Preload("Zone.Name").Preload("Method.Name").Preload("Gender.Name").Preload("Activity.Name").Preload("Species.Name").Preload("Age.Name").Preload("Tags").Where("list_id=?", list.Id).Where("reports.archived IS NULL").Joins("JOIN reports ON reports.id=report_id").Order("reports.`from` IS NULL, reports.`from` ASC").Find(&reports).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	// Prefer an IANA zone so each date gets the offset (DST) of its own day;
	// timezone_offset is kept for older clients.
	localTime := func(t time.Time) time.Time {
		return t.Add(time.Duration(r.TimezoneOffset) * time.Minute).UTC()
	}
	if loc, err := time.LoadLocation(r.TimeZone); r.TimeZone != "" && err == nil {
		localTime = func(t time.Time) time.Time {
			return t.In(loc)
		}
	}

	lang := user.Language.Code
	name := func(t models.Translation) string {
		return csvText(translated(t, lang))
	}

	rows := [][]string{
		{"species", "scientific_name", "first_seen", "seen_until", "quantity", "method", "age", "sex", "activity", "comment", "site", "municipality", "latitude", "longitude", "labels"},
	}
	encountered := make(map[int64]bool)
	for _, report := range reports {
		if encountered[report.SpeciesId] {
			continue
		}
		encountered[report.SpeciesId] = true

		row := make([]string, len(rows[0]))
		row[0] = name(report.Species.Name)
		row[1] = csvText(report.Species.Name.La.String)
		if report.From.Valid {
			row[2] = csvTime(report.From.Time, localTime)
		}
		if report.Until.Valid {
			row[3] = csvTime(report.Until.Time, localTime)
		}
		if !report.Quantity.IsZero() {
			row[4] = strconv.FormatInt(report.Quantity.ValueOrZero(), 10)
		}
		if report.Method != nil {
			row[5] = name(report.Method.Name)
		}
		if report.Age != nil {
			row[6] = name(report.Age.Name)
		}
		if report.Gender != nil {
			row[7] = name(report.Gender.Name)
		}
		if report.Activity != nil {
			row[8] = name(report.Activity.Name)
		}
		if report.Comment != nil {
			row[9] = csvText(report.Comment.Text)
		}
		row[10] = csvText(report.Site.String)
		if report.Zone != nil {
			row[11] = name(report.Zone.Name)
		}
		if !report.Latitude.IsZero() {
			row[12] = strconv.FormatFloat(report.Latitude.Float64, 'f', -1, 64)
		}
		if !report.Longitude.IsZero() {
			row[13] = strconv.FormatFloat(report.Longitude.Float64, 'f', -1, 64)
		}
		labels := []string{}
		for _, t := range report.Tags {
			labels = append(labels, t.Name)
		}
		row[14] = csvText(strings.Join(labels, ", "))
		rows = append(rows, row)
	}

	data, err := writeCSV(rows, lang)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.String(http.StatusOK, data)
}

func ReportsBasis(c *gin.Context) {
	type request struct {
		SpeciesUuid string `json:"species_uuid"`
	}

	type response struct {
		Species    models.Species    `json:"species"`
		Ages       []models.Age      `json:"ages"`
		Genders    []models.Gender   `json:"genders"`
		Methods    []models.Method   `json:"methods"`
		Activities []models.Activity `json:"activities"`
		Lists      []models.List     `json:"lists"`
		Tags       []models.Tag      `json:"tags"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	var resp response

	if err := models.DB.Preload("Name").Where("uuid=?", r.SpeciesUuid).First(&resp.Species).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := models.DB.Preload("Name").Find(&resp.Ages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := models.DB.Preload("Name").Find(&resp.Genders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := models.DB.Preload("Name").Find(&resp.Methods).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := models.DB.Preload("Name").Find(&resp.Activities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := models.DB.Preload("Name").Where("user_id=?", user.Id).Where("archived IS NULL").Find(&resp.Lists).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := models.DB.Where("user_id=?", user.Id).Where("archived IS NULL").Find(&resp.Tags).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}
