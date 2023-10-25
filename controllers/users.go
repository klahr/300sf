// SPDX-License-Identifier: GPL-2.0-or-later
package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"git.sr.ht/~klahr/300sf/config"
	"git.sr.ht/~klahr/300sf/models"
	"git.sr.ht/~klahr/300sf/table"
	"git.sr.ht/~klahr/300sf/utils"
	"github.com/gin-gonic/gin"
	"gopkg.in/guregu/null.v4/zero"
)

func Users(c *gin.Context) {
	type request struct {
		Email         string `json:"email"`
		DisplayName   string `json:"display_name"`
		Password      string `json:"password"`
		AcceptedTerms bool   `json:"accepted_terms"`
		Portrait      string `json:"portrait"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	if !utils.VerifyDisplayName(r.DisplayName) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorSignUpInvalidDisplayName,
			"message": "invalid display name",
		})
		return
	}

	if !utils.VerifyEmail(r.Email) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorSignUpInvalidEmail,
			"message": "invalid email",
		})
		return
	}

	if !utils.VerifyPassword(r.Password) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorSignUpInvalidPassword,
			"message": "invalid password",
		})
		return
	}

	if !r.AcceptedTerms {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorSignUpTermsNotAccepted,
			"message": "terms not accepted",
		})
		return
	}

	p := utils.Password{}
	p.Create(r.Password)

	user := models.User{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateShortUuid(),
		},
		Email:           r.Email,
		DisplayName:     r.DisplayName,
		Password:        p.Value,
		PasswordSalt:    p.Salt,
		Permissions:     1,
		Notifications:   8,
		AcceptedTermsAt: zero.TimeFrom(time.Now()),
		Portrait:        zero.StringFrom(r.Portrait),
	}

	var language models.Language
	if err := models.DB.Where("code=?", c.GetString("app_language")).First(&language).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorSignUpTermsNotAccepted,
			"message": "invalid language",
		})
		return
	}

	user.LanguageId = language.Id

	tx := models.DB.Begin()

	if err := tx.Create(&user).Error; err != nil {
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
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			tx.Rollback()
			return
		}

		if err := tx.Create(&list).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			tx.Rollback()
			return
		}
	}

	et, err := utils.GetEmailTemplate("welcome", language.Code)
	if err != nil {
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
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	oneWeekAgo := time.Now().AddDate(0, 0, -14)
	var lastReports []models.Report
	if err := tx.Where("user_id=?", user.Id).Where("timestamp>=?", oneWeekAgo).Where("archived IS NULL").Preload("Tags").Preload("Species.Name").Preload("Zone.Name").Preload("Age.Name").Preload("Gender.Name").Preload("Method.Name").Preload("Comment").Preload("Lists.Name").Order("timestamp DESC").Order("`from` DESC").Limit(50).Find(&lastReports).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var lists []models.List
	if err := tx.Where("user_id=?", user.Id).Where("archived IS NULL").Preload("Species.Order.Name").Preload("Name").Preload("Species.Name").Find(&lists).Error; err != nil {
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
		if err := tx.Table("list_reports").Select("reports.*").Where("reports.archived IS NULL").Where("list_id=?", list.Id).Joins("JOIN reports ON reports.id=report_id").Preload("Lists.Name").Preload("Species.Order.Name").Preload("Species.Name").Preload("Zone.Name").Preload("Age.Name").Preload("Gender.Name").Preload("Method.Name").Preload("Tags").Find(&listReport.Reports).Error; err != nil {
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
	type response struct {
		User    models.User     `json:"user"`
		Reports []models.Report `json:"reports"`
	}

	var user models.User
	if err := models.DB.Where("uuid=?", c.Param("uuid")).Where("deleted_at IS NULL").First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var reports []models.Report
	if err := models.DB.Where("user_id=?", user.Id).Where("archived IS NULL").Find(&reports).Error; err != nil {
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
	var users []models.User
	if err := models.DB.Where("deleted_at IS NULL").Find(&users).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, users)
}

func UsersWhoAmI(c *gin.Context) {
	var user models.User
	if err := models.DB.Preload("Language").Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
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
	c.JSON(http.StatusTeapot, nil)
}

func UsersMigrate(c *gin.Context) {
	type request struct {
		OldUsername string `json:"old_username"`
		OldPassword string `json:"old_password"`
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

	res, code, err := migrate(user, r.OldUsername, r.OldPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    code,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res)
}

func UsersLanguage(c *gin.Context) {
	type request struct {
		LanguageCode string `json:"language_code"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "could not parse request",
			"code":    ErrorParse,
		})
		return
	}

	var language models.Language
	if err := models.DB.Where("code=?", r.LanguageCode).First(&language).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "could not find species",
			"code":    ErrorRecordNotFound,
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

	models.DB.Table(table.Users).Where("id=?", user.Id).UpdateColumn("language_id", language.Id)

	c.JSON(http.StatusNoContent, nil)
}

func UsersPortrait(c *gin.Context) {
	type request struct {
		Portrait string `json:"portrait"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
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

func UsersResetPassword(c *gin.Context) {
	type request struct {
		Email string `json:"email"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "could not parse request",
			"code":    ErrorParse,
		})
		return
	}

	var user models.User
	if err := models.DB.Where("email=?", r.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	var language models.Language
	if err := models.DB.Where("code=?", c.GetString("app_language")).First(&language).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorSignUpTermsNotAccepted,
			"message": "invalid language",
		})
		return
	}

	cpr := models.ChangePasswordRequest{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateShortUuid(),
		},
		ExpireAt: time.Now().Add(1 * time.Hour),
		Token:    utils.ComputeHmac256(user.Email+time.Now().String(), config.AppConfig.HashingSecret),
		UserId:   user.Id,
	}

	et, err := utils.GetEmailTemplate("set_password", language.Code)
	if err != nil {
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	if err := tx.Create(&cpr).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	tx.Commit()

	c.JSON(http.StatusNoContent, nil)
}

func UsersSetPassword(c *gin.Context) {
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

	var cpr models.ChangePasswordRequest
	if err := tx.Where("token=?", c.Param("token")).First(&cpr).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	if cpr.ExpireAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorLinkExpired,
			"message": "link expired",
		})
		tx.Rollback()
		return
	}

	var user models.User
	if err := tx.Where("id=?", cpr.UserId).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	if !utils.VerifyPassword(r.Password) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrorSignUpInvalidPassword,
			"message": "invalid password",
		})
		return
	}

	p := utils.Password{}
	p.Create(r.Password)

	if err := tx.Table(table.Users).Where("id=?", user.Id).UpdateColumn("password", p.Value).UpdateColumn("password_salt", p.Salt).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		tx.Rollback()
		return
	}

	if err := tx.Table(table.ChangePasswordRequests).Where("id=?", cpr.Id).UpdateColumn("expire_at", time.Now()).Error; err != nil {
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
