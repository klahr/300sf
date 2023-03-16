package controllers

import (
	"net/http"
	"time"

	"git.sr.ht/~klahr/300sf/config"
	"git.sr.ht/~klahr/300sf/models"
	"git.sr.ht/~klahr/300sf/utils"
	"github.com/gin-gonic/gin"
	"gopkg.in/guregu/null.v4/zero"
)

func Users(c *gin.Context) {
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

	if !r.AcceptedTerms {
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
			Uuid:      utils.GenerateShortUuid(),
		},
		Email:           r.Email,
		DisplayName:     r.DisplayName,
		Password:        p.Value,
		PasswordSalt:    p.Salt,
		Permissions:     1,
		Notifications:   8,
		AcceptedTermsAt: zero.TimeFrom(time.Now()),
	}

	tx := models.DB.Begin()

	if err := tx.Omit("icon_id").Create(&user).Error; err != nil {
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
			La: listTemplate.Name.La,
		}, models.VisibilityPrivate, user)
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

	tx.Commit()

	c.JSON(http.StatusOK, user)
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

	for i := range reports {
		reports[i].Comment = zero.NewString("", false)
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
	type response struct {
		User    models.User     `json:"user"`
		Reports []models.Report `json:"reports"`
	}

	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
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

	res, _ := migrate(user, r.OldUsername, r.OldPassword)

	c.JSON(http.StatusOK, res)
}
