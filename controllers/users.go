package controllers

import (
	"net/http"
	"time"

	"git.sr.ht/~klahr/300sf/models"
	"git.sr.ht/~klahr/300sf/utils"
	"github.com/gin-gonic/gin"
	"gopkg.in/guregu/null.v4/zero"
)

func User(c *gin.Context) {
	type request struct {
		Email         string `json:"email"`
		DisplayName   string `json:"display_name"`
		Password      string `json:"password"`
		IconUuid      string `json:"icon_uuid"`
		AcceptedTerms bool   `json:"accepted_terms"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -10001,
			"message": err.Error(),
		})
		return
	}

	if r.AcceptedTerms == false {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -20001,
			"message": "terms not accepted",
		})
		return
	}

	if !utils.VerifyEmail(r.Email) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -20002,
			"message": "invalid email",
		})
		return
	}

	if !utils.VerifyDisplayName(r.DisplayName) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -20003,
			"message": "invalid display name",
		})
		return
	}

	if len(r.Password) < 4 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -20004,
			"message": "invalid password",
		})
		return
	}

	p := utils.Password{}
	p.Create(r.Password)

	user := models.User{
		Model: models.Model{
			Timestamp: time.Now(),
			Uuid:      utils.GenerateUuid(),
		},
		Email:           r.Email,
		DisplayName:     r.DisplayName,
		Password:        p.Value,
		PasswordSalt:    p.Salt,
		Permissions:     1,
		Notifications:   8,
		AcceptedTermsAt: zero.TimeFrom(time.Now()),
	}

	if err := models.DB.Omit("icon_id").Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}
