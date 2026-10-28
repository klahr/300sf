// SPDX-License-Identifier: GPL-2.0-or-later
package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/klahr/300sf/api/config"
	"github.com/klahr/300sf/api/models"
	"github.com/klahr/300sf/api/table"
	"github.com/klahr/300sf/api/utils"
	"gopkg.in/guregu/null.v4/zero"
)

type tag struct {
	Id       int64  `json:"id"`
	Name     string `json:"name"`
	Archived int64  `json:"archived"`
}

type tag_eukaryote struct {
	TagId       int64 `json:"tagid"`
	EukaryoteId int64 `json:"eukaryoteid"`
	Assigned    int64 `json:"assigned"`
}

type list_visible struct {
	ListId   int64 `json:"listid"`
	BirdId   int64 `json:"birdid"`
	Assigned int64 `json:"assigned"`
}

type list struct {
	ListId  int64   `json:"listid"`
	Name    string  `json:"name"`
	NameEn  string  `json:"name_en"`
	IapId   string  `json:"iapid"`
	ClassId int64   `json:"classid"`
	Content []int64 `json:"content"`
}

type discovered struct {
	EukaryoteId int64  `json:"eukaryoteid"`
	Name        string `json:"name"`
	DateSeen    string `json:"dateseen"`
	Place       string `json:"place"`
	Comment     string `json:"comment"`
	Discovered  bool   `json:"discovered"`
}

func migrate(user models.User, username, password string) (interface{}, int, error) {
	if user.ImportedAt.Valid {
		return nil, ErrorMigrateCompleted, fmt.Errorf("already imported")
	}

	loginJsonRequest := fmt.Sprintf("{\"jsonrpc\":\"2.0\",\"method\":\"SessionManager.create\",\"params\": {\"email\":\"%s\",\"password\":\"%s\"},\"id\":1}", username, password)
	loginResponse, err := http.PostForm("https://300sf.se/legacy/ajax_api.php", url.Values{
		"data": {loginJsonRequest},
	})

	if err != nil {
		return nil, ErrorInternalServerError, err
	}

	defer loginResponse.Body.Close()
	body, _ := io.ReadAll(loginResponse.Body)

	type jsonrpc_response struct {
		Result struct {
			Token  string `json:"token"`
			UserId int64  `json:"userid"`
		}
	}

	var response jsonrpc_response
	json.Unmarshal(body, &response)

	if response.Result.UserId == 0 {
		return nil, ErrorMigrateInvalidCredentials, fmt.Errorf("invalid username or password")
	}

	dataJsonRequest := fmt.Sprintf("{\"jsonrpc\":\"2.0\",\"method\":\"DataManager.sync\",\"params\": {\"userid\":%d},\"id\":1}", response.Result.UserId)
	dataResponse, err := http.PostForm("https://300sf.se/legacy/ajax_api.php", url.Values{
		"data": {dataJsonRequest},
	})

	if err != nil {
		return nil, ErrorInternalServerError, err
	}

	defer dataResponse.Body.Close()
	body, _ = io.ReadAll(dataResponse.Body)

	type data struct {
		JsonRpc string `json:"jsonrpc"`
		Result  struct {
			Discovered    []discovered    `json:"discovered"`
			Lists         []list          `json:"lists"`
			Tags          []tag           `json:"tags"`
			TagEukaryotes []tag_eukaryote `json:"tag_eukaryotes"`
			ListVisibles  []list_visible
		} `json:"result"`
	}

	var d data
	json.Unmarshal(body, &d)

	species := fetchAllSpecies()

	tx := models.DB.Begin()

	var userLists []models.List
	if err := tx.Where("lists.user_id=?", user.Id).Preload("ListTemplate.Species").Joins("JOIN list_templates ON list_templates.id=lists.list_template_id").Where("list_templates.uuid IN (?)", config.AppConfig.DefaultLists).Find(&userLists).Error; err != nil {
		tx.Rollback()
		return nil, ErrorInternalServerError, err
	}

	for _, dc := range d.Result.Discovered {
		if dc.Discovered {
			if b := getBird(species, dc.Name); b != nil {
				report := models.Report{
					Model: models.Model{
						Timestamp: time.Now(),
						Uuid:      utils.GenerateShortUuid(),
					},
					SpeciesId: b.Id,
					UserId:    user.Id,
					Quantity:  zero.IntFrom(1),
					Comment: &models.Comment{
						Model: models.Model{
							Timestamp: time.Now(),
							Uuid:      utils.GenerateShortUuid(),
						},
						UserId: user.Id,
						Text:   dc.Comment,
					},
				}

				if dc.DateSeen != "" {
					dateSeen, _ := time.Parse("2006-01-02", dc.DateSeen)
					report.From = zero.TimeFrom(dateSeen)
				}

				if dc.Place != "" {
					report.Site = zero.StringFrom(dc.Place)
				}

				tags := getTags(d.Result.Tags, d.Result.TagEukaryotes, dc.EukaryoteId)
				for _, t := range tags {
					var existingTag models.Tag
					if err := tx.Where("name=?", t.Name).Where("user_id=?", user.Id).First(&existingTag).Error; err == nil {
						report.Tags = append(report.Tags, existingTag)
					} else {
						report.Tags = append(report.Tags, models.Tag{
							Model: models.Model{
								Timestamp: time.Now(),
								Uuid:      utils.GenerateShortUuid(),
							},
							Name:   t.Name,
							UserId: user.Id,
						})
					}
				}

				tx.Create(&report)

				for _, t := range userLists {
					if isInSpeciesList(b.Id, t.ListTemplate.Species) {
						if err := tx.Create(&models.ListReport{
							ListId:   t.Id,
							ReportId: report.Id,
						}).Error; err != nil {
							tx.Rollback()
							return nil, ErrorInsert, err
						}
						break
					}
				}
			}
		}
	}

	tx.Table(table.Users).Where("id=?", user.Id).UpdateColumn("imported_at", time.Now())

	tx.Commit()

	return nil, 0, nil
}

func isInSpeciesList(speciesId int64, species []models.Species) bool {
	for _, s := range species {
		if s.Id == speciesId {
			return true
		}
	}
	return false
}

func getBird(birds []models.Species, name string) *models.Species {
	for _, b := range birds {
		if b.Name.Sv.String == name {
			return &b
		}
	}
	return nil
}

func getTags(tags []tag, te []tag_eukaryote, id int64) []tag {
	var ts []tag
	for _, t := range te {
		if t.Assigned == 1 && t.EukaryoteId == id {
			if tx := getTag(tags, t.TagId); tx != nil {
				ts = append(ts, *tx)
			}
		}
	}
	return ts
}

func getTag(tags []tag, id int64) *tag {
	for _, t := range tags {
		if t.Id == id && t.Archived != 1 {
			return &t
		}
	}
	return nil
}

func fetchAllSpecies() []models.Species {
	var birds []models.Species
	models.DB.Preload("Name").Find(&birds)

	return birds
}
