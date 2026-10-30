// SPDX-License-Identifier: AGPL-3.0-or-later
package controllers

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/table"
	"github.com/klahr/300sf/server/utils"
	"gopkg.in/guregu/null.v4/zero"
)

var tagColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func Tags(c *gin.Context) {
	type request struct {
		Name  string      `json:"name"`
		Color zero.String `json:"color"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Code: ErrorParse, Message: "could not parse request"})
		return
	}
	r.Name = strings.TrimSpace(r.Name)
	if r.Name == "" {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Code: ErrorInvalidRequest, Message: "name required"})
		return
	}
	if r.Color.String != "" && !tagColor.MatchString(r.Color.String) {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Code: ErrorInvalidRequest, Message: "color is #rrggbb"})
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

	tag := models.Tag{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateShortUuid(),
		},
		UserId: user.Id,
		Name:   r.Name,
	}

	if r.Color.Valid {
		tag.Color = zero.StringFrom(r.Color.String)
	}

	if err := tx.Create(&tag).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, tag)
}

func TagsArchive(c *gin.Context) {
	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := models.DB.Table(table.Tags).Where("user_id=?", user.Id).Where("uuid=?", c.Param("uuid")).Where("archived IS NULL").UpdateColumn("archived", time.Now()).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func TagsList(c *gin.Context) {
	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var tags []models.Tag
	if err := models.DB.Table(table.Tags).Where("user_id=?", user.Id).Where("archived IS NULL").Find(&tags).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	type response struct {
		Result []models.Tag `json:"result"`
	}

	c.JSON(http.StatusOK, response{
		Result: tags,
	})
}

func TagsRename(c *gin.Context) {
	type request struct {
		Name string `json:"name"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Code: ErrorParse, Message: "could not parse request"})
		return
	}
	r.Name = strings.TrimSpace(r.Name)
	if r.Name == "" {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Code: ErrorInvalidRequest, Message: "name required"})
		return
	}

	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var tag models.Tag
	if err := models.DB.Table(table.Tags).Where("user_id=?", user.Id).Where("uuid=?", c.Param("uuid")).Where("archived IS NULL").First(&tag).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := models.DB.Table(table.Tags).Where("id=?", tag.Id).UpdateColumn("name", r.Name).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func TagsDetails(c *gin.Context) {
	var result models.Tag

	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := models.DB.Where("uuid=?", c.Param("uuid")).Where("user_id=?", user.Id).First(&result).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	type response struct {
		Result models.Tag `json:"result"`
	}

	c.JSON(http.StatusOK, response{
		Result: result,
	})
}
