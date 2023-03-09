package controllers

import (
	"net/http"
	"time"

	"git.sr.ht/~klahr/300sf/models"
	"git.sr.ht/~klahr/300sf/utils"
	"github.com/gin-gonic/gin"
)

func Species(c *gin.Context) {
	type request struct {
		Name      models.Translation `json:"name"`
		OrderUuid string             `json:"order_uuid"`
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

	var order models.Order
	if err := tx.Where("uuid=?", r.OrderUuid).First(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	species := models.Species{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateShortUuid(),
		},
		NameId:  name.Id,
		Name:    name,
		OrderId: order.Id,
	}

	if err := tx.Create(&species).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, species)
}

func SpeciesList(c *gin.Context) {
	var result []models.Species

	if err := models.DB.Preload("Name").Preload("Order.Name").Find(&result).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	type response struct {
		Result []models.Species `json:"result"`
	}

	c.JSON(http.StatusOK, response{
		Result: result,
	})
}
