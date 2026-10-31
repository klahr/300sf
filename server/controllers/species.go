// SPDX-License-Identifier: AGPL-3.0-or-later
package controllers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/klahr/300sf/server/config"
	"github.com/klahr/300sf/server/log"
	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/table"
	"github.com/klahr/300sf/server/utils"
	"github.com/klahr/300sf/server/wikipedia"
	"github.com/klahr/300sf/server/xenocanto"
)

func Species(c *gin.Context) {
	type request struct {
		Name      models.Translation `json:"name"`
		OrderUuid string             `json:"order_uuid"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "could not parse request",
			"code":    ErrorParse,
		})
		return
	}

	tx := models.DB.Begin()
	defer tx.Rollback()

	name := r.Name
	name.Timestamp = time.Now()
	name.Uuid = utils.GenerateShortUuid()
	if err := tx.Create(&name).Error; err != nil {
		log.Error(c, 0, err.Error())
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "could not insert translation",
			"code":    ErrorInsert,
		})
		return
	}

	var order models.Order
	if err := tx.Where("uuid=?", r.OrderUuid).First(&order).Error; err != nil {
		log.Error(c, 0, err.Error())
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
		log.Error(c, 0, err.Error())
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
		log.Error(c, 0, err.Error())
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
		log.Error(c, 0, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "could not parse request",
			"code":    ErrorParse,
		})
		return
	}

	if r.Category != "a" && r.Category != "b" && r.Category != "c" && r.Category != "d" && r.Category != "e" && r.Category != "-" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid category",
			"code":    ErrorRecordNotFound,
		})
		return
	}

	tx := models.DB.Begin()
	defer tx.Rollback()

	var species models.Species
	if err := tx.Where("uuid=?", c.Param("uuid")).First(&species).Error; err != nil {
		log.Error(c, 0, err.Error())
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "could not find species",
			"code":    ErrorRecordNotFound,
		})
		return
	}

	var zone models.Zone
	if err := tx.Where("uuid=?", r.ZoneUuid).First(&zone).Error; err != nil {
		log.Error(c, 0, err.Error())
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "could not find zone",
			"code":    ErrorRecordNotFound,
		})
		return
	}

	var rarity models.Rarity
	if err := tx.Where("zone_id=?", zone.Id).Where("species_id=?", species.Id).First(&rarity).Error; err == nil {
		if r.Category == "-" {
			if err := tx.Delete(&rarity).Error; err != nil {
				tx.Rollback()
				log.Error(c, 0, err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "could not delete rarity",
					"code":    ErrorDelete,
				})
				return
			}
		} else {
			if err := tx.Table(table.Rarities).Where("id=?", rarity.Id).UpdateColumn("category", string(strings.ToUpper(r.Category)[0])).Error; err != nil {
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
			Category:  strings.ToUpper(r.Category),
		}
		if err := tx.Create(&rarity).Error; err != nil {
			tx.Rollback()
			log.Error(c, 0, err.Error())
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

	lang := wikipedia.Language(c.GetString("app_language"))
	// Keyed by species, not path: the same details are asked for at / and /api.
	cacheKey := "/v1/species/" + c.Param("uuid") + "?lang=" + lang
	cached, err := models.Cache.Get(cacheKey).Result()
	if err == redis.Nil {
	} else if err != nil {
		log.Error(c, 0, err.Error())
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

	// Not yet cached means asking Wikipedia and xeno-canto, which visitors
	// without an account may only do now and then.
	if c.GetHeader("Authorization") == "" && Throttled("species/"+ClientNetwork(c.ClientIP()), 20, time.Minute) {
		c.JSON(http.StatusTooManyRequests, gin.H{"message": "too many requests", "code": ErrorTooManyApiCalls})
		return
	}

	if err := models.DB.Where("uuid=?", c.Param("uuid")).Preload("Name").Preload("Order.Name").First(&result).Error; err != nil {
		log.Error(c, 0, err.Error())
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
			Url:       r.File,
			Filename:  r.Filename,
			Type:      r.Type,
			Sex:       r.Sex,
			Stage:     r.Stage,
			Method:    r.Method,
			Sono:      r.Sono.Small,
			Recorder:  r.Rec,
			Country:   r.Cnt,
			Location:  r.Loc,
			Latitude:  utils.Str2Float(r.Lat),
			Longitude: utils.Str2Float(r.Lng),
		})
	}

	if lang == "sv" && !result.Name.Sv.Valid {
		lang = "en"
	} else if lang == "de" && !result.Name.De.Valid {
		lang = "en"
	}

	r := response{
		Result:     result,
		Recordings: recs,
		Wikipedia:  wikipedia.GetData(lang, result.Name),
	}

	if cached, err := json.Marshal(r); err == nil {
		models.Cache.Set(cacheKey, cached, time.Second*time.Duration(config.AppConfig.WikipediaCacheDuration))
	}

	c.JSON(http.StatusOK, r)
}

// searchCatalog is every species with its names, and which are Swedish (on a
// public species list), so that search needs no query per key press.
type searchCatalog struct {
	species []models.Species
	swedish map[int64]bool
	at      time.Time
}

var (
	searchCatalogMu sync.Mutex
	searchCatalogAt searchCatalog
)

func loadSearchCatalog() (searchCatalog, error) {
	searchCatalogMu.Lock()
	defer searchCatalogMu.Unlock()
	if time.Since(searchCatalogAt.at) < 10*time.Minute {
		return searchCatalogAt, nil
	}
	var species []models.Species
	if err := models.DB.Preload("Name").Find(&species).Error; err != nil {
		return searchCatalogAt, err
	}
	var swedish []int64
	if err := models.DB.Table("list_template_species").
		Joins("JOIN list_templates ON list_templates.id=list_template_species.list_template_id").
		Where("list_templates.visibility=?", models.VisibilityPublic).
		Where("list_templates.archived IS NULL").
		Distinct().
		Pluck("list_template_species.species_id", &swedish).Error; err != nil {
		return searchCatalogAt, err
	}
	searchCatalogAt = searchCatalog{species: species, swedish: map[int64]bool{}, at: time.Now()}
	for _, id := range swedish {
		searchCatalogAt.swedish[id] = true
	}
	return searchCatalogAt, nil
}

// SpeciesSearch finds species by name for anyone, signed in or not, ranked
// like the app's search: names that start with the query first, Swedish
// species before others.
func SpeciesSearch(c *gin.Context) {
	if c.GetHeader("Authorization") == "" && Throttled("species-search/"+ClientNetwork(c.ClientIP()), 120, time.Minute) {
		c.JSON(http.StatusTooManyRequests, gin.H{"message": "too many requests", "code": ErrorTooManyApiCalls})
		return
	}
	query := strings.ToLower(strings.TrimSpace(c.Query("q")))
	type response struct {
		Result []models.Species `json:"result"`
	}
	if len([]rune(query)) < 2 || len(query) > 64 {
		c.JSON(http.StatusOK, response{Result: []models.Species{}})
		return
	}
	catalog, err := loadSearchCatalog()
	if err != nil {
		log.Error(c, 0, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error", "code": ErrorInternalServerError})
		return
	}

	lang := c.Query("lang")
	type scored struct {
		species models.Species
		name    string
		score   int
	}
	var found []scored
	for _, s := range catalog.species {
		name := strings.ToLower(utils.Translation2Value(s.Name, lang))
		if name == "" {
			name = strings.ToLower(s.Name.En.String)
		}
		latin := strings.ToLower(s.Name.La.String)
		score := -1
		switch {
		case strings.HasPrefix(name, query):
			score = 0
		case wordStartsWith(name, query):
			score = 1
		case strings.Contains(name, query) || strings.HasPrefix(latin, query):
			score = 2
		case strings.Contains(latin, query):
			score = 3
		}
		if score < 0 {
			continue
		}
		if !catalog.swedish[s.Id] {
			score += 10
		}
		found = append(found, scored{species: s, name: name, score: score})
	}
	sort.Slice(found, func(i, j int) bool {
		if found[i].score != found[j].score {
			return found[i].score < found[j].score
		}
		return found[i].name < found[j].name
	})
	result := []models.Species{}
	for i := 0; i < len(found) && i < 8; i++ {
		result = append(result, found[i].species)
	}
	c.JSON(http.StatusOK, response{Result: result})
}

func wordStartsWith(name, query string) bool {
	for _, w := range strings.FieldsFunc(name, func(r rune) bool { return r == ' ' || r == '-' }) {
		if strings.HasPrefix(w, query) {
			return true
		}
	}
	return false
}
