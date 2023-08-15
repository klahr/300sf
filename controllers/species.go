// SPDX-License-Identifier: GPL-2.0-or-later
package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"git.sr.ht/~klahr/300sf/config"
	"git.sr.ht/~klahr/300sf/models"
	"git.sr.ht/~klahr/300sf/table"
	"git.sr.ht/~klahr/300sf/utils"
	"git.sr.ht/~klahr/300sf/wikipedia"
	"git.sr.ht/~klahr/300sf/xenocanto"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

func Species(c *gin.Context) {
	type request struct {
		Name      models.Translation `json:"name"`
		OrderUuid string             `json:"order_uuid"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "could not parse request",
			"code":    ErrorParse,
		})
		return
	}

	tx := models.DB.Begin()

	name := r.Name
	name.Timestamp = time.Now()
	name.Uuid = utils.GenerateShortUuid()
	if err := tx.Create(&name).Error; err != nil {
		fmt.Println(err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "could not insert translation",
			"code":    ErrorInsert,
		})
		return
	}

	var order models.Order
	if err := tx.Where("uuid=?", r.OrderUuid).First(&order).Error; err != nil {
		fmt.Println(err)
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "could not find order",
			"code":    ErrorRecordNotFound,
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
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "could not insert species",
			"code":    ErrorInsert,
		})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, species)
}

func SpeciesList(c *gin.Context) {
	var result []models.Species

	if err := models.DB.Preload("Name").Preload("Order.Name").Find(&result).Error; err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "species not found",
			"code":    ErrorRecordNotFound,
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

func SpeciesSetRarity(c *gin.Context) {
	type request struct {
		Category string `json:"category"`
		ZoneUuid string `json:"zone_uuid"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "could not parse request",
			"code":    ErrorParse,
		})
		return
	}

	if r.Category != "a" && r.Category != "b" && r.Category != "c" && r.Category != "d" && r.Category != "e" && r.Category != "-" {
		fmt.Println("invalid category")
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid category",
			"code":    ErrorRecordNotFound,
		})
		return
	}

	tx := models.DB.Begin()

	var species models.Species
	if err := tx.Where("uuid=?", c.Param("uuid")).First(&species).Error; err != nil {
		fmt.Println(err)
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "could not find species",
			"code":    ErrorRecordNotFound,
		})
		return
	}

	var zone models.Zone
	if err := tx.Where("uuid=?", r.ZoneUuid).First(&zone).Error; err != nil {
		fmt.Println(err)
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "could not find zone",
			"code":    ErrorRecordNotFound,
		})
		return
	}

	var rarity models.Rarity
	if err := tx.Where("zone_id=?", zone.Id).Where("species_id=?", species.Id).First(&rarity).Error; err != nil {
		if r.Category == "-" {
			if err := tx.Delete(&rarity).Error; err != nil {
				tx.Rollback()
				fmt.Println(err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "could not delete rarity",
					"code":    ErrorDelete,
				})
				return
			}
		} else {
			if err := models.DB.Table(table.Rarities).Where("rarity_id=?", rarity.Id).UpdateColumn("category", r.Category).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}
		}
	} else {
		rarity = models.Rarity{
			SpeciesId: species.Id,
			ZoneId:    zone.Id,
			Category:  r.Category,
		}
		if err := tx.Create(&rarity).Error; err != nil {
			tx.Rollback()
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "could not insert rarity",
				"code":    ErrorInsert,
			})
			return
		}
	}

	tx.Commit()

	c.JSON(http.StatusNoContent, nil)
}

func SpeciesDetails(c *gin.Context) {
	var result models.Species

	type response struct {
		Result     models.Species     `json:"result"`
		Recordings []models.Recording `json:"recordings"`
		Wikipedia  wikipedia.Data     `json:"wikipedia"`
	}

	lang := c.GetString("language")
	cacheKey := c.Request.RequestURI + lang
	cached, err := models.Cache.Get(cacheKey).Result()
	if err == redis.Nil {
	} else if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "internal server error",
			"code":    ErrorInternalServerError,
		})
		return
	} else {
		if cached != "" {
			var r response
			json.Unmarshal([]byte(cached), &r)
			c.JSON(http.StatusOK, r)
			return
		}
	}

	if err := models.DB.Where("uuid=?", c.Param("uuid")).Preload("Name").Preload("Order.Name").First(&result).Error; err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "species not found",
			"code":    ErrorRecordNotFound,
		})
		return
	}

	q := result.Name.En.String
	q = strings.ReplaceAll(q, " ", "%20")

	recordings := xenocanto.Recordings(q).Recordings
	var recs []models.Recording
	for i, r := range recordings {
		if i >= 25 {
			break
		}
		recs = append(recs, models.Recording{
			Url:      r.File,
			Filename: r.Filename,
		})
	}

	r := response{
		Result:     result,
		Recordings: recs,
		Wikipedia:  wikipedia.GetData(c.GetString("language"), result),
	}

	if cached, err := json.Marshal(r); err == nil {
		models.Cache.Set(cacheKey, cached, time.Second*time.Duration(config.AppConfig.CacheDuration))
	}

	c.JSON(http.StatusOK, r)
}
