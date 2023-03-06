package controllers

import (
	"net/http"
	"time"

	"git.sr.ht/~klahr/300sf/models"
	"git.sr.ht/~klahr/300sf/table"
	"git.sr.ht/~klahr/300sf/utils"
	"github.com/gin-gonic/gin"
)

func Gender(c *gin.Context) {
	type request struct {
		Name models.Translation `json:"name"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
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

	gender := models.Gender{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateUuid(),
		},
		NameId: name.Id,
		Name:   name,
	}

	if err := models.DB.Create(&gender).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gender)
}

func GenderList(c *gin.Context) {
	var genders []models.Gender
	if err := models.DB.Table(table.Genders).Preload("Name").Find(&genders).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	type response struct {
		Genders []models.Gender `json:"genders"`
	}

	c.JSON(http.StatusOK, response{
		Genders: genders,
	})
}
