// SPDX-License-Identifier: AGPL-3.0-or-later
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/table"
	"gopkg.in/guregu/null.v4/zero"
)

func TranslationList(c *gin.Context) {
	var translations []models.Translation
	if err := models.DB.Table(table.Translations).Find(&translations).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	type response struct {
		Translations []models.Translation `json:"translations"`
	}

	c.JSON(http.StatusOK, response{
		Translations: translations,
	})
}

func TranslationUpdate(c *gin.Context) {
	type request struct {
		Uuid string `json:"uuid"`
		models.Translation
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	var translation models.Translation
	if err := models.DB.Where("uuid=?", r.Uuid).First(&translation).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	} else {
		if r.En.Valid {
			translation.En = zero.StringFrom(r.En.String)
		}
		if r.Sv.Valid {
			translation.Sv = zero.StringFrom(r.Sv.String)
		}
		if r.De.Valid {
			translation.De = zero.StringFrom(r.De.String)
		}
		if r.La.Valid {
			translation.La = zero.StringFrom(r.La.String)
		}
		if err := models.DB.Where("id=?", translation.Model.Id).Updates(&translation).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusNoContent, nil)
}
