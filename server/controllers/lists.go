// SPDX-License-Identifier: AGPL-3.0-or-later
package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/table"
	"github.com/klahr/300sf/server/utils"
	"gopkg.in/guregu/null.v4/zero"
)

func Lists(c *gin.Context) {
	type request struct {
		Name       models.Translation `json:"name"`
		Visibility int64              `json:"visibility"`
		Template   zero.String        `json:"template"`
		From       zero.String        `json:"from"`
		Until      zero.String        `json:"until"`
		Color      zero.String        `json:"color"`
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

	list, err := CreateList(r.Template, r.Name, r.Visibility, user, r.Color)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if r.From.Valid {
		list.From = zero.TimeFrom(utils.Str2Time(r.From.String))
	}
	if r.Until.Valid {
		list.Until = zero.TimeFrom(utils.Str2Time(r.Until.String))
	}

	if err := models.DB.Create(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, list)
}

func CreateList(template zero.String, name models.Translation, visibility int64, user models.User, color zero.String) (*models.List, error) {
	name.Timestamp = time.Now()
	name.Uuid = utils.GenerateShortUuid()
	if err := models.DB.Create(&name).Error; err != nil {
		return nil, err
	}

	list := models.List{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateShortUuid(),
		},
		UserId:     user.Id,
		Visibility: visibility,
		Name:       name,
		Color:      color,
	}

	if template.Valid {
		var t models.ListTemplate
		if err := models.DB.Where("uuid=?", template.String).Where("user_id=? OR visibility=?", user.Id, models.VisibilityPublic).Preload("Species").First(&t).Error; err != nil {
			return nil, err
		}
		list.ListTemplateId = zero.IntFrom(t.Id)
		list.Species = t.Species
	}

	return &list, nil
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
	if err := models.DB.Table(table.Lists).Where("user_id=?", user.Id).Where("archived IS NULL").Preload("Name").Preload("Species.Name").Preload("Species.Order.Name").Preload("ListTemplate").Find(&lists).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	for i := range lists {
		lists[i].Template = lists[i].ListTemplate.Uuid
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

func ListsArchived(c *gin.Context) {
	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var lists []models.List
	if err := models.DB.Table(table.Lists).Where("user_id=?", user.Id).Where("archived IS NOT NULL").Order("archived DESC").Preload("Name").Preload("Species").Find(&lists).Error; err != nil {
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

func ListsRestore(c *gin.Context) {
	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := models.DB.Table(table.Lists).Where("user_id=?", user.Id).Where("uuid=?", c.Param("uuid")).Where("archived IS NOT NULL").UpdateColumn("archived", nil).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListsDelete removes an archived list for good. Its sightings stay; only their link to the list goes.
func ListsDelete(c *gin.Context) {
	tx := models.DB.Begin()
	defer tx.Rollback()

	var user models.User
	if err := tx.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var list models.List
	if err := tx.Where("user_id=?", user.Id).Where("uuid=?", c.Param("uuid")).Where("archived IS NOT NULL").First(&list).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := tx.Delete(&list).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := tx.Where("id=?", list.NameId).Delete(&models.Translation{}).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
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

	if err := models.DB.Model(&models.Translation{}).Where("id=?", list.NameId).Updates(translationText(r.Name)).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func ListsColor(c *gin.Context) {
	type request struct {
		Color zero.String `json:"color"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil || len(r.Color.String) > 16 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid color",
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

	if err := models.DB.Model(&list).Update("color", r.Color).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func ListsTimespan(c *gin.Context) {
	type request struct {
		From  zero.String `json:"from"`
		Until zero.String `json:"until"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
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

	var list models.List
	if err := tx.Where("user_id=?", user.Id).Where("uuid=?", c.Param("uuid")).First(&list).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if r.From.Valid {
		var err error
		if r.From.String == "*" {
			err = tx.Table(table.Lists).Where("id=?", list.Id).UpdateColumn("from", nil).Error
		} else {
			err = tx.Table(table.Lists).Where("id=?", list.Id).UpdateColumn("from", r.From.String).Error
		}
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}
	}

	if r.Until.Valid {
		var err error
		if r.Until.String == "*" {
			err = tx.Table(table.Lists).Where("id=?", list.Id).UpdateColumn("until", nil).Error
		} else {
			err = tx.Table(table.Lists).Where("id=?", list.Id).UpdateColumn("until", r.Until.String).Error
		}
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}
	}

	tx.Commit()

	c.JSON(http.StatusNoContent, nil)
}

func ListsDetails(c *gin.Context) {
	var result models.List

	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := models.DB.Where("uuid=?", c.Param("uuid")).Where("user_id=? OR visibility<>?", user.Id, models.VisibilityPrivate).Where("archived IS NULL").Preload("Name").Preload("Species.Name").Preload("Species.Order.Name").First(&result).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	type response struct {
		Result models.List `json:"result"`
	}

	c.JSON(http.StatusOK, response{
		Result: result,
	})
}

func ListsReports(c *gin.Context) {
	var result []models.Report

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

	var list models.List
	if err := tx.Where("user_id=?", user.Id).Where("uuid=?", c.Param("uuid")).First(&list).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := tx.Table("list_reports").Select("reports.*").Where("reports.archived IS NULL").Where("list_id=?", list.Id).Joins("JOIN reports ON reports.id=report_id").Preload("Species.Name").Preload("Zone.Name").Preload("Age.Name").Preload("Gender.Name").Preload("Method.Name").Preload("Tags", "archived IS NULL").Find(&result).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	tx.Commit()

	type response struct {
		Result []models.Report `json:"result"`
	}

	c.JSON(http.StatusOK, response{
		Result: result,
	})
}

func ListsForSpecies(c *gin.Context) {
	type listAvailable struct {
		List      models.List `json:"list"`
		Available bool        `json:"available"`
	}

	type response struct {
		Result []listAvailable `json:"result"`
	}

	var result response

	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	var species models.Species
	if err := models.DB.Where("uuid=?", c.Param("species_uuid")).First(&species).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	var availableLists []models.List
	if err := models.DB.Table("lists").Joins("JOIN list_species ON list_species.list_id=lists.id AND list_species.species_id=?", species.Id).Where("user_id=?", user.Id).Where("archived IS NULL").Find(&availableLists).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	var allLists []models.List
	if err := models.DB.Preload("Name").Where("user_id=?", user.Id).Where("archived IS NULL").Find(&allLists).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	for _, list := range allLists {
		found := false
		for _, a := range availableLists {
			if a.Uuid == list.Uuid {
				found = true
				break
			}
		}
		result.Result = append(result.Result, listAvailable{
			List:      list,
			Available: found,
		})
	}

	c.JSON(http.StatusOK, result)
}

type CompareUser struct {
	DisplayName string `json:"display_name"`
}

type CompareSpecies struct {
	Uuid string `json:"uuid"`
}

type CompareReport struct {
	Species CompareSpecies `json:"species"`
}

type CompareEntry struct {
	User    CompareUser     `json:"user"`
	Reports []CompareReport `json:"reports"`
}

// Has reports whether this list has a sighting of the species.
func (e CompareEntry) Has(speciesUuid string) bool {
	for _, r := range e.Reports {
		if r.Species.Uuid == speciesUuid {
			return true
		}
	}
	return false
}

type CompareResult struct {
	Species       []models.Species `json:"species"`
	Result        []CompareEntry   `json:"result"`
	ReferenceList models.List      `json:"reference_list"`
}

// CompareLists shows which species each list has, for lists that are not
// private. Only display names and species are returned, never positions,
// times or comments. The first uuid is the reference list.
func CompareLists(uuids []string) (*CompareResult, error) {
	if len(uuids) == 0 || len(uuids) > 20 {
		return nil, fmt.Errorf("between 1 and 20 uuids required")
	}

	var lists []models.List
	if err := models.DB.Where("uuid IN ?", uuids).Where("visibility<>?", models.VisibilityPrivate).Where("archived IS NULL").Preload("Name").Find(&lists).Error; err != nil {
		return nil, err
	}

	var resp CompareResult
	found := false
	for _, list := range lists {
		if list.Uuid == uuids[0] {
			resp.ReferenceList = list
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("not found")
	}

	if err := models.DB.Table("species").
		Select("species.*").
		Joins("JOIN list_species ON list_species.species_id=species.id").
		Where("list_species.list_id = ?", resp.ReferenceList.Id).
		Preload("Name").
		Find(&resp.Species).Error; err != nil {
		return nil, err
	}

	for _, list := range lists {
		var owner models.User
		if err := models.DB.Where("id=?", list.UserId).First(&owner).Error; err != nil {
			return nil, err
		}
		var speciesUuids []string
		if err := models.DB.Table("list_reports").
			Select("DISTINCT species.uuid").
			Joins("JOIN reports ON reports.id=list_reports.report_id JOIN species ON species.id=reports.species_id").
			Where("list_reports.list_id=?", list.Id).
			Where("reports.archived IS NULL").
			Pluck("species.uuid", &speciesUuids).Error; err != nil {
			return nil, err
		}
		entry := CompareEntry{User: CompareUser{DisplayName: owner.DisplayName}, Reports: []CompareReport{}}
		for _, u := range speciesUuids {
			entry.Reports = append(entry.Reports, CompareReport{Species: CompareSpecies{Uuid: u}})
		}
		resp.Result = append(resp.Result, entry)
	}

	return &resp, nil
}

func ListsCompare(c *gin.Context) {
	type request struct {
		Uuids []string `json:"uuids"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	// Anyone can call this, so limit how fast list ids can be guessed.
	if throttle(c, "compare/"+c.ClientIP(), 60, time.Minute) {
		return
	}

	resp, err := CompareLists(r.Uuids)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}
