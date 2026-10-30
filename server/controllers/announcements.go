// SPDX-License-Identifier: AGPL-3.0-or-later
package controllers

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/table"
	"github.com/klahr/300sf/server/utils"
	"gopkg.in/guregu/null.v4/zero"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Announcements are the posts of the news feed. Admins write them; everyone
// signed in reads them in the feed, and the ones they haven't read yet are
// also shown on the home page.

const (
	maxNewsTitle   = 160
	maxNewsBody    = 10000
	newsPageSize   = 20
	maxNewsPerPage = 50
)

var (
	newsIcon  = regexp.MustCompile(`^[a-z0-9_]{1,32}$`)
	newsColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
)

// publishedAt orders the feed: when the post went out, or was made.
const publishedAt = "COALESCE(announcements.`from`, announcements.timestamp)"

// published limits a query to the posts readers see now.
func published(db *gorm.DB, now time.Time) *gorm.DB {
	return db.Where("announcements.on_hold IS NULL").
		Where("announcements.`from` IS NULL OR announcements.`from`<=?", now).
		Where("announcements.until IS NULL OR announcements.until>?", now)
}

// NewsPost is a post as a reader sees it.
type NewsPost struct {
	Uuid        string               `json:"uuid"`
	Title       models.LocalizedText `json:"title"`
	Body        models.LocalizedText `json:"body"`
	Icon        string               `json:"icon"`
	Color       string               `json:"color"`
	PublishedAt time.Time            `json:"published_at"`
	Read        bool                 `json:"read"`
}

func newsPost(a models.Announcement, read bool) NewsPost {
	at := a.Timestamp
	if a.From.Valid {
		at = a.From.Time
	}
	return NewsPost{Uuid: a.Uuid, Title: a.Title, Body: a.Body, Icon: a.Icon, Color: a.Color, PublishedAt: at, Read: read}
}

// readIds are the posts among announcements that the user has read. Posts
// from before the user signed up count as read, so a new account doesn't
// start with the whole history unread.
func readIds(user *models.User, announcements []models.Announcement) (map[int64]bool, error) {
	read := map[int64]bool{}
	if len(announcements) == 0 {
		return read, nil
	}
	ids := make([]int64, len(announcements))
	for i, a := range announcements {
		ids[i] = a.Id
		if newsPost(a, false).PublishedAt.Before(user.Timestamp) {
			read[a.Id] = true
		}
	}
	var readIds []int64
	if err := models.DB.Table("announcement_users").Where("user_id=? AND announcement_id IN ?", user.Id, ids).Pluck("announcement_id", &readIds).Error; err != nil {
		return nil, err
	}
	for _, id := range readIds {
		read[id] = true
	}
	return read, nil
}

// AnnouncementsFeed is the news feed, newest first, a page at a time:
// ?before=<uuid of the last post seen> gives the next page.
func AnnouncementsFeed(c *gin.Context) {
	user, err := currentUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Message: err.Error(), Code: ErrorRecordNotFound})
		return
	}
	limit := newsPageSize
	if n, err := strconv.Atoi(c.Query("limit")); err == nil && n > 0 {
		limit = min(n, maxNewsPerPage)
	}

	now := time.Now()
	q := published(models.DB.Model(&models.Announcement{}), now)
	if before := c.Query("before"); before != "" {
		var last models.Announcement
		if err := models.DB.Where("uuid=?", before).First(&last).Error; err != nil {
			c.JSON(http.StatusBadRequest, GeneralErrorResponse{Message: "unknown post", Code: ErrorRecordNotFound})
			return
		}
		at := newsPost(last, false).PublishedAt
		q = q.Where(publishedAt+"<? OR ("+publishedAt+"=? AND announcements.id<?)", at, at, last.Id)
	}
	var announcements []models.Announcement
	if err := q.Order(publishedAt + " DESC").Order("announcements.id DESC").Limit(limit + 1).Find(&announcements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, GeneralErrorResponse{Message: err.Error(), Code: ErrorUnknown})
		return
	}
	more := len(announcements) > limit
	if more {
		announcements = announcements[:limit]
	}
	read, err := readIds(user, announcements)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GeneralErrorResponse{Message: err.Error(), Code: ErrorUnknown})
		return
	}
	posts := []NewsPost{}
	for _, a := range announcements {
		posts = append(posts, newsPost(a, read[a.Id]))
	}
	c.JSON(http.StatusOK, gin.H{"data": posts, "more": more})
}

// AnnouncementsList is the published posts the user hasn't read, newest
// first: the home page shows them, and the menu counts them.
func AnnouncementsList(c *gin.Context) {
	user, err := currentUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Message: err.Error(), Code: ErrorRecordNotFound})
		return
	}
	var announcements []models.Announcement
	if err := published(models.DB.Model(&models.Announcement{}), time.Now()).
		Where(publishedAt+">=?", user.Timestamp).
		Joins("LEFT JOIN announcement_users ON announcement_users.announcement_id=announcements.id AND announcement_users.user_id=?", user.Id).
		Where("announcement_users.announcement_id IS NULL").
		Order(publishedAt + " DESC").Order("announcements.id DESC").
		Limit(maxNewsPerPage).
		Find(&announcements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, GeneralErrorResponse{Message: err.Error(), Code: ErrorUnknown})
		return
	}
	posts := []NewsPost{}
	for _, a := range announcements {
		posts = append(posts, newsPost(a, false))
	}
	c.JSON(http.StatusOK, GeneralSuccessResponse{Message: "success", Data: posts})
}

// AnnouncementsRead marks a post read. Reading it again changes nothing.
func AnnouncementsRead(c *gin.Context) {
	user, err := currentUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Message: err.Error(), Code: ErrorRecordNotFound})
		return
	}
	var announcement models.Announcement
	if err := models.DB.Where("uuid=?", c.Param("uuid")).First(&announcement).Error; err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Message: err.Error(), Code: ErrorRecordNotFound})
		return
	}
	if err := models.DB.Clauses(clause.Insert{Modifier: "IGNORE"}).Create(&models.AnnouncementUser{
		AnnouncementId: announcement.Id,
		UserId:         user.Id,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, GeneralErrorResponse{Message: err.Error(), Code: ErrorInsert})
		return
	}
	c.JSON(http.StatusOK, GeneralSuccessResponse{Message: "success"})
}

// AnnouncementsReadAll marks every published post read.
func AnnouncementsReadAll(c *gin.Context) {
	user, err := currentUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Message: err.Error(), Code: ErrorRecordNotFound})
		return
	}
	now := time.Now()
	if err := models.DB.Exec("INSERT IGNORE INTO announcement_users (announcement_id, user_id) "+
		"SELECT announcements.id, ? FROM announcements WHERE announcements.on_hold IS NULL "+
		"AND (announcements.`from` IS NULL OR announcements.`from`<=?) AND (announcements.until IS NULL OR announcements.until>?)",
		user.Id, now, now).Error; err != nil {
		c.JSON(http.StatusInternalServerError, GeneralErrorResponse{Message: err.Error(), Code: ErrorInsert})
		return
	}
	c.JSON(http.StatusOK, GeneralSuccessResponse{Message: "success"})
}

// newsRequest is a post as an admin writes it.
type newsRequest struct {
	Title models.LocalizedText `json:"title"`
	Body  models.LocalizedText `json:"body"`
	From  zero.Time            `json:"from"`
	Until zero.Time            `json:"until"`
	Icon  string               `json:"icon"`
	Color string               `json:"color"`
	// A draft is on hold: nobody sees it until it is published.
	Draft bool `json:"draft"`
}

func trimText(t models.LocalizedText) models.LocalizedText {
	trim := func(s zero.String) zero.String {
		return zero.StringFrom(strings.TrimSpace(s.String))
	}
	return models.LocalizedText{Sv: trim(t.Sv), En: trim(t.En), De: trim(t.De)}
}

// validate tidies the post and says what is wrong with it, if anything.
// English is required: readers in other languages fall back to it.
func (r *newsRequest) validate() error {
	r.Title = trimText(r.Title)
	r.Body = trimText(r.Body)
	r.Icon = strings.TrimSpace(r.Icon)
	if r.Icon == "" {
		r.Icon = "campaign"
	}
	for _, s := range []string{r.Title.Sv.String, r.Title.En.String, r.Title.De.String} {
		if utf8.RuneCountInString(s) > maxNewsTitle {
			return errors.New("title too long")
		}
	}
	for _, s := range []string{r.Body.Sv.String, r.Body.En.String, r.Body.De.String} {
		if utf8.RuneCountInString(s) > maxNewsBody {
			return errors.New("text too long")
		}
	}
	switch {
	case r.Title.En.String == "" || r.Body.En.String == "":
		return errors.New("an English title and text are required")
	case !newsIcon.MatchString(r.Icon):
		return errors.New("invalid icon")
	case !newsColor.MatchString(r.Color):
		return errors.New("invalid color")
	case r.From.Valid && r.Until.Valid && !r.From.Time.Before(r.Until.Time):
		return errors.New("the post must be taken down after it is published")
	}
	return nil
}

func (r *newsRequest) apply(a *models.Announcement) {
	a.Name = r.Title.En.String
	a.Title = r.Title
	a.Body = r.Body
	a.From = r.From
	a.Until = r.Until
	a.Icon = r.Icon
	a.Color = r.Color
}

// Announcements writes a new post.
func Announcements(c *gin.Context) {
	var r newsRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Message: err.Error(), Code: ErrorParse})
		return
	}
	if err := r.validate(); err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Message: err.Error(), Code: ErrorInvalidRequest})
		return
	}
	user, err := currentUser(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Message: err.Error(), Code: ErrorRecordNotFound})
		return
	}

	announcement := models.Announcement{
		Model:  models.Model{Timestamp: time.Now(), Uuid: utils.GenerateShortUuid()},
		UserId: zero.IntFrom(user.Id),
	}
	r.apply(&announcement)
	if r.Draft {
		announcement.OnHold = zero.TimeFrom(time.Now())
	}
	if err := models.DB.Create(&announcement).Error; err != nil {
		c.JSON(http.StatusInternalServerError, GeneralErrorResponse{Message: err.Error(), Code: ErrorInsert})
		return
	}
	c.JSON(http.StatusOK, GeneralSuccessResponse{Message: "success", Data: announcement})
}

// AnnouncementsUpdate rewrites a post. Whether it is published is changed
// with pause and resume, not here.
func AnnouncementsUpdate(c *gin.Context) {
	var r newsRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Message: err.Error(), Code: ErrorParse})
		return
	}
	if err := r.validate(); err != nil {
		c.JSON(http.StatusBadRequest, GeneralErrorResponse{Message: err.Error(), Code: ErrorInvalidRequest})
		return
	}
	var announcement models.Announcement
	if err := models.DB.Where("uuid=?", c.Param("uuid")).First(&announcement).Error; err != nil {
		c.JSON(http.StatusNotFound, GeneralErrorResponse{Message: "not found", Code: ErrorRecordNotFound})
		return
	}
	r.apply(&announcement)
	// Select every field, so cleared dates and texts are saved as empty.
	if err := models.DB.Model(&announcement).
		Select("name", "title_sv", "title_en", "title_de", "body_sv", "body_en", "body_de", "from", "until", "icon", "color").
		Updates(&announcement).Error; err != nil {
		c.JSON(http.StatusInternalServerError, GeneralErrorResponse{Message: err.Error(), Code: ErrorInsert})
		return
	}
	c.JSON(http.StatusOK, GeneralSuccessResponse{Message: "success", Data: announcement})
}

func setOnHold(c *gin.Context, value any) {
	res := models.DB.Table(table.Announcements).Where("uuid=?", c.Param("uuid")).UpdateColumn("on_hold", value)
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, GeneralErrorResponse{Message: res.Error.Error(), Code: ErrorInsert})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, GeneralErrorResponse{Message: "not found", Code: ErrorRecordNotFound})
		return
	}
	c.JSON(http.StatusOK, GeneralSuccessResponse{Message: "success"})
}

// AnnouncementsPause takes a post down (or keeps it a draft).
func AnnouncementsPause(c *gin.Context) {
	setOnHold(c, time.Now())
}

// AnnouncementsResume publishes a post: at once, or at its from date.
func AnnouncementsResume(c *gin.Context) {
	setOnHold(c, nil)
}

func AnnouncementsDelete(c *gin.Context) {
	if err := models.DB.Table(table.Announcements).Where("uuid=?", c.Param("uuid")).Delete(&models.Announcement{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, GeneralErrorResponse{Message: err.Error(), Code: ErrorInsert})
		return
	}
	c.JSON(http.StatusOK, GeneralSuccessResponse{Message: "success"})
}

// AdminNewsPost is a post as the admin tool shows it.
type AdminNewsPost struct {
	models.Announcement
	PublishedAt time.Time `json:"published_at"`
	// How many have read it.
	Reads int64 `json:"reads"`
}

// AnnouncementsAll is every post, drafts and old ones too, newest first.
func AnnouncementsAll(c *gin.Context) {
	var announcements []models.Announcement
	if err := models.DB.Order(publishedAt + " DESC").Order("announcements.id DESC").Find(&announcements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, GeneralErrorResponse{Message: err.Error(), Code: ErrorUnknown})
		return
	}
	type count struct {
		AnnouncementId int64
		N              int64
	}
	var counts []count
	if err := models.DB.Table("announcement_users").Select("announcement_id, COUNT(*) AS n").Group("announcement_id").Scan(&counts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, GeneralErrorResponse{Message: err.Error(), Code: ErrorUnknown})
		return
	}
	reads := map[int64]int64{}
	for _, n := range counts {
		reads[n.AnnouncementId] = n.N
	}
	posts := []AdminNewsPost{}
	for _, a := range announcements {
		posts = append(posts, AdminNewsPost{Announcement: a, PublishedAt: newsPost(a, false).PublishedAt, Reads: reads[a.Id]})
	}
	c.JSON(http.StatusOK, GeneralSuccessResponse{Message: "success", Data: posts})
}
