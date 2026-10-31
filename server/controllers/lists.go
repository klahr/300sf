// SPDX-License-Identifier: AGPL-3.0-or-later
package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/table"
	"github.com/klahr/300sf/server/utils"
	"gopkg.in/guregu/null.v4/zero"
	"gorm.io/gorm"
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

	from, errFrom := parseListDate(r.From.String)
	until, errUntil := parseListDate(r.Until.String)
	if errFrom != nil || errUntil != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorParse,
			"message": "from and until are YYYY-MM-DD or RFC 3339",
		})
		return
	}

	if !models.ValidVisibility(r.Visibility) {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Code: ErrorInvalidRequest, Message: "visibility is 0, 1 or 2"})
		return
	}

	list, err := CreateList(r.Template, r.Name, models.ListVisibility(r.Visibility), user, r.Color)
	if errors.Is(err, ErrTemplateNotFound) {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Code: ErrorRecordNotFound, Message: err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	list.From = from
	list.Until = until

	if err := models.DB.Create(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, list)
}

// parseListDate reads a list's from or until: a date, or a time in RFC 3339.
// Empty is no date.
func parseListDate(s string) (zero.Time, error) {
	if s == "" {
		return zero.Time{}, nil
	}
	if t, err := time.Parse(time.DateOnly, s); err == nil {
		return zero.TimeFrom(t), nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return zero.Time{}, err
	}
	return zero.TimeFrom(t), nil
}

// ErrTemplateNotFound: the template is unknown, or another user's private one.
var ErrTemplateNotFound = errors.New("template not found")

func CreateList(template zero.String, name models.Translation, visibility int64, user models.User, color zero.String) (*models.List, error) {
	// The template first, so a bad one leaves no name behind.
	var t models.ListTemplate
	if template.Valid {
		if err := models.DB.Where("uuid=?", template.String).Where("user_id=? OR visibility=?", user.Id, models.VisibilityPublic).Preload("Species").First(&t).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrTemplateNotFound
			}
			return nil, err
		}
	}

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

// ListsTimespan sets a list's from and until: a date or an RFC 3339 time,
// "*" for none, and left as it is when empty.
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

	updates := map[string]interface{}{}
	for column, value := range map[string]zero.String{"from": r.From, "until": r.Until} {
		if !value.Valid {
			continue
		}
		if value.String == "*" {
			updates[column] = nil
			continue
		}
		t, err := parseListDate(value.String)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    ErrorParse,
				"message": "from and until are YYYY-MM-DD, RFC 3339 or *",
			})
			return
		}
		updates[column] = t
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

	if len(updates) > 0 {
		if err := models.DB.Table(table.Lists).Where("id=?", list.Id).UpdateColumns(updates).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListsVisibility sets who can see a list: private, anyone with the link, or public.
func ListsVisibility(c *gin.Context) {
	type request struct {
		Visibility *int64 `json:"visibility"`
	}

	// A missing visibility must not quietly make the list private.
	var r request
	if err := c.ShouldBindJSON(&r); err != nil || r.Visibility == nil || !models.ValidVisibility(*r.Visibility) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorInvalidRequest,
			"message": "visibility is 0, 1 or 2",
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

	// An old 6-character id is too easy to guess to be shared, so the list gets
	// a new id when it stops being private; links to the old id stop working.
	visibility := models.ListVisibility(*r.Visibility)
	uuid := list.Uuid
	if visibility != models.VisibilityPrivate && len(uuid) < utils.ShortUuidLength {
		uuid = utils.GenerateShortUuid()
	}
	if err := models.DB.Model(&list).UpdateColumns(map[string]any{"visibility": visibility, "uuid": uuid}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"uuid": uuid})
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

	if err := models.DB.Where("uuid=?", c.Param("uuid")).Where("user_id=? OR (visibility<>? AND CHAR_LENGTH(uuid)>=?)", user.Id, models.VisibilityPrivate, utils.ShortUuidLength).Where("archived IS NULL").Preload("Name").Preload("Species.Name").Preload("Species.Order.Name").First(&result).Error; err != nil {
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

	if err := tx.Table("list_reports").Select("reports.*").Where("reports.archived IS NULL").Where("list_id=?", list.Id).Joins("JOIN reports ON reports.id=report_id").Preload("Species.Name").Preload("Zone", withoutPolygon).Preload("Zone.Name").Preload("Age.Name").Preload("Gender.Name").Preload("Method.Name").Preload("Tags", "archived IS NULL").Find(&result).Error; err != nil {
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

// CompareList says which list an entry is: one user can share several.
type CompareList struct {
	Uuid string             `json:"uuid"`
	Name models.Translation `json:"name"`
}

type CompareEntry struct {
	List    CompareList     `json:"list"`
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
// times or comments. The first uuid is the reference list. Lists with old
// 6-character ids are left out: those ids are short enough to guess.
func CompareLists(uuids []string) (*CompareResult, error) {
	if len(uuids) == 0 || len(uuids) > 20 {
		return nil, fmt.Errorf("between 1 and 20 uuids required")
	}

	var lists []models.List
	if err := models.DB.Where("uuid IN ?", uuids).Where("visibility<>?", models.VisibilityPrivate).Where("CHAR_LENGTH(uuid)>=?", utils.ShortUuidLength).Where("archived IS NULL").Preload("Name").Find(&lists).Error; err != nil {
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

	// In the order asked for, so the reference comes first.
	byUuid := map[string]models.List{}
	for _, list := range lists {
		byUuid[list.Uuid] = list
	}
	for _, uuid := range uuids {
		list, ok := byUuid[uuid]
		if !ok {
			continue
		}
		delete(byUuid, uuid)
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
		entry := CompareEntry{
			List:    CompareList{Uuid: list.Uuid, Name: list.Name},
			User:    CompareUser{DisplayName: owner.DisplayName},
			Reports: []CompareReport{},
		}
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
	if throttle(c, "compare/"+ClientNetwork(c.ClientIP()), 60, time.Minute) {
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

// SharedSpecies is a species on a shared list and when it was first seen
// there; Seen is zero when it hasn't been.
type SharedSpecies struct {
	Uuid  string
	Name  models.Translation
	Order models.Translation
	Seen  time.Time
}

// SharedList is a list's progress as its share page shows it: names, species
// and the day each was first seen, never where or anything else.
type SharedList struct {
	Name    models.Translation
	Color   string
	Owner   string
	Seen    int
	Species []SharedSpecies
	// Latest are the most recently seen species, newest first.
	Latest []SharedSpecies
}

// ShareList is the progress of a list that is not private, under the same
// rules as CompareLists.
func ShareList(uuid string) (*SharedList, error) {
	var list models.List
	if err := models.DB.Where("uuid=?", uuid).Where("visibility<>?", models.VisibilityPrivate).Where("CHAR_LENGTH(uuid)>=?", utils.ShortUuidLength).Where("archived IS NULL").Preload("Name").First(&list).Error; err != nil {
		return nil, err
	}
	var owner models.User
	if err := models.DB.Where("id=?", list.UserId).First(&owner).Error; err != nil {
		return nil, err
	}

	var species []models.Species
	if err := models.DB.Table("species").
		Select("species.*").
		Joins("JOIN list_species ON list_species.species_id=species.id").
		Where("list_species.list_id=?", list.Id).
		Order("species.id").
		Preload("Name").
		Preload("Order.Name").
		Find(&species).Error; err != nil {
		return nil, err
	}

	var firsts []struct {
		SpeciesId int64
		First     time.Time
	}
	if err := models.DB.Table("list_reports").
		Select("reports.species_id AS species_id, MIN(reports.`from`) AS first").
		Joins("JOIN reports ON reports.id=list_reports.report_id").
		Where("list_reports.list_id=?", list.Id).
		Where("reports.archived IS NULL").
		Group("reports.species_id").
		Scan(&firsts).Error; err != nil {
		return nil, err
	}
	seen := map[int64]time.Time{}
	for _, f := range firsts {
		seen[f.SpeciesId] = f.First
	}

	result := SharedList{Name: list.Name, Color: list.Color.String, Owner: owner.DisplayName}
	for _, s := range species {
		entry := SharedSpecies{Uuid: s.Uuid, Name: s.Name, Order: s.Order.Name, Seen: seen[s.Id]}
		result.Species = append(result.Species, entry)
		if !entry.Seen.IsZero() {
			result.Seen++
			result.Latest = append(result.Latest, entry)
		}
	}
	sort.SliceStable(result.Latest, func(i, j int) bool { return result.Latest[i].Seen.After(result.Latest[j].Seen) })
	if len(result.Latest) > 5 {
		result.Latest = result.Latest[:5]
	}
	return &result, nil
}

// ShareListOwner is the uuid of the owner of a list that can be shared, for
// its share page's live updates.
func ShareListOwner(uuid string) (string, error) {
	var owners []string
	if err := models.DB.Table("lists").Joins("JOIN users ON users.id=lists.user_id").
		Where("lists.uuid=?", uuid).Where("lists.visibility<>?", models.VisibilityPrivate).
		Where("CHAR_LENGTH(lists.uuid)>=?", utils.ShortUuidLength).Where("lists.archived IS NULL").
		Limit(1).Pluck("users.uuid", &owners).Error; err != nil {
		return "", err
	}
	if len(owners) == 0 {
		return "", gorm.ErrRecordNotFound
	}
	return owners[0], nil
}
