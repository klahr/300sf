package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/api/models"
	"github.com/klahr/300sf/api/table"
	"github.com/klahr/300sf/api/utils"
	"gopkg.in/guregu/null.v4/zero"
)

func Competitions(c *gin.Context) {
	type request struct {
		Name                  models.Translation `json:"name"`
		From                  time.Time          `json:"from"`
		Until                 time.Time          `json:"until"`
		ListTemplateUuid      string             `json:"list_template_uuid"`
		Description           models.Translation `json:"description"`
		EndOnReports          int64              `json:"end_on_reports"`
		MaxCompetitors        int64              `json:"max_competitors"`
		Visibility            int64              `json:"visibility"`
		JoinFrom              time.Time          `json:"join_from"`
		JoinUntil             time.Time          `json:"join_until"`
		Icon                  string             `json:"icon"`
		Color                 string             `json:"color"`
		CompetitionZonesUuids []string           `json:"competition_zones_uuids"`
		Password              string             `json:"password"`
		CreatorJoin           bool               `json:"creator_join"`
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
		return
	}

	name := r.Name
	name.Timestamp = time.Now()
	name.Uuid = utils.GenerateShortUuid()
	if err := tx.Create(&name).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	description := r.Description
	description.Timestamp = time.Now()
	description.Uuid = utils.GenerateShortUuid()
	if err := tx.Create(&description).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	var listTemplate models.ListTemplate
	if err := tx.Where("uuid=?", r.ListTemplateUuid).First(&listTemplate).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	competition := models.Competition{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateShortUuid(),
		},
		Name:           name,
		Description:    description,
		UserId:         zero.IntFrom(user.Id),
		From:           r.From,
		Until:          r.Until,
		ListTemplateId: zero.IntFrom(listTemplate.Id),
		Visibility:     r.Visibility,
		JoinFrom:       r.JoinFrom,
		JoinUntil:      r.JoinUntil,
		Icon:           r.Icon,
		Color:          r.Color,
	}

	if r.EndOnReports > 0 {
		competition.EndOnReports = zero.IntFrom(r.EndOnReports)
	}
	if r.MaxCompetitors > 0 {
		competition.MaxCompetitors = zero.IntFrom(r.MaxCompetitors)
	}
	if r.Password != "" {
		p := utils.Password{}
		p.Create(r.Password)
		competition.Password = zero.StringFrom(p.Value)
		competition.PasswordSalt = zero.StringFrom(p.Salt)
	}

	for _, z := range r.CompetitionZonesUuids {
		var zone models.Zone
		if err := tx.Where("uuid=?", z).First(&zone).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    -10001,
				"message": err.Error(),
			})
			return
		}
	}

	if r.CreatorJoin {
		if err := tx.Table("competition_users").Create(&models.CompetitionUsers{
			CompetitionId: competition.Id,
			UserId:        user.Id,
		}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    -10001,
				"message": "could not add creator",
			})
			return
		}
	}

	tx.Commit()

	c.JSON(http.StatusOK, competition)
}

func CompetitionsBasis(c *gin.Context) {
	type response struct {
		Zones         []models.Zone         `json:"zones"`
		ListTemplates []models.ListTemplate `json:"list_templates"`
	}

	var r response

	if err := models.DB.Preload("Name").Where("parent_zone_id IS NULL").Find(&r.Zones).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -10001,
			"message": "could not find zones",
		})
		return
	}

	var user models.User
	models.DB.Where("uuid=?", c.GetString("uuid")).First(&user)

	if err := models.DB.Table(table.ListTemplates).Where("user_id=? OR visibility=?", user.Id, models.VisibilityPublic).Where("archived IS NULL").Preload("Name").Preload("Description").Preload("Species").Find(&r.ListTemplates).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, r)
}
