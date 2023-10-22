// SPDX-License-Identifier: GPL-2.0-or-later
package controllers

import (
	"net/http"
	"strings"
	"time"

	"git.sr.ht/~klahr/300sf/models"
	"git.sr.ht/~klahr/300sf/table"
	"git.sr.ht/~klahr/300sf/types"
	"git.sr.ht/~klahr/300sf/utils"
	"github.com/gin-gonic/gin"
	"gopkg.in/guregu/null.v4/zero"
)

func Reports(c *gin.Context) {
	type request struct {
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

	var user models.User
	if err := tx.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		tx.Rollback()
		return
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
			Uuid:      utils.GenerateShortUuid(),
		},
		UserId:    user.Id,
		SpeciesId: species.Id,
		From:      r.From,
		Until:     r.Until,
		Latitude:  r.Latitude,
		Longitude: r.Longitude,
		Quantity:  r.Quantity,
		Comment: &models.Comment{
			Model: models.Model{
				Timestamp: time.Now(),
				Uuid:      utils.GenerateShortUuid(),
			},
			UserId: user.Id,
			Text:   r.Comment.String,
		},
	}

	if zones := utils.GetZonesFromPosition(types.LatLon{
		Latitude:  r.Latitude.Float64,
		Longitude: r.Longitude.Float64,
	}); len(zones) > 0 {
		report.ZoneId = zero.IntFrom(zones[len(zones)-1].Id)
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

	for i := range r.ListUuids {
		r.ListUuids[i] = "'" + r.ListUuids[i] + "'"
	}

	var lists []models.List
	if len(r.ListUuids) > 0 {
		if err := tx.Where("uuid IN (" + strings.Join(r.ListUuids, ",") + ")").Find(&lists).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    -10001,
				"message": err.Error(),
			})
			tx.Rollback()
			return
		}
	}

	report.Lists = lists

	for i := range r.TagUuids {
		r.TagUuids[i] = "'" + r.TagUuids[i] + "'"
	}

	var tags []models.Tag
	if len(r.TagUuids) > 0 {
		if err := tx.Where("uuid IN (" + strings.Join(r.TagUuids, ",") + ")").Where("archived IS NULL").Find(&tags).Error; err != nil {
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
	if err := models.DB.Table(table.Reports).Where("user_id=?", user.Id).Where("archived IS NULL").Preload("Tags").Preload("Zone.Name").Preload("Lists").Preload("Species.Name").Preload("Age.Name").Preload("Gender.Name").Preload("Activity.Name").Preload("Method.Name").Preload("Lists.Name").Preload("Comment").Order("timestamp DESC").Find(&reports).Error; err != nil {
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
	if err := models.DB.Where("uuid=?", r.ListUuid).First(&list).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var reports []models.Report
	if err := models.DB.Table(table.ListReports).Select("reports.*").Where("archived IS NULL").Where("list_reports.list_id=?", list.Id).Joins("JOIN reports ON list_reports.report_id=reports.id").Preload("Species.Name").Preload("Zone.Name").Find(&reports).Error; err != nil {
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
	if err := models.DB.Where("uuid=?", c.Param("uuid")).Where("user_id=?", user.Id).Preload("Zone.Name").Preload("Species.Name").Preload("Age.Name").Preload("Method.Name").Preload("Activity.Name").Preload("Lists.Name").Preload("Tags").Preload("Comment").Preload("Notification.Comment").Preload("Notification.Comments.User").First(&result).Error; err != nil {
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
	if err := models.DB.Where("user_id=?", user.Id).Where("uuid=?", c.Param("uuid")).Preload("Tags").First(&report).Error; err != nil {
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
	if err := models.DB.Where("user_id=?", user.Id).Where("uuid=?", c.Param("uuid")).Preload("Tags").First(&report).Error; err != nil {
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

func ReportsBasis(c *gin.Context) {
	type response struct {
		Species    []models.Species  `json:"species"`
		Ages       []models.Age      `json:"ages"`
		Genders    []models.Gender   `json:"genders"`
		Methods    []models.Method   `json:"methods"`
		Activities []models.Activity `json:"activities"`
		Lists      []models.List     `json:"lists"`
		Tags       []models.Tag      `json:"tags"`
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

	if err := models.DB.Preload("Name").Find(&resp.Species).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
	}

	if err := models.DB.Preload("Name").Find(&resp.Ages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
	}

	if err := models.DB.Preload("Name").Find(&resp.Genders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
	}

	if err := models.DB.Preload("Name").Find(&resp.Methods).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
	}

	if err := models.DB.Preload("Name").Find(&resp.Activities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
	}

	if err := models.DB.Preload("Name").Where("user_id=?", user.Id).Find(&resp.Lists).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
	}

	if err := models.DB.Where("user_id=?", user.Id).Find(&resp.Tags).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
	}

	c.JSON(http.StatusOK, resp)
}
