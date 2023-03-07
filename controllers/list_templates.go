package controllers

import (
	"net/http"
	"time"

	"git.sr.ht/~klahr/300sf/models"
	"git.sr.ht/~klahr/300sf/table"
	"git.sr.ht/~klahr/300sf/utils"
	"github.com/gin-gonic/gin"
)

func ListTemplates(c *gin.Context) {
	type request struct {
		Name       models.Translation `json:"name"`
		Visibility int64              `json:"visibility"`
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

	listTemplate := models.ListTemplate{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateUuid(),
		},
		UserId:     user.Id,
		Visibility: r.Visibility,
		Name:       r.Name,
	}

	if err := models.DB.Create(&listTemplate).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
	}

	c.JSON(http.StatusOK, listTemplate)
}

func ListTemplatesArchive(c *gin.Context) {
	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := models.DB.Table(table.ListTemplates).Where("user_id=?", user.Id).Where("uuid=?", c.Param("uuid")).Where("archived IS NULL").UpdateColumn("archived", time.Now()).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func ListTemplatesList(c *gin.Context) {
	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var listTemplates []models.ListTemplate
	if err := models.DB.Table(table.ListTemplates).Where("user_id=? OR visibility=?", user.Id, models.VisibilityPublic).Where("archived IS NULL").Preload("Name").Preload("Species").Find(&listTemplates).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	type response struct {
		Result []models.ListTemplate `json:"result"`
	}

	c.JSON(http.StatusOK, response{
		Result: listTemplates,
	})
}

func ListTemplatesAddSpecies(c *gin.Context) {
	type request struct {
		Uuid        string `json:"uuid"`
		SpeciesUuid string `json:"species_uuid"`
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

	var listTemplate models.ListTemplate
	if err := models.DB.Where("user_id=?", user.Id).Where("uuid=?", r.Uuid).Preload("Species").Find(&listTemplate).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var species models.Species
	if err := models.DB.Where("uuid=?", r.SpeciesUuid).Find(&species).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	type list_template_species struct {
		Timestamp      time.Time `json:"timestamp"`
		ListTemplateId int64     `json:"list_template_id"`
		SpeciesId      int64     `json:"species_id"`
	}

	if err := models.DB.Table("list_template_species").Create(&list_template_species{
		Timestamp:      time.Now(),
		ListTemplateId: listTemplate.Id,
		SpeciesId:      species.Id,
	}).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func ListTemplatesRemoveSpecies(c *gin.Context) {
	type request struct {
		Uuid        string `json:"uuid"`
		SpeciesUuid string `json:"species_uuid"`
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

	var listTemplate models.ListTemplate
	if err := models.DB.Where("user_id=?", user.Id).Where("uuid=?", r.Uuid).Preload("Species").Find(&listTemplate).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var species models.Species
	if err := models.DB.Where("uuid=?", r.SpeciesUuid).Find(&species).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	type list_template_species struct {
		Timestamp      time.Time `json:"timestamp"`
		ListTemplateId int64     `json:"list_template_id"`
		SpeciesId      int64     `json:"species_id"`
	}

	if err := models.DB.Where("list_template_id=?", listTemplate.Id).Where("species_id=?", species.Id).Delete(&list_template_species{}).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
