// SPDX-License-Identifier: GPL-2.0-or-later
package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/guregu/null.v4/zero"
	"gorm.io/gorm/clause"

	"git.sr.ht/~klahr/300sf/auth"
	"git.sr.ht/~klahr/300sf/config"
	"git.sr.ht/~klahr/300sf/log"
	"git.sr.ht/~klahr/300sf/models"
	"git.sr.ht/~klahr/300sf/table"
	"git.sr.ht/~klahr/300sf/utils"
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
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorInternalServerError,
			"message": err.Error(),
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
	if err := models.DB.Where("code=?", c.GetString("app_language")).First(&language).Error; err != nil {
		log.Error(c, ErrorInternalServerError, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorInternalServerError,
			"message": "invalid language",
		})
		return
	}

	user.LanguageId = language.Id

	tx := models.DB.Begin()

	if err := tx.Create(&user).Error; err != nil {
		log.Error(c, ErrorInternalServerError, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
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

	et, err := utils.GetEmailTemplate("welcome", language.Code)
	if err != nil {
		log.Error(c, ErrorInternalServerError, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	et.Txt = strings.ReplaceAll(et.Txt, "${DISPLAY_NAME}", r.DisplayName)
	et.Html = strings.ReplaceAll(et.Html, "${DISPLAY_NAME}", r.DisplayName)

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
		ReplyEmail: et.Email,
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

func UsersExportData(c *gin.Context) {
	log.Info(c, 0, "UsersExportData")

	type exportUser struct {
		Uuid          string      `json:"uuid"`
		DisplayName   string      `json:"display_name"`
		Email         string      `json:"email"`
		Scopes        string      `json:"scopes"`
		Notifications int64       `json:"notifications"`
		LastSeenAt    zero.Time   `json:"last_seen_at"`
		Language      string      `json:"language"`
		Portrait      zero.String `json:"portrait"`
	}
	type exportReport struct {
		Uuid      string             `json:"uuid"`
		Species   models.Translation `json:"species"`
		From      string             `json:"from"`
		Until     string             `json:"until"`
		Latitude  zero.Float         `json:"latitude"`
		Longitude zero.Float         `json:"longitude"`
		Quantity  zero.Int           `json:"quantity"`
		Age       models.Translation `json:"age"`
		Activity  models.Translation `json:"activity"`
		Gender    models.Translation `json:"gender"`
		Method    models.Translation `json:"method"`
		Comment   string             `json:"comment"`
		Archived  string             `json:"archived"`
		Lists     []string           `json:"lists"`
		Tags      []string           `json:"tags"`
	}
	type exportList struct {
		Uuid       string             `json:"uuid"`
		Name       models.Translation `json:"name"`
		From       string             `json:"from"`
		Until      string             `json:"until"`
		Visibility int64              `json:"visibility"`
		Archived   string             `json:"archived"`
		Species    []string           `json:"species"`
		Color      zero.String        `json:"color"`
	}
	type exportedUserData struct {
		User    exportUser      `json:"user"`
		Reports []exportReport  `json:"reports"`
		ApiKeys []models.ApiKey `json:"api_keys"`
		Lists   []exportList    `json:"lists"`
		Tags    []models.Tag    `json:"tags"`
	}
	var res exportedUserData

	var user models.User
	if err := models.DB.Preload(clause.Associations).Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		log.Error(c, ErrorInternalServerError, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorInternalServerError,
			"message": err.Error(),
		})
		return
	}

	res.User = exportUser{
		Uuid:          user.Uuid,
		DisplayName:   user.DisplayName,
		Email:         user.Email,
		Scopes:        user.Scopes,
		Notifications: user.Notifications,
		LastSeenAt:    user.LastSeenAt,
		Language:      user.Language.Code,
		Portrait:      user.Portrait,
	}

	var reports []models.Report
	if err := models.DB.Preload(clause.Associations).Preload("Species.Name").Preload("Age.Name").Preload("Activity.Name").Preload("Gender.Name").Preload("Method.Name").Where("user_id=?", user.Id).Find(&reports).Error; err != nil {
		log.Error(c, ErrorInternalServerError, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorInternalServerError,
			"message": err.Error(),
		})
		return
	}

	for _, r := range reports {
		rep := exportReport{
			Uuid:      r.Uuid,
			Species:   r.Species.Name,
			Latitude:  r.Latitude,
			Longitude: r.Longitude,
			Quantity:  r.Quantity,
			Comment:   r.Comment.Text,
		}

		if r.From.Valid {
			rep.From = r.From.Time.String()
		}
		if r.Until.Valid {
			rep.Until = r.Until.Time.String()
		}
		if r.Age != nil {
			rep.Age = r.Age.Name
		}
		if r.Activity != nil {
			rep.Activity = r.Activity.Name
		}
		if r.Gender != nil {
			rep.Gender = r.Gender.Name
		}
		if r.Method != nil {
			rep.Method = r.Method.Name
		}
		for _, l := range r.Lists {
			rep.Lists = append(rep.Lists, l.Uuid)
		}
		for _, t := range r.Tags {
			rep.Tags = append(rep.Tags, t.Uuid)
		}

		res.Reports = append(res.Reports, rep)
	}

	if err := models.DB.Preload(clause.Associations).Where("user_id=?", user.Id).Find(&res.ApiKeys).Error; err != nil {
		log.Error(c, ErrorInternalServerError, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorInternalServerError,
			"message": err.Error(),
		})
		return
	}

	var lists []models.List
	if err := models.DB.Preload(clause.Associations).Where("user_id=?", user.Id).Find(&lists).Error; err != nil {
		log.Error(c, ErrorInternalServerError, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorInternalServerError,
			"message": err.Error(),
		})
		return
	}

	for _, l := range lists {
		li := exportList{
			Uuid:       l.Uuid,
			Name:       l.Name,
			Visibility: l.Visibility,
			Color:      l.Color,
		}

		if l.From.Valid {
			li.From = l.From.Time.String()
		}
		if l.Until.Valid {
			li.Until = l.Until.Time.String()
		}
		if l.Archived.Valid {
			li.Archived = l.Archived.Time.String()
		}

		for _, s := range l.Species {
			li.Species = append(li.Species, s.Uuid)
		}

		res.Lists = append(res.Lists, li)
	}

	if err := models.DB.Preload(clause.Associations).Where("user_id=?", user.Id).Find(&res.Tags).Error; err != nil {
		log.Error(c, ErrorInternalServerError, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorInternalServerError,
			"message": err.Error(),
		})
		return
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

	// TODO: This supports anything, which is not OK.
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

	emailAddress := strings.TrimSpace(r.Email)

	var user models.User
	if err := models.DB.Preload("Language").Where("email=?", emailAddress).First(&user).Error; err != nil {
		log.Error(c, ErrorRecordNotFound, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorRecordNotFound,
			"message": "no account with that email address",
		})
		return
	}

	cpr := models.ChangePasswordRequest{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateShortUuid(),
		},
		ExpireAt: time.Now().Add(24 * time.Hour),
		Token:    utils.ComputeHmac256(user.Email+time.Now().String(), config.AppConfig.HashingSecret),
		UserId:   user.Id,
	}

	// Users predating the language column have no language set, fall back to the
	// language of the requesting client rather than looking up an empty template.
	languageCode := user.Language.Code
	if languageCode == "" {
		languageCode = defaultLanguageCode(c.GetString("app_language"))
	}

	et, err := utils.GetEmailTemplate("set_password", languageCode)
	if err != nil {
		log.Error(c, 0, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	resetLink := fmt.Sprintf("%s/set-password?token=%s", config.AppConfig.Domain, cpr.Token)
	et.Txt = strings.ReplaceAll(et.Txt, "${LINK}", resetLink)
	et.Html = strings.ReplaceAll(et.Html, "${LINK}", resetLink)

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
		ReplyEmail: et.Email,
		ReplyName:  et.Name,
	}

	tx := models.DB.Begin()

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

	var cpr models.ChangePasswordRequest
	if err := tx.Where("token=?", c.Param("token")).First(&cpr).Error; err != nil {
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

	if err := tx.Table(table.ChangePasswordRequests).Where("id=?", cpr.Id).UpdateColumn("expire_at", time.Now()).Error; err != nil {
		log.Error(c, ErrorInternalServerError, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	tx.Commit()

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

	p := utils.Password{}
	if !p.IsMatching(r.Password, user.Password, user.PasswordSalt) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -30002,
			"message": "invalid password",
		})
		tx.Rollback()
		return
	}

	dr := models.DeleteAccountRequest{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateShortUuid(),
		},
		ExpireAt: time.Now().Add(1 * time.Hour),
		Token:    utils.ComputeHmac256(user.Email+time.Now().String(), config.AppConfig.HashingSecret),
		UserId:   user.Id,
	}

	et, err := utils.GetEmailTemplate("delete_account", language.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	deleteLink := fmt.Sprintf("%s/delete-account?token=%s", config.AppConfig.Domain, dr.Token)
	et.Txt = strings.ReplaceAll(et.Txt, "${LINK}", deleteLink)
	et.Html = strings.ReplaceAll(et.Html, "${LINK}", deleteLink)

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
		ReplyEmail: et.Email,
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

	p := utils.Password{}
	if !p.IsMatching(r.Password, user.Password, user.PasswordSalt) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -30002,
			"message": "invalid password",
		})
		tx.Rollback()
		return
	}

	var dr models.DeleteAccountRequest
	if err := tx.Table(table.DeleteAccountRequests).Where("token=?", c.Param("token")).Where("delete_at IS NULL").First(&dr).Error; err != nil {
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

	c.JSON(http.StatusNoContent, nil)
}
