package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/api/models"
)

func AnalyzeResolve(c *gin.Context) {
	type bird struct {
		Name        string  `json:"name"`
		Probability float64 `json:"probability"`
	}
	type request struct {
		Birds []bird `json:"birds"`
	}
	type result struct {
		Species     models.Species `json:"species"`
		Probability float64        `json:"probability"`
	}
	type response struct {
		Results []result `json:"results"`
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

	var res response
	for _, b := range r.Birds {
		names := strings.Split(b.Name, "_")
		if len(names) < 2 {
			continue
		}
		var species models.Species
		if err := tx.Preload("Name").Joins("JOIN translations ON translations.id=name_id").Where("translations.en=? OR translations.la=?", names[1], names[0]).First(&species).Error; err != nil {
			continue
		}
		res.Results = append(res.Results, result{
			Species:     species,
			Probability: b.Probability,
		})
	}

	c.JSON(http.StatusOK, res)
}
