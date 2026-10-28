package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/klahr/300sf/api/config"
	"github.com/klahr/300sf/api/models"
	"github.com/klahr/300sf/api/wikipedia"
)

func OrderDetails(c *gin.Context) {
	var result models.Order

	type response struct {
		Result    models.Order   `json:"result"`
		Wikipedia wikipedia.Data `json:"wikipedia"`
	}

	lang := wikipedia.Language(c.GetString("app_language"))
	cacheKey := c.Request.URL.Path + "?lang=" + lang
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

	if err := models.DB.Where("uuid=?", c.Param("uuid")).Preload("Name").First(&result).Error; err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "order not found",
			"code":    ErrorRecordNotFound,
		})
		return
	}

	r := response{
		Result:    result,
		Wikipedia: wikipedia.GetData(lang, result.Name),
	}

	if cached, err := json.Marshal(r); err == nil {
		models.Cache.Set(cacheKey, cached, time.Second*time.Duration(config.AppConfig.WikipediaCacheDuration))
	}

	c.JSON(http.StatusOK, r)
}
