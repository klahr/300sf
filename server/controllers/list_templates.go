// SPDX-License-Identifier: AGPL-3.0-or-later
package controllers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/server/auth"
	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/table"
	"github.com/klahr/300sf/server/utils"
	"gorm.io/gorm"
)

func ListTemplates(c *gin.Context) {
	type request struct {
		Name         models.Translation `json:"name"`
		Visibility   int64              `json:"visibility"`
		Description  models.Translation `json:"description"`
		SpeciesUuids []string           `json:"species_uuids"`
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

	var user models.User
	if err := tx.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	// Only admins share species lists with everyone.
	if r.Visibility == models.VisibilityPublic && !user.HasScope(auth.AuthScopeAdmin) {
		r.Visibility = models.VisibilityPrivate
	}

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

	description := r.Description
	description.Timestamp = time.Now()
	description.Uuid = utils.GenerateShortUuid()
	if err := tx.Create(&description).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	listTemplate := models.ListTemplate{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateShortUuid(),
		},
		UserId:      user.Id,
		Visibility:  r.Visibility,
		Name:        name,
		Description: description,
	}

	if err := tx.Create(&listTemplate).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if len(r.SpeciesUuids) > 0 {
		if err := setListTemplateSpecies(tx, listTemplate.Id, r.SpeciesUuids); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}
	}

	tx.Commit()

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
	models.DB.Where("uuid=?", c.GetString("uuid")).First(&user)

	var listTemplates []models.ListTemplate
	if err := models.DB.Table(table.ListTemplates).Where("user_id=? OR visibility=?", user.Id, models.VisibilityPublic).Where("archived IS NULL").Preload("Name").Preload("Description").Preload("Species").Find(&listTemplates).Error; err != nil {
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

func ListTemplatesListMe(c *gin.Context) {
	var user models.User
	models.DB.Where("uuid=?", c.GetString("uuid")).First(&user)

	var listTemplates []models.ListTemplate
	if err := models.DB.Table(table.ListTemplates).Where("user_id=?", user.Id).Where("archived IS NULL").Preload("Name").Preload("Description").Preload("Species").Find(&listTemplates).Error; err != nil {
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
	if err := models.DB.Where("user_id=?", user.Id).Where("uuid=?", c.Param("uuid")).Preload("Species").First(&listTemplate).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "template not found",
		})
		return
	}

	var species models.Species
	if err := models.DB.Where("uuid=?", r.SpeciesUuid).First(&species).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := models.DB.Table("list_template_species").Create(&models.ListTemplateSpecies{
		ListTemplateId: listTemplate.Id,
		SpeciesId:      species.Id,
	}).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "already added",
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func ListTemplatesRemoveSpecies(c *gin.Context) {
	type request struct {
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
	if err := models.DB.Where("user_id=?", user.Id).Where("uuid=?", c.Param("uuid")).Preload("Species").First(&listTemplate).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "template not found",
		})
		return
	}

	var species models.Species
	if err := models.DB.Where("uuid=?", r.SpeciesUuid).First(&species).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "species not found",
		})
		return
	}

	if err := models.DB.Where("list_template_id=?", listTemplate.Id).Where("species_id=?", species.Id).Delete(&models.ListTemplateSpecies{}).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func ListTemplatesName(c *gin.Context) {
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

	var listTemplate models.ListTemplate
	if err := models.DB.Where("user_id=?", user.Id).Where("uuid=?", c.Param("uuid")).First(&listTemplate).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := models.DB.Model(&models.Translation{}).Where("id=?", listTemplate.NameId).Updates(translationText(r.Name)).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func ListTemplatesDescription(c *gin.Context) {
	type request struct {
		Description models.Translation `json:"description"`
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
	if err := models.DB.Where("user_id=?", user.Id).Where("uuid=?", c.Param("uuid")).First(&listTemplate).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := models.DB.Model(&models.Translation{}).Where("id=?", listTemplate.DescriptionId).Updates(translationText(r.Description)).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func ListTemplatesDetails(c *gin.Context) {
	var result models.ListTemplate

	var user models.User
	models.DB.Where("uuid=?", c.GetString("uuid")).First(&user)

	if err := models.DB.Where("uuid=?", c.Param("uuid")).Where("user_id=? OR visibility=?", user.Id, models.VisibilityPublic).Preload("Name").Preload("Description").Preload("Species.Name").Preload("Species.Order.Name").First(&result).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	type response struct {
		Result models.ListTemplate `json:"result"`
	}

	c.JSON(http.StatusOK, response{
		Result: result,
	})
}

// setListTemplateSpecies makes speciesUuids the species of the template, and of every checklist made
// from it, so the lists follow their species list. Sightings stay on the checklists.
func setListTemplateSpecies(tx *gorm.DB, templateId int64, speciesUuids []string) error {
	var species []models.Species
	if err := tx.Where("uuid IN ?", speciesUuids).Find(&species).Error; err != nil {
		return err
	}
	if len(species) != len(speciesUuids) {
		return errors.New("unknown species")
	}
	ids := make([]int64, len(species))
	for i := range species {
		ids[i] = species[i].Id
	}

	if err := tx.Where("list_template_id=?", templateId).Delete(&models.ListTemplateSpecies{}).Error; err != nil {
		return err
	}
	rows := make([]models.ListTemplateSpecies, len(ids))
	for i := range ids {
		rows[i] = models.ListTemplateSpecies{ListTemplateId: templateId, SpeciesId: ids[i]}
	}
	if len(rows) > 0 {
		if err := tx.Table(table.ListTemplateSpecies).Create(&rows).Error; err != nil {
			return err
		}
	}

	lists := tx.Table(table.Lists).Select("id").Where("list_template_id=?", templateId)
	if err := tx.Exec("DELETE FROM list_species WHERE list_id IN (?)", lists).Error; err != nil {
		return err
	}
	if len(ids) > 0 {
		return tx.Exec("INSERT INTO list_species (list_id, species_id) SELECT l.id, s.id FROM lists l JOIN species s ON s.id IN ? WHERE l.list_template_id=?", ids, templateId).Error
	}
	return nil
}

// ListTemplatesSpecies replaces the species on the user's own species list.
func ListTemplatesSpecies(c *gin.Context) {
	type request struct {
		SpeciesUuids []string `json:"species_uuids"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil || len(r.SpeciesUuids) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "species required",
		})
		return
	}

	tx := models.DB.Begin()
	defer tx.Rollback()

	var user models.User
	if err := tx.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var listTemplate models.ListTemplate
	if err := tx.Where("user_id=?", user.Id).Where("uuid=?", c.Param("uuid")).Where("archived IS NULL").First(&listTemplate).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "template not found",
		})
		return
	}

	if err := setListTemplateSpecies(tx, listTemplate.Id, r.SpeciesUuids); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	tx.Commit()

	c.JSON(http.StatusNoContent, nil)
}

// ListTemplatesVisibility shares an admin's species list with everyone, or makes it private again.
func ListTemplatesVisibility(c *gin.Context) {
	type request struct {
		Visibility int64 `json:"visibility"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil || (r.Visibility != models.VisibilityPrivate && r.Visibility != models.VisibilityPublic) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid visibility",
		})
		return
	}

	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	result := models.DB.Table(table.ListTemplates).Where("user_id=?", user.Id).Where("uuid=?", c.Param("uuid")).Where("archived IS NULL").UpdateColumn("visibility", r.Visibility)
	if result.Error != nil || result.RowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "template not found",
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
