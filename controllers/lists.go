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

	name := r.Name
	name.Timestamp = time.Now()
	name.Uuid = utils.GenerateUuid()
	if err := models.DB.Create(&name).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	list := models.List{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateUuid(),
		},
		UserId:     user.Id,
		Visibility: r.Visibility,
		Name:       name,
	}

	if r.Template.Valid {
		var template models.ListTemplate
		if err := models.DB.Where("uuid=?", r.Template).Preload("Species").First(&template).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}
		list.Species = template.Species
	}

	if err := models.DB.Create(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, list)
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
	if err := models.DB.Table(table.Lists).Where("user_id=?", user.Id).Where("archived IS NULL").Preload("Name").Preload("Species.Name").Find(&lists).Error; err != nil {
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
	c.JSON(http.StatusTeapot, nil)
}
