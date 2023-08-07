// SPDX-License-Identifier: GPL-2.0-or-later
package controllers

import (
	"net/http"
	"time"

	"git.sr.ht/~klahr/300sf/models"
	"git.sr.ht/~klahr/300sf/utils"
	"github.com/gin-gonic/gin"
)

func CreateNamedRecord[T models.Age | models.Order | models.Gender | models.Activity | models.Method](c *gin.Context) {
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

	tx := models.DB.Begin()

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

	t := T{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateShortUuid(),
		},
		NameId: name.Id,
		Name:   name,
	}

	if err := tx.Create(&t).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, t)
}

func NamedRecordList[T models.Age | models.Order | models.Gender | models.Activity | models.Method | models.Zone](c *gin.Context) {
	var result []T

	if err := models.DB.Preload("Name").Find(&result).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	type response struct {
		Result []T `json:"result"`
	}

	c.JSON(http.StatusOK, response{
		Result: result,
	})
}

func NamedRecordDetails[T models.Age | models.Order | models.Gender | models.Activity | models.Method | models.Zone](c *gin.Context) {
	var result T

	if err := models.DB.Where("uuid=?", c.Param("uuid")).Preload("Name").First(&result).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	type response struct {
		Result T `json:"result"`
	}

	c.JSON(http.StatusOK, response{
		Result: result,
	})
}
