// SPDX-License-Identifier: AGPL-3.0-or-later
package controllers

import (
	"encoding/json"
	"errors"
	"mime"
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
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// reportTime is a sighting time: RFC 3339, or a bare "YYYY-MM-DD", which is
// stored as midnight UTC, the API's "date only".
type reportTime struct{ zero.Time }

func (t *reportTime) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil && len(s) == len("2006-01-02") {
		d, err := time.Parse("2006-01-02", s)
		if err != nil {
			return err
		}
		t.Time = zero.TimeFrom(d)
		return nil
	}
	return t.Time.UnmarshalJSON(b)
}

// reportFields are what a sighting is made of, for both reporting and editing one.
type reportFields struct {
	SpeciesUuid  string      `json:"species_uuid"`
	From         reportTime  `json:"from"`
	Until        reportTime  `json:"until"`
	Latitude     zero.Float  `json:"latitude"`
	Longitude    zero.Float  `json:"longitude"`
	Quantity     zero.Int    `json:"quantity"`
	AgeUuid      zero.String `json:"age_uuid"`
	ActivityUuid zero.String `json:"activity_uuid"`
	GenderUuid   zero.String `json:"gender_uuid"`
	MethodUuid   zero.String `json:"method_uuid"`
	Comment      zero.String `json:"comment"`
	Site         zero.String `json:"site"`
	ListUuids    []string    `json:"list_uuids"`
	TagUuids     []string    `json:"tag_uuids"`
}

// lookupId returns the id of the record of type T with this uuid, or null for none.
func lookupId[T any](tx *gorm.DB, uuid zero.String) (zero.Int, error) {
	if !uuid.Valid {
		return zero.Int{}, nil
	}
	var row struct{ Id int64 }
	if err := tx.Model(new(T)).Select("id").Where("uuid=?", uuid.String).Take(&row).Error; err != nil {
		return zero.Int{}, err
	}
	return zero.IntFrom(row.Id), nil
}

// setReportFields fills in report from r, all of it: what r leaves out is
// cleared. It reports whether the place has to be looked up once saved.
func setReportFields(tx *gorm.DB, user models.User, report *models.Report, r reportFields) (bool, error) {
	var species models.Species
	if err := tx.Where("uuid=?", r.SpeciesUuid).First(&species).Error; err != nil {
		return false, errors.New("species not found")
	}
	switch {
	case r.Quantity.Int64 < 0:
		return false, errors.New("quantity must not be negative")
	case r.Latitude.Float64 < -90 || r.Latitude.Float64 > 90 || r.Longitude.Float64 < -180 || r.Longitude.Float64 > 180:
		return false, errors.New("position out of range")
	case r.From.Valid && r.Until.Valid && r.Until.Time.Time.Before(r.From.Time.Time):
		return false, errors.New("until is before from")
	}
	report.SpeciesId = species.Id
	report.From = r.From.Time
	report.Until = r.Until.Time
	report.Latitude = r.Latitude
	report.Longitude = r.Longitude
	report.Quantity = r.Quantity
	report.Site = r.Site

	// Known places get their zone now; a new place is looked up once the
	// sighting is saved (places.Locate).
	locateLater := false
	report.ZoneId = zero.Int{}
	report.Zone = nil
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

	var err error
	if report.AgeId, err = lookupId[models.Age](tx, r.AgeUuid); err != nil {
		return false, err
	}
	if report.ActivityId, err = lookupId[models.Activity](tx, r.ActivityUuid); err != nil {
		return false, err
	}
	if report.GenderId, err = lookupId[models.Gender](tx, r.GenderUuid); err != nil {
		return false, err
	}
	if report.MethodId, err = lookupId[models.Method](tx, r.MethodUuid); err != nil {
		return false, err
	}

	report.Lists = []models.List{}
	if len(r.ListUuids) > 0 {
		if err := tx.Where("uuid IN ?", r.ListUuids).Where("user_id=?", user.Id).Find(&report.Lists).Error; err != nil {
			return false, err
		}
	}

	report.Tags = []models.Tag{}
	if len(r.TagUuids) > 0 {
		if err := tx.Where("uuid IN ?", r.TagUuids).Where("user_id=?", user.Id).Where("archived IS NULL").Find(&report.Tags).Error; err != nil {
			return false, err
		}
	}

	return locateLater, nil
}

// loadReport reads a sighting with everything a client shows of it.
func loadReport(db *gorm.DB, id int64) (models.Report, error) {
	var report models.Report
	err := db.Where("id=?", id).Preload("Zone", withoutPolygon).Preload("Zone.Name").Preload("Species.Name").Preload("Species.Order.Name").Preload("Age.Name").Preload("Gender.Name").Preload("Method.Name").Preload("Activity.Name").Preload("Lists.Name").Preload("Tags", "archived IS NULL").Preload("Comment").Preload("Notification.Comment").Preload("Notification.Comments").First(&report).Error
	return report, err
}

func Reports(c *gin.Context) {
	type request struct {
		reportFields
		// Optional, from apps that report offline: the same sighting sent again is stored once.
		Uuid                string      `json:"uuid"`
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
			stored, err := loadReport(tx, existing.Id)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, stored)
			return
		}
		uuid = r.Uuid
	}

	report := models.Report{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      uuid,
		},
		UserId: user.Id,
		Comment: &models.Comment{
			Model: models.Model{
				Timestamp: time.Now(),
				Uuid:      utils.GenerateShortUuid(),
			},
			UserId: user.Id,
			Text:   r.Comment.String,
		},
	}

	locateLater, err := setReportFields(tx, user, &report, r.reportFields)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

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

	stored, err := loadReport(models.DB, report.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, stored)
}

// ReportsUpdate edits a sighting: every field is replaced, so what the
// request leaves out is cleared.
func ReportsUpdate(c *gin.Context) {
	var r reportFields
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
		return
	}

	var report models.Report
	if err := tx.Where("uuid=?", c.Param("uuid")).Where("user_id=?", user.Id).Where("archived IS NULL").First(&report).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorRecordNotFound,
			"message": "report not found",
		})
		return
	}

	locateLater, err := setReportFields(tx, user, &report, r)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	// Select writes the cleared (null) fields too.
	if err := tx.Model(&report).Select("species_id", "from", "until", "latitude", "longitude", "zone_id", "quantity", "age_id", "activity_id", "gender_id", "method_id", "site").Omit(clause.Associations).Updates(&report).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	if err := tx.Model(&report).Association("Lists").Replace(report.Lists); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	if err := tx.Model(&report).Association("Tags").Replace(report.Tags); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if report.CommentId.Valid {
		err = tx.Model(&models.Comment{}).Where("id=?", report.CommentId.Int64).Update("text", r.Comment.String).Error
	} else {
		comment := models.Comment{
			Model: models.Model{
				Timestamp: time.Now(),
				Uuid:      utils.GenerateShortUuid(),
			},
			UserId: user.Id,
			Text:   r.Comment.String,
		}
		if err = tx.Create(&comment).Error; err == nil {
			err = tx.Model(&report).UpdateColumn("comment_id", comment.Id).Error
		}
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	if locateLater && !Throttled("geocode/"+c.GetString("uuid"), 100, time.Hour) {
		places.Locate(report.Id)
	}

	stored, err := loadReport(models.DB, report.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, stored)
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
	if err := models.DB.Table(table.Reports).Where("user_id=?", user.Id).Where("archived IS NULL").Preload("Tags", "archived IS NULL").Preload("Zone", withoutPolygon).Preload("Zone.Name").Preload("Lists").Preload("Species.Name").Preload("Age.Name").Preload("Gender.Name").Preload("Activity.Name").Preload("Method.Name").Preload("Lists.Name").Preload("Comment").Order("timestamp DESC").Find(&reports).Error; err != nil {
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
	if err := models.DB.Table(table.ListReports).Select("reports.*").Where("reports.archived IS NULL").Where("list_reports.list_id=?", list.Id).Joins("JOIN reports ON list_reports.report_id=reports.id").Preload("Species.Name").Preload("Zone", withoutPolygon).Preload("Zone.Name").Find(&reports).Error; err != nil {
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

	if err := models.DB.Where("uuid=?", c.Param("uuid")).Where("user_id=?", user.Id).Preload("Zone", withoutPolygon).Preload("Zone.Name").Preload("Species.Name").Preload("Age.Name").Preload("Gender.Name").Preload("Method.Name").Preload("Activity.Name").Preload("Lists.Name").Preload("Tags", "archived IS NULL").Preload("Comment").Preload("Notification.Comment").Preload("Notification.Comments").First(&result).Error; err != nil {
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
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Code: ErrorParse, Message: "could not parse request"})
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
	// An archived tag would be stored but never shown.
	if err := models.DB.Where("uuid=?", r.TagUuid).Where("user_id=?", user.Id).Where("archived IS NULL").First(&tag).Error; err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Code: ErrorRecordNotFound, Message: "tag not found"})
		return
	}
	for _, t := range report.Tags {
		if t.Id == tag.Id {
			c.JSON(http.StatusBadRequest, GeneralErrorResponse{Code: ErrorInsert, Message: "already on the sighting"})
			return
		}
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
	if err := models.DB.Where("uuid=?", r.ListUuid).Where("user_id=?", user.Id).Preload("Name").First(&list).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	// Oldest first, so the row kept for each species is the sighting that ticked it off the list.
	var reports []models.Report
	if err := models.DB.Table("list_reports").Select("reports.*").Preload("Comment").Preload("Zone", withoutPolygon).Preload("Zone.Name").Preload("Method.Name").Preload("Gender.Name").Preload("Activity.Name").Preload("Species.Name").Preload("Age.Name").Preload("Tags").Where("list_id=?", list.Id).Where("reports.archived IS NULL").Joins("JOIN reports ON reports.id=report_id").Order("reports.`from` IS NULL, reports.`from` ASC").Find(&reports).Error; err != nil {
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

	// Named after the list; FormatMediaType encodes non-ASCII names (RFC 2231).
	filename := list.Uuid
	names := map[string]zero.String{"sv": list.Name.Sv, "en": list.Name.En, "de": list.Name.De}
	for _, n := range []zero.String{names[user.Language.Code], list.Name.Sv, list.Name.En, list.Name.De} {
		if n.String != "" {
			filename = n.String
			break
		}
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filename + ".csv"}))
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

// ReportsAddList puts a sighting on one of the user's checklists, without
// resending the whole sighting.
func ReportsAddList(c *gin.Context) {
	report, list, ok := reportAndList(c)
	if !ok {
		return
	}
	if err := models.DB.Table(table.ListReports).Create(&models.ListReport{ListId: list.Id, ReportId: report.Id}).Error; err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Code: ErrorInsert, Message: "already on the list"})
		return
	}
	c.Status(http.StatusNoContent)
}

// ReportsRemoveList takes a sighting off one of the user's checklists.
func ReportsRemoveList(c *gin.Context) {
	report, list, ok := reportAndList(c)
	if !ok {
		return
	}
	if err := models.DB.Where("report_id=? AND list_id=?", report.Id, list.Id).Delete(&models.ListReport{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, GeneralErrorResponse{Code: ErrorDelete, Message: err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// reportAndList finds the user's non-archived sighting in the path and their
// checklist in the body, or answers 400.
func reportAndList(c *gin.Context) (models.Report, models.List, bool) {
	var r struct {
		ListUuid string `json:"list_uuid"`
	}
	var report models.Report
	var list models.List
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Code: ErrorParse, Message: err.Error()})
		return report, list, false
	}
	user, err := currentUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Code: ErrorRecordNotFound, Message: err.Error()})
		return report, list, false
	}
	if err := models.DB.Where("uuid=? AND user_id=? AND archived IS NULL", c.Param("uuid"), user.Id).First(&report).Error; err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Code: ErrorRecordNotFound, Message: "sighting not found"})
		return report, list, false
	}
	if err := models.DB.Where("uuid=? AND user_id=?", r.ListUuid, user.Id).First(&list).Error; err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Code: ErrorRecordNotFound, Message: "list not found"})
		return report, list, false
	}
	return report, list, true
}
