package controllers

import (
	"net/http"
	"strings"
	"time"

	"git.sr.ht/~klahr/300sf/models"
	"git.sr.ht/~klahr/300sf/table"
	"git.sr.ht/~klahr/300sf/utils"
	"github.com/gin-gonic/gin"
	"gopkg.in/guregu/null.v4/zero"
)

func Reports(c *gin.Context) {
	type request struct {
		SpeciesUuid  string      `json:"species_uuid"`
		From         zero.Time   `json:"from"`
		Until        zero.Time   `json:"until"`
		Latitude     zero.Float  `json:"latitude"`
		Longitude    zero.Float  `json:"longitude"`
		Quantity     zero.Int    `json:"quantity"`
		AgeUuid      zero.String `json:"age_uuid"`
		ActivityUuid zero.String `json:"activity_uuid"`
		GenderUuid   zero.String `json:"gender_uuid"`
		MethodUuid   zero.String `json:"method_uuid"`
		Comment      zero.String `json:"comment"`
		ListUuids    []string    `json:"list_uuids"`
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

	var species models.Species
	if err := models.DB.Where("uuid=?", r.SpeciesUuid).First(&species).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	report := models.Report{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateUuid(),
		},
		UserId:    user.Id,
		SpeciesId: species.Id,
		From:      r.From,
		Until:     r.Until,
		Latitude:  r.Latitude,
		Longitude: r.Longitude,
		Quantity:  r.Quantity,
		Comment:   r.Comment,
	}

	if r.AgeUuid.Valid {
		var age models.Age
		if err := models.DB.Where("uuid=?", r.AgeUuid.String).First(&age).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    -10001,
				"message": err.Error(),
			})
			return
		}
		report.AgeId = zero.IntFrom(age.Id)
	}

	if r.ActivityUuid.Valid {
		var activity models.Activity
		if err := models.DB.Where("uuid=?", r.ActivityUuid.String).First(&activity).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    -10001,
				"message": err.Error(),
			})
			return
		}
		report.ActivityId = zero.IntFrom(activity.Id)
	}

	if r.GenderUuid.Valid {
		var gender models.Gender
		if err := models.DB.Where("uuid=?", r.GenderUuid.String).First(&gender).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    -10001,
				"message": err.Error(),
			})
			return
		}
		report.GenderId = zero.IntFrom(gender.Id)
	}

	if r.MethodUuid.Valid {
		var method models.Method
		if err := models.DB.Where("uuid=?", r.MethodUuid.String).First(&method).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    -10001,
				"message": err.Error(),
			})
			return
		}
		report.MethodId = zero.IntFrom(method.Id)
	}

	var lists []models.List
	if len(r.ListUuids) > 0 {
		if err := models.DB.Where("uuid IN (" + strings.Join(r.ListUuids, ",") + ")").Find(&lists).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    -10001,
				"message": err.Error(),
			})
			return
		}
	}

	report.Lists = lists

	if err := models.DB.Create(&report).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
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
	if err := models.DB.Table(table.Reports).Where("user_id=?", user.Id).Preload("Lists").Find(&reports).Error; err != nil {
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
