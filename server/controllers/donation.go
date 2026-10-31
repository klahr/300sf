// SPDX-License-Identifier: AGPL-3.0-or-later
package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/server/log"
	"github.com/klahr/300sf/server/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// The donation page's progress: how much is raised of the goal, in whole SEK.
// Admins set both (settings donation_raised and donation_goal).

type Donation struct {
	Raised int64 `json:"raised"`
	Goal   int64 `json:"goal"`
}

type setting struct {
	Name  string `gorm:"primaryKey"`
	Value string
}

func (setting) TableName() string {
	return "settings"
}

func loadDonation(db *gorm.DB) (Donation, error) {
	var rows []setting
	if err := db.Where("name IN ?", []string{"donation_raised", "donation_goal"}).Find(&rows).Error; err != nil {
		return Donation{}, err
	}
	var d Donation
	for _, r := range rows {
		v, _ := strconv.ParseInt(r.Value, 10, 64)
		switch r.Name {
		case "donation_raised":
			d.Raised = v
		case "donation_goal":
			d.Goal = v
		}
	}
	return d, nil
}

// DonationGet is the donation progress, for anyone.
func DonationGet(c *gin.Context) {
	d, err := loadDonation(models.DB)
	if err != nil {
		log.Error(c, ErrorInternalServerError, err.Error())
		c.JSON(http.StatusInternalServerError, GeneralErrorResponse{Code: ErrorInternalServerError, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, d)
}

// DonationSet sets the donation progress.
func DonationSet(c *gin.Context) {
	var r struct {
		Raised *int64 `json:"raised" binding:"required"`
		Goal   *int64 `json:"goal" binding:"required"`
	}
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Code: ErrorParse, Message: err.Error()})
		return
	}
	if *r.Raised < 0 || *r.Goal <= 0 {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Code: ErrorInvalidRequest, Message: "raised must be 0 or more and goal more than 0"})
		return
	}
	rows := []setting{
		{Name: "donation_raised", Value: strconv.FormatInt(*r.Raised, 10)},
		{Name: "donation_goal", Value: strconv.FormatInt(*r.Goal, 10)},
	}
	if err := models.DB.Clauses(clause.OnConflict{UpdateAll: true}).Create(&rows).Error; err != nil {
		log.Error(c, ErrorInternalServerError, err.Error())
		c.JSON(http.StatusInternalServerError, GeneralErrorResponse{Code: ErrorInternalServerError, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, Donation{Raised: *r.Raised, Goal: *r.Goal})
}
