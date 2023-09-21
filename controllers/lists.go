// SPDX-License-Identifier: GPL-2.0-or-later
package controllers

import (
	"net/http"
	"time"

	"git.sr.ht/~klahr/300sf/models"
	"git.sr.ht/~klahr/300sf/table"
	"git.sr.ht/~klahr/300sf/utils"
	"github.com/gin-gonic/gin"
	"gopkg.in/guregu/null.v4/zero"
)

func Lists(c *gin.Context) {
	type request struct {
		Name       models.Translation `json:"name"`
		Visibility int64              `json:"visibility"`
		Template   zero.String        `json:"template"`
		From       zero.String        `json:"from"`
		Until      zero.String        `json:"until"`
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

	list, err := CreateList(r.Template, r.Name, r.Visibility, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if r.From.Valid {
		list.From = zero.TimeFrom(utils.Str2Time(r.From.String))
	}
	if r.Until.Valid {
		list.Until = zero.TimeFrom(utils.Str2Time(r.Until.String))
	}

	if err := models.DB.Create(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, list)
}

func CreateList(template zero.String, name models.Translation, visibility int64, user models.User) (*models.List, error) {
	name.Timestamp = time.Now()
	name.Uuid = utils.GenerateShortUuid()
	if err := models.DB.Create(&name).Error; err != nil {
		return nil, err
	}

	list := models.List{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateShortUuid(),
		},
		UserId:     user.Id,
		Visibility: visibility,
		Name:       name,
	}

	if template.Valid {
		var t models.ListTemplate
		if err := models.DB.Where("uuid=?", template.String).Preload("Species").First(&t).Error; err != nil {
			return nil, err
		}
		list.ListTemplateId = zero.IntFrom(t.Id)
		list.Species = t.Species
	}

	return &list, nil
}

func ListsList(c *gin.Context) {
	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var lists []models.List
	if err := models.DB.Table(table.Lists).Where("user_id=?", user.Id).Where("archived IS NULL").Preload("Name").Preload("Species.Name").Preload("Species.Order.Name").Find(&lists).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	type response struct {
		Result []models.List `json:"result"`
	}

	c.JSON(http.StatusOK, response{
		Result: lists,
	})
}

func ListsArchive(c *gin.Context) {
	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := models.DB.Table(table.Lists).Where("user_id=?", user.Id).Where("uuid=?", c.Param("uuid")).Where("archived IS NULL").UpdateColumn("archived", time.Now()).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func ListsRename(c *gin.Context) {
	type request struct {
		Name models.Translation `json:"name"`
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

	var list models.List
	if err := models.DB.Where("user_id=?", user.Id).Where("uuid=?", c.Param("uuid")).First(&list).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := models.DB.Where("id=?", list.NameId).Updates(&r.Name).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func ListsTimespan(c *gin.Context) {
	type request struct {
		From  zero.String `json:"from"`
		Until zero.String `json:"until"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
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
		return
	}

	var list models.List
	if err := tx.Where("user_id=?", user.Id).Where("uuid=?", c.Param("uuid")).First(&list).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if r.From.Valid {
		var err error
		if r.From.String == "*" {
			err = tx.Table(table.Lists).Where("id=?", list.Id).UpdateColumn("from", nil).Error
		} else {
			err = tx.Table(table.Lists).Where("id=?", list.Id).UpdateColumn("from", r.From.String).Error
		}
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}
	}

	if r.Until.Valid {
		var err error
		if r.Until.String == "*" {
			err = tx.Table(table.Lists).Where("id=?", list.Id).UpdateColumn("until", nil).Error
		} else {
			err = tx.Table(table.Lists).Where("id=?", list.Id).UpdateColumn("until", r.Until.String).Error
		}
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}
	}

	tx.Commit()

	c.JSON(http.StatusNoContent, nil)
}

func ListsDetails(c *gin.Context) {
	var result models.List

	if err := models.DB.Where("uuid=?", c.Param("uuid")).Preload("Name").Preload("Species.Name").Preload("Species.Order.Name").First(&result).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	type response struct {
		Result models.List `json:"result"`
	}

	c.JSON(http.StatusOK, response{
		Result: result,
	})
}

func ListsReports(c *gin.Context) {
	var result []models.Report

	tx := models.DB.Begin()

	var user models.User
	if err := tx.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	var list models.List
	if err := tx.Where("user_id=?", user.Id).Where("uuid=?", c.Param("uuid")).First(&list).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := tx.Table("list_reports").Select("reports.*").Where("list_id=?", list.Id).Joins("JOIN reports ON reports.id=report_id").Preload("Species.Name").Preload("Zone.Name").Preload("Age.Name").Preload("Gender.Name").Preload("Method.Name").Preload("Tags").Find(&result).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	tx.Commit()

	type response struct {
		Result []models.Report `json:"result"`
	}

	c.JSON(http.StatusOK, response{
		Result: result,
	})
}
