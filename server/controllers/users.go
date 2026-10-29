// SPDX-License-Identifier: AGPL-3.0-or-later
package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/guregu/null.v4/zero"

	"github.com/klahr/300sf/server/auth"
	"github.com/klahr/300sf/server/config"
	"github.com/klahr/300sf/server/log"
	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/table"
	"github.com/klahr/300sf/server/utils"
)

func Users(c *gin.Context) {
	log.Info(c, 0, "Users")

	type request struct {
		Email         string `json:"email"`
		DisplayName   string `json:"display_name"`
		Username      string `json:"username"`
		Password      string `json:"password"`
		AcceptedTerms bool   `json:"accepted_terms"`
		Portrait      string `json:"portrait"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		log.Error(c, -10001, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	// Each sign-up sends an email, so limit how many one address can make.
	if throttle(c, "signup/"+c.ClientIP(), 10, time.Hour) {
		return
	}

	if r.Portrait != "" && !utils.VerifyPortrait(r.Portrait) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorParse,
			"message": "invalid portrait",
		})
		return
	}

	if !utils.VerifyDisplayName(r.DisplayName) {
		log.Error(c, ErrorSignUpInvalidDisplayName, "invalid display name")
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorSignUpInvalidDisplayName,
			"message": "invalid display name",
		})
		return
	}

	if !utils.VerifyUsername(r.Username) {
		log.Error(c, ErrorSignUpInvalidUsername, "invalid username")
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorSignUpInvalidUsername,
			"message": "invalid username",
		})
		return
	}

	if !utils.VerifyEmail(r.Email) {
		log.Error(c, ErrorSignUpInvalidEmail, "invalid email")
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorSignUpInvalidEmail,
			"message": "invalid email",
		})
		return
	}

	if !utils.VerifyPassword(r.Password) {
		log.Error(c, ErrorSignUpInvalidPassword, "invalid password")
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorSignUpInvalidPassword,
			"message": "invalid password",
		})
		return
	}

	if !r.AcceptedTerms {
		log.Error(c, ErrorSignUpTermsNotAccepted, "terms not accepted")
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorSignUpTermsNotAccepted,
			"message": "terms not accepted",
		})
		return
	}

	p := utils.Password{}
	if err := p.Create(r.Password); err != nil {
		log.Error(c, ErrorInternalServerError, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    ErrorInternalServerError,
			"message": "could not create account",
		})
		return
	}

	user := models.User{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateShortUuid(),
		},
		Email:           r.Email,
		DisplayName:     r.DisplayName,
		Username:        r.Username,
		Password:        p.Value,
		PasswordSalt:    p.Salt,
		Notifications:   8,
		AcceptedTermsAt: zero.TimeFrom(time.Now()),
		Portrait:        zero.StringFrom(r.Portrait),
		Scopes:          auth.AuthScopeUser,
	}

	var language models.Language
	if err := models.DB.Where("code=?", defaultLanguageCode(c.GetString("app_language"))).First(&language).Error; err != nil {
		log.Error(c, ErrorInternalServerError, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorInternalServerError,
			"message": "invalid language",
		})
		return
	}

	user.LanguageId = language.Id

	tx := models.DB.Begin()
	defer tx.Rollback()

	if err := tx.Create(&user).Error; err != nil {
		log.Error(c, ErrorInternalServerError, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    ErrorInternalServerError,
			"message": "could not create account",
		})
		tx.Rollback()
		return
	}

	defaultLists := config.AppConfig.DefaultLists
	for _, uuid := range defaultLists {
		var listTemplate models.ListTemplate
		if err := tx.Where("uuid=?", uuid).Preload("Name").First(&listTemplate).Error; err != nil {
			log.Error(c, ErrorInternalServerError, err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			tx.Rollback()
			return
		}

		list, err := CreateList(zero.StringFrom(uuid), models.Translation{
			Sv: listTemplate.Name.Sv,
			En: listTemplate.Name.En,
			De: listTemplate.Name.De,
			La: listTemplate.Name.La,
		}, models.VisibilityPrivate, user, zero.NewString("", false))
		if err != nil {
			log.Error(c, ErrorInternalServerError, err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			tx.Rollback()
			return
		}

		if err := tx.Create(&list).Error; err != nil {
			log.Error(c, ErrorInternalServerError, err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			tx.Rollback()
			return
		}
	}

	et, err := utils.GetEmailTemplate("welcome", language.Code, config.AppConfig.Domain, map[string]string{"DISPLAY_NAME": r.DisplayName})
	if err != nil {
		log.Error(c, ErrorInternalServerError, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	email := models.Email{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateShortUuid(),
		},
		FromEmail:  et.Email,
		ToEmail:    r.Email,
		FromName:   et.Name,
		ToName:     r.DisplayName,
		Subject:    et.Subject,
		TextPart:   et.Txt,
		HtmlPart:   et.Html,
		SendAt:     time.Now(),
		ReplyEmail: et.ReplyTo,
		ReplyName:  et.Name,
	}

	email.Checksum = utils.ComputeHmac256(email.Subject+email.HtmlPart+email.SendAt.String(), config.AppConfig.HashingSecret)
	if err := tx.Create(&email).Error; err != nil {
		log.Error(c, ErrorInternalServerError, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, user)
}

func UsersMe(c *gin.Context) {
	log.Info(c, 0, "UsersMe")

	type listReports struct {
		List    models.List     `json:"list"`
		Reports []models.Report `json:"reports"`
	}
	type response struct {
		User        models.User     `json:"user"`
		LastReports []models.Report `json:"last_reports"`
		ListReports []listReports   `json:"list_reports"`
	}

	tx := models.DB.Begin()
	defer tx.Rollback()

	var user models.User
	if err := tx.Where("uuid=?", c.GetString("uuid")).Where("deleted_at IS NULL").Preload("Language").First(&user).Error; err != nil {
		log.Error(c, 0, err.Error())
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	oneWeekAgo := time.Now().AddDate(0, 0, -14)
	var lastReports []models.Report
	if err := tx.Where("user_id=?", user.Id).Where("timestamp>=?", oneWeekAgo).Where("archived IS NULL").Preload("Tags", "archived IS NULL").Preload("Species.Name").Preload("Zone.Name").Preload("Age.Name").Preload("Gender.Name").Preload("Method.Name").Preload("Comment").Preload("Lists.Name").Order("timestamp DESC").Order("`from` DESC").Limit(50).Find(&lastReports).Error; err != nil {
		log.Error(c, 0, err.Error())
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var lists []models.List
	if err := tx.Where("user_id=?", user.Id).Where("archived IS NULL").Preload("Species.Order.Name").Preload("Name").Preload("Species.Name").Find(&lists).Error; err != nil {
		log.Error(c, 0, err.Error())
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	r := response{
		User:        user,
		LastReports: lastReports,
	}

	for _, list := range lists {
		listReport := listReports{
			List: list,
		}
		if err := tx.Table("list_reports").Select("reports.*").Where("reports.archived IS NULL").Where("list_id=?", list.Id).Joins("JOIN reports ON reports.id=report_id").Preload("Lists.Name").Preload("Species.Order.Name").Preload("Species.Name").Preload("Zone.Name").Preload("Age.Name").Preload("Gender.Name").Preload("Method.Name").Preload("Tags", "archived IS NULL").Find(&listReport.Reports).Error; err != nil {
			log.Error(c, 0, err.Error())
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}
		r.ListReports = append(r.ListReports, listReport)
	}

	tx.Commit()

	c.JSON(http.StatusOK, r)
}

func UsersGetUser(c *gin.Context) {
	log.Info(c, 0, "UsersGetUser")

	type response struct {
		User    models.User     `json:"user"`
		Reports []models.Report `json:"reports"`
	}

	var user models.User
	if err := models.DB.Where("uuid=?", c.Param("uuid")).Where("deleted_at IS NULL").First(&user).Error; err != nil {
		log.Error(c, 0, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var reports []models.Report
	if err := models.DB.Where("user_id=?", user.Id).Where("archived IS NULL").Find(&reports).Error; err != nil {
		log.Error(c, 0, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response{
		User:    user,
		Reports: reports,
	})
}

func UsersGetUsers(c *gin.Context) {
	log.Info(c, 0, "UsersGetUsers")

	var users []models.User
	if err := models.DB.Where("deleted_at IS NULL").Find(&users).Error; err != nil {
		log.Error(c, ErrorInternalServerError, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, users)
}

func UsersWhoAmI(c *gin.Context) {
	log.Info(c, 0, "UsersWhoAmI")

	var user models.User
	if err := models.DB.Preload("Language").Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		log.Error(c, 0, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GeneralSuccessResponse{
		Data:    user,
		Message: "success",
	})
}

// UsersExportData returns a copy of everything the user has stored: profile, checklists, labels and
// sightings, with names spelled out next to the ids that link them. Times are ISO 8601; a date alone
// means no time of day was recorded.
func UsersExportData(c *gin.Context) {
	log.Info(c, 0, "UsersExportData")

	type exportRef struct {
		Uuid string `json:"uuid"`
		Name string `json:"name"`
	}
	type exportSpecies struct {
		Uuid string     `json:"uuid"`
		Name exportName `json:"name"`
	}
	type exportUser struct {
		Uuid        string  `json:"uuid"`
		DisplayName string  `json:"display_name"`
		Username    string  `json:"username"`
		Email       string  `json:"email"`
		Language    string  `json:"language"`
		Portrait    *string `json:"portrait"`
		LastSeenAt  *string `json:"last_seen_at"`
	}
	type exportList struct {
		Uuid     string          `json:"uuid"`
		Name     exportName      `json:"name"`
		Color    *string         `json:"color"`
		From     *string         `json:"from"`
		Until    *string         `json:"until"`
		Archived *string         `json:"archived"`
		Species  []exportSpecies `json:"species"`
	}
	type exportLabel struct {
		Uuid     string  `json:"uuid"`
		Name     string  `json:"name"`
		Color    *string `json:"color"`
		Archived *string `json:"archived"`
	}
	type exportSighting struct {
		Uuid         string        `json:"uuid"`
		Species      exportSpecies `json:"species"`
		From         *string       `json:"from"`
		Until        *string       `json:"until"`
		Quantity     *int64        `json:"quantity"`
		Age          *exportName   `json:"age"`
		Sex          *exportName   `json:"sex"`
		Activity     *exportName   `json:"activity"`
		Method       *exportName   `json:"method"`
		Comment      *string       `json:"comment"`
		Site         *string       `json:"site"`
		Municipality *exportName   `json:"municipality"`
		Latitude     *float64      `json:"latitude"`
		Longitude    *float64      `json:"longitude"`
		Archived     *string       `json:"archived"`
		Checklists   []exportRef   `json:"checklists"`
		Labels       []exportRef   `json:"labels"`
	}
	type export struct {
		Format     string           `json:"format"`
		Version    int              `json:"version"`
		ExportedAt string           `json:"exported_at"`
		User       exportUser       `json:"user"`
		Checklists []exportList     `json:"checklists"`
		Labels     []exportLabel    `json:"labels"`
		Sightings  []exportSighting `json:"sightings"`
	}

	fail := func(err error) {
		log.Error(c, ErrorInternalServerError, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorInternalServerError,
			"message": err.Error(),
		})
	}
	optString := func(s string) *string {
		if s == "" {
			return nil
		}
		return &s
	}
	optTime := func(t zero.Time) *string {
		if !t.Valid || t.Time.IsZero() {
			return nil
		}
		s := jsonTime(t.Time)
		return &s
	}
	optName := func(t models.Translation) *exportName {
		n := newExportName(t)
		return &n
	}

	var user models.User
	if err := models.DB.Preload("Language").Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		fail(err)
		return
	}

	res := export{
		Format:     "300sf-export",
		Version:    2,
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		User: exportUser{
			Uuid:        user.Uuid,
			DisplayName: user.DisplayName,
			Username:    user.Username,
			Email:       user.Email,
			Language:    user.Language.Code,
			Portrait:    optString(user.Portrait.String),
			LastSeenAt:  optTime(user.LastSeenAt),
		},
		Checklists: []exportList{},
		Labels:     []exportLabel{},
		Sightings:  []exportSighting{},
	}

	var lists []models.List
	if err := models.DB.Preload("Name").Preload("Species.Name").Where("user_id=?", user.Id).Find(&lists).Error; err != nil {
		fail(err)
		return
	}
	for _, l := range lists {
		li := exportList{
			Uuid:     l.Uuid,
			Name:     newExportName(l.Name),
			Color:    optString(l.Color.String),
			From:     optTime(l.From),
			Until:    optTime(l.Until),
			Archived: optTime(l.Archived),
			Species:  []exportSpecies{},
		}
		for _, s := range l.Species {
			li.Species = append(li.Species, exportSpecies{Uuid: s.Uuid, Name: newExportName(s.Name)})
		}
		res.Checklists = append(res.Checklists, li)
	}

	var tags []models.Tag
	if err := models.DB.Where("user_id=?", user.Id).Find(&tags).Error; err != nil {
		fail(err)
		return
	}
	for _, t := range tags {
		res.Labels = append(res.Labels, exportLabel{
			Uuid:     t.Uuid,
			Name:     t.Name,
			Color:    optString(t.Color.String),
			Archived: optTime(t.Archived),
		})
	}

	var reports []models.Report
	if err := models.DB.Preload("Species.Name").Preload("Age.Name").Preload("Activity.Name").Preload("Gender.Name").Preload("Method.Name").Preload("Comment").Preload("Zone.Name").Preload("Lists.Name").Preload("Tags").Where("user_id=?", user.Id).Order("`from` IS NULL, `from` ASC").Find(&reports).Error; err != nil {
		fail(err)
		return
	}
	for _, r := range reports {
		s := exportSighting{
			Uuid:       r.Uuid,
			Species:    exportSpecies{Uuid: r.Species.Uuid, Name: newExportName(r.Species.Name)},
			From:       optTime(r.From),
			Until:      optTime(r.Until),
			Site:       optString(r.Site.String),
			Archived:   optTime(r.Archived),
			Checklists: []exportRef{},
			Labels:     []exportRef{},
		}
		if !r.Quantity.IsZero() {
			q := r.Quantity.ValueOrZero()
			s.Quantity = &q
		}
		if r.Age != nil {
			s.Age = optName(r.Age.Name)
		}
		if r.Gender != nil {
			s.Sex = optName(r.Gender.Name)
		}
		if r.Activity != nil {
			s.Activity = optName(r.Activity.Name)
		}
		if r.Method != nil {
			s.Method = optName(r.Method.Name)
		}
		if r.Comment != nil {
			s.Comment = optString(r.Comment.Text)
		}
		if r.Zone != nil {
			s.Municipality = optName(r.Zone.Name)
		}
		if !r.Latitude.IsZero() && !r.Longitude.IsZero() {
			lat, lng := r.Latitude.Float64, r.Longitude.Float64
			s.Latitude, s.Longitude = &lat, &lng
		}
		for _, l := range r.Lists {
			s.Checklists = append(s.Checklists, exportRef{Uuid: l.Uuid, Name: translated(l.Name, user.Language.Code)})
		}
		for _, t := range r.Tags {
			s.Labels = append(s.Labels, exportRef{Uuid: t.Uuid, Name: t.Name})
		}
		res.Sightings = append(res.Sightings, s)
	}

	c.JSON(http.StatusOK, res)
}

func UsersMigrate(c *gin.Context) {
	log.Info(c, 0, "UsersMigrate")

	type request struct {
		OldUsername string `json:"old_username"`
		OldPassword string `json:"old_password"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		log.Error(c, -10001, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	// Each attempt tries a password on the old site; don't let this be used
	// to guess them.
	if throttle(c, "migrate/"+c.GetString("uuid"), 10, time.Hour) {
		return
	}

	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		log.Error(c, -10001, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	res, code, err := migrate(user, r.OldUsername, r.OldPassword)
	if err != nil {
		log.Error(c, int64(code), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    code,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res)
}

func UsersLanguage(c *gin.Context) {
	log.Info(c, 0, "UsersLanguage")

	type request struct {
		LanguageCode string `json:"language_code"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		log.Error(c, ErrorParse, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "could not parse request",
			"code":    ErrorParse,
		})
		return
	}

	var language models.Language
	if err := models.DB.Where("code=?", r.LanguageCode).First(&language).Error; err != nil {
		log.Error(c, ErrorRecordNotFound, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "could not find species",
			"code":    ErrorRecordNotFound,
		})
		return
	}

	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		log.Error(c, -10001, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	models.DB.Table(table.Users).Where("id=?", user.Id).UpdateColumn("language_id", language.Id)

	c.JSON(http.StatusNoContent, nil)
}

func UsersPortrait(c *gin.Context) {
	log.Info(c, 0, "UsersPortrait")

	type request struct {
		Portrait string `json:"portrait"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		log.Error(c, ErrorParse, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "could not parse request",
			"code":    ErrorParse,
		})
		return
	}

	if !utils.VerifyPortrait(r.Portrait) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid portrait",
			"code":    ErrorParse,
		})
		return
	}

	models.DB.Table(table.Users).Where("uuid=?", c.GetString("uuid")).UpdateColumn("portrait", r.Portrait)

	c.JSON(http.StatusNoContent, nil)
}

// defaultLanguageCode maps a client supplied language onto one we ship email
// templates for, falling back to English.
func defaultLanguageCode(code string) string {
	switch code {
	case "en", "sv", "de":
		return code
	default:
		return "en"
	}
}

func UsersResetPassword(c *gin.Context) {
	log.Info(c, 0, "UsersResetPassword")

	type request struct {
		Email string `json:"email"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		log.Error(c, ErrorParse, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "could not parse request",
			"code":    ErrorParse,
		})
		return
	}

	if throttle(c, "reset-password/"+c.ClientIP(), 10, time.Hour) {
		return
	}

	emailAddress := strings.TrimSpace(r.Email)

	// The answer is the same whether or not there is an account, so this
	// can't be used to find out who has one.
	var user models.User
	if err := models.DB.Preload("Language").Where("email=?", emailAddress).First(&user).Error; err != nil {
		log.Info(c, 0, "password reset for unknown email")
		c.JSON(http.StatusNoContent, nil)
		return
	}

	// One reset email per account every two minutes, so the endpoint can't be
	// used to flood someone's inbox.
	var recent int64
	models.DB.Model(&models.ChangePasswordRequest{}).Where("user_id=?", user.Id).Where("timestamp>?", time.Now().Add(-2*time.Minute)).Count(&recent)
	if recent > 0 {
		log.Info(c, 0, user.Uuid+" password reset throttled")
		c.JSON(http.StatusNoContent, nil)
		return
	}

	token, tokenHash := utils.NewToken()
	cpr := models.ChangePasswordRequest{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateShortUuid(),
		},
		ExpireAt: time.Now().Add(24 * time.Hour),
		Token:    tokenHash,
		UserId:   user.Id,
	}

	// Users predating the language column have no language set, fall back to the
	// language of the requesting client rather than looking up an empty template.
	languageCode := user.Language.Code
	if languageCode == "" {
		languageCode = defaultLanguageCode(c.GetString("app_language"))
	}

	resetLink := fmt.Sprintf("%s/set-password?token=%s", config.AppConfig.Domain, token)
	et, err := utils.GetEmailTemplate("set_password", languageCode, config.AppConfig.Domain, map[string]string{"LINK": resetLink})
	if err != nil {
		log.Error(c, 0, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	email := models.Email{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateShortUuid(),
		},
		FromEmail:  et.Email,
		ToEmail:    user.Email,
		FromName:   et.Name,
		ToName:     user.DisplayName,
		Subject:    et.Subject,
		TextPart:   et.Txt,
		HtmlPart:   et.Html,
		SendAt:     time.Now(),
		ReplyEmail: et.ReplyTo,
		ReplyName:  et.Name,
	}

	tx := models.DB.Begin()
	defer tx.Rollback()

	email.Checksum = utils.ComputeHmac256(email.Subject+email.HtmlPart+email.SendAt.String(), config.AppConfig.HashingSecret)
	if err := tx.Create(&email).Error; err != nil {
		log.Error(c, 0, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	if err := tx.Create(&cpr).Error; err != nil {
		log.Error(c, 0, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	tx.Commit()

	log.Info(c, 0, user.Uuid+" requested a password reset")

	c.JSON(http.StatusNoContent, nil)
}

func UsersSetPassword(c *gin.Context) {
	log.Info(c, 0, "UsersSetPassword")

	type request struct {
		Password string `json:"password"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		log.Error(c, ErrorInternalServerError, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "could not parse request",
			"code":    ErrorParse,
		})
		return
	}

	tx := models.DB.Begin()
	defer tx.Rollback()

	var cpr models.ChangePasswordRequest
	if err := tx.Where("token=?", utils.HashToken(c.Param("token"))).First(&cpr).Error; err != nil {
		log.Error(c, ErrorRecordNotFound, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorRecordNotFound,
			"message": "unknown or already used reset link",
		})
		tx.Rollback()
		return
	}

	if cpr.ExpireAt.Before(time.Now()) {
		log.Error(c, ErrorLinkExpired, "reset link expired, issued "+cpr.Timestamp.String())
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorLinkExpired,
			"message": "link expired",
		})
		tx.Rollback()
		return
	}

	var user models.User
	if err := tx.Where("id=?", cpr.UserId).First(&user).Error; err != nil {
		log.Error(c, ErrorInternalServerError, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	if !utils.VerifyPassword(r.Password) {
		log.Error(c, ErrorSignUpInvalidPassword, "rejected new password for "+user.Uuid)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorSignUpInvalidPassword,
			"message": "invalid password",
		})
		tx.Rollback()
		return
	}

	p := utils.Password{}
	if err := p.Create(r.Password); err != nil {
		log.Error(c, ErrorInternalServerError, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorInternalServerError,
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	if err := tx.Table(table.Users).Where("id=?", user.Id).UpdateColumn("password", p.Value).UpdateColumn("password_salt", p.Salt).Error; err != nil {
		log.Error(c, ErrorInternalServerError, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	// Use up this link and any other reset link sent to the user.
	if err := tx.Table(table.ChangePasswordRequests).Where("user_id=?", user.Id).Where("expire_at>?", time.Now()).UpdateColumn("expire_at", time.Now()).Error; err != nil {
		log.Error(c, ErrorInternalServerError, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	tx.Commit()

	// Whoever knew the old password may still be signed in.
	RevokeUserSessions(user.Uuid)

	c.JSON(http.StatusNoContent, nil)
}

func UsersRequestDelete(c *gin.Context) {
	type request struct {
		Password string `json:"password"`
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

	var user models.User
	if err := tx.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	var language models.Language
	if err := tx.Where("id=?", user.LanguageId).First(&language).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorSignUpTermsNotAccepted,
			"message": "invalid language",
		})
		tx.Rollback()
		return
	}

	if PasswordLocked(user.Email) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"code":    ErrorTooManyLoginAttempts,
			"message": "too many wrong passwords, try again later",
		})
		return
	}
	p := utils.Password{}
	if !p.IsMatching(r.Password, user.Password, user.PasswordSalt) {
		PasswordFailed(user.Email)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -30002,
			"message": "invalid password",
		})
		tx.Rollback()
		return
	}

	// Each request sends an email.
	if throttle(c, "request-delete/"+user.Uuid, 3, time.Hour) {
		return
	}

	token, tokenHash := utils.NewToken()
	dr := models.DeleteAccountRequest{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateShortUuid(),
		},
		ExpireAt: time.Now().Add(1 * time.Hour),
		Token:    tokenHash,
		UserId:   user.Id,
	}

	deleteLink := fmt.Sprintf("%s/delete-account?token=%s", config.AppConfig.Domain, token)
	et, err := utils.GetEmailTemplate("delete_account", language.Code, config.AppConfig.Domain, map[string]string{"LINK": deleteLink})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	email := models.Email{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateShortUuid(),
		},
		FromEmail:  et.Email,
		ToEmail:    user.Email,
		FromName:   et.Name,
		ToName:     user.DisplayName,
		Subject:    et.Subject,
		TextPart:   et.Txt,
		HtmlPart:   et.Html,
		SendAt:     time.Now(),
		ReplyEmail: et.ReplyTo,
		ReplyName:  et.Name,
	}

	email.Checksum = utils.ComputeHmac256(email.Subject+email.HtmlPart+email.SendAt.String(), config.AppConfig.HashingSecret)
	if err := tx.Create(&email).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	if err := tx.Create(&dr).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	tx.Commit()
	c.JSON(http.StatusNoContent, nil)
}

func UsersConfirmDelete(c *gin.Context) {
	type request struct {
		Password string `json:"password"`
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

	var user models.User
	if err := tx.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	var language models.Language
	if err := tx.Where("id=?", user.LanguageId).First(&language).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorSignUpTermsNotAccepted,
			"message": "invalid language",
		})
		tx.Rollback()
		return
	}

	if PasswordLocked(user.Email) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"code":    ErrorTooManyLoginAttempts,
			"message": "too many wrong passwords, try again later",
		})
		return
	}
	p := utils.Password{}
	if !p.IsMatching(r.Password, user.Password, user.PasswordSalt) {
		PasswordFailed(user.Email)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -30002,
			"message": "invalid password",
		})
		tx.Rollback()
		return
	}

	var dr models.DeleteAccountRequest
	if err := tx.Table(table.DeleteAccountRequests).Where("token=?", utils.HashToken(c.Param("token"))).Where("user_id=?", user.Id).Where("delete_at IS NULL").First(&dr).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorRecordNotFound,
			"message": "not found",
		})
		tx.Rollback()
		return
	}

	if dr.ExpireAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorLinkExpired,
			"message": "link expired",
		})
		tx.Rollback()
		return
	}

	if err := tx.Table(table.DeleteAccountRequests).Where("id=?", dr.Id).UpdateColumn("delete_at", time.Now().Add(time.Hour*168)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    ErrorRecordNotFound,
			"message": "link expired",
		})
		tx.Rollback()
		return
	}

	if err := tx.Table(table.Users).Where("id=?", user.Id).UpdateColumn("deactivated_at", time.Now()).UpdateColumn("delete_account_request_id", dr.Id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    ErrorRecordNotFound,
			"message": "link expired",
		})
		tx.Rollback()
		return
	}

	tx.Commit()

	RevokeUserSessions(user.Uuid)

	c.JSON(http.StatusNoContent, nil)
}
