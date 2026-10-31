// SPDX-License-Identifier: AGPL-3.0-or-later
package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/server/log"
	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/table"
	"gorm.io/gorm"
)

// Admin statistics: how the site grows and who comes back. Activity is a
// sighting, the only thing the database keeps a history of; someone who only
// looks counts as inactive.

const (
	statsWeeks  = 52
	statsMonths = 12
)

// StatsCohort is the accounts made in one month, and how many of them reported
// a sighting in that month (Retained[0]) and each month after.
type StatsCohort struct {
	Month    string  `json:"month"`
	Size     int64   `json:"size"`
	Retained []int64 `json:"retained"`
}

// StatsCompetition is a recent competition and how many take part.
type StatsCompetition struct {
	Name         models.Translation `json:"name"`
	Participants int64              `json:"participants"`
}

type AdminStatsResponse struct {
	// Mondays of the weeks the weekly series count, oldest first.
	Weeks        []string `json:"weeks"`
	Signups      []int64  `json:"signups"`
	Accounts     []int64  `json:"accounts"`
	WeeklyActive []int64  `json:"weekly_active"`
	Sightings    []int64  `json:"sightings"`
	Lists        []int64  `json:"lists"`
	Competitions []int64  `json:"competitions"`
	// The months the monthly series count, oldest first.
	Months        []string           `json:"months"`
	MonthlyActive []int64            `json:"monthly_active"`
	Cohorts       []StatsCohort      `json:"cohorts"`
	SharedLists   int64              `json:"shared_lists"`
	Recent        []StatsCompetition `json:"recent_competitions"`
}

// weekStart is the SQL for the Monday of column's week.
func weekStart(column string) string {
	return "DATE(DATE_SUB(" + column + ", INTERVAL WEEKDAY(" + column + ") DAY))"
}

// perPeriod counts rows of q per period (the SQL period of each row) into
// one value per key, zero where there are none.
func perPeriod(q *gorm.DB, period, count string, keys []string, layout string) ([]int64, error) {
	var rows []struct {
		Period time.Time
		N      int64
	}
	if err := q.Select(period + " AS period, " + count + " AS n").Group("period").Scan(&rows).Error; err != nil {
		return nil, err
	}
	at := map[string]int64{}
	for _, r := range rows {
		at[r.Period.Format(layout)] = r.N
	}
	out := make([]int64, len(keys))
	for i, k := range keys {
		out[i] = at[k]
	}
	return out, nil
}

func AdminStats(c *gin.Context) {
	log.Info(c, 0, "AdminStats")
	// The database keeps local times (loc=Local in data-source), so periods are counted in them.
	res, err := adminStats(time.Now())
	if err != nil {
		log.Error(c, ErrorInternalServerError, err.Error())
		c.JSON(http.StatusInternalServerError, GeneralErrorResponse{Code: ErrorInternalServerError, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func adminStats(now time.Time) (*AdminStatsResponse, error) {
	var res AdminStatsResponse
	db := models.DB

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	monday := today.AddDate(0, 0, -((int(today.Weekday()) + 6) % 7))
	first := monday.AddDate(0, 0, -7*(statsWeeks-1))
	for i := 0; i < statsWeeks; i++ {
		res.Weeks = append(res.Weeks, first.AddDate(0, 0, 7*i).Format(time.DateOnly))
	}
	firstMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -(statsMonths - 1), 0)
	for i := 0; i < statsMonths; i++ {
		res.Months = append(res.Months, firstMonth.AddDate(0, i, 0).Format("2006-01"))
	}

	weekly := func(tbl, column, count string, where ...any) ([]int64, error) {
		q := db.Table(tbl).Where(column+" >= ?", first)
		if len(where) > 0 {
			q = q.Where(where[0], where[1:]...)
		}
		return perPeriod(q, weekStart(column), count, res.Weeks, time.DateOnly)
	}
	var err error
	if res.Signups, err = weekly(table.Users, "timestamp", "COUNT(*)"); err != nil {
		return nil, err
	}
	// Everyone who ever signed up, at the end of each week; deleted accounts too,
	// so the line only grows with signups.
	var before int64
	if err := db.Table(table.Users).Where("timestamp < ?", first).Count(&before).Error; err != nil {
		return nil, err
	}
	for _, n := range res.Signups {
		before += n
		res.Accounts = append(res.Accounts, before)
	}
	// Archived sightings still show that someone was active.
	if res.WeeklyActive, err = weekly(table.Reports, "timestamp", "COUNT(DISTINCT user_id)"); err != nil {
		return nil, err
	}
	if res.Sightings, err = weekly(table.Reports, "timestamp", "COUNT(*)", "archived IS NULL"); err != nil {
		return nil, err
	}
	if res.Lists, err = weekly(table.Lists, "timestamp", "COUNT(*)"); err != nil {
		return nil, err
	}
	if res.Competitions, err = weekly("competitions", "timestamp", "COUNT(*)"); err != nil {
		return nil, err
	}

	monthly := db.Table(table.Reports).Where("timestamp >= ?", firstMonth)
	if res.MonthlyActive, err = perPeriod(monthly, "DATE(DATE_FORMAT(timestamp, '%Y-%m-01'))", "COUNT(DISTINCT user_id)", res.Months, "2006-01"); err != nil {
		return nil, err
	}

	if res.Cohorts, err = statsCohorts(firstMonth, res.Months); err != nil {
		return nil, err
	}

	if err := db.Table(table.Lists).Where("visibility<>? AND archived IS NULL", models.VisibilityPrivate).Count(&res.SharedLists).Error; err != nil {
		return nil, err
	}

	var competitions []models.Competition
	if err := db.Where("archived IS NULL").Order("timestamp DESC").Limit(10).Preload("Name").Find(&competitions).Error; err != nil {
		return nil, err
	}
	res.Recent = []StatsCompetition{}
	for _, comp := range competitions {
		s := StatsCompetition{Name: comp.Name}
		if err := db.Table("competition_users").Where("competition_id=?", comp.Id).Count(&s.Participants).Error; err != nil {
			return nil, err
		}
		res.Recent = append(res.Recent, s)
	}
	return &res, nil
}

// statsCohorts are the accounts of each month since firstMonth, and how many
// of them reported a sighting in each month from the one they signed up in.
func statsCohorts(firstMonth time.Time, months []string) ([]StatsCohort, error) {
	var sizes []struct {
		Month string
		N     int64
	}
	if err := models.DB.Table(table.Users).Where("timestamp >= ?", firstMonth).
		Select("DATE_FORMAT(timestamp, '%Y-%m') AS month, COUNT(*) AS n").Group("month").Scan(&sizes).Error; err != nil {
		return nil, err
	}
	var active []struct {
		Month string
		After int
		N     int64
	}
	if err := models.DB.Table(table.Users+" AS u").Joins("JOIN "+table.Reports+" r ON r.user_id=u.id").
		Where("u.timestamp >= ? AND r.timestamp >= u.timestamp", firstMonth).
		Select("DATE_FORMAT(u.timestamp, '%Y-%m') AS month, PERIOD_DIFF(DATE_FORMAT(r.timestamp, '%Y%m'), DATE_FORMAT(u.timestamp, '%Y%m')) AS `after`, COUNT(DISTINCT u.id) AS n").
		Group("month, `after`").Scan(&active).Error; err != nil {
		return nil, err
	}

	cohorts := make([]StatsCohort, len(months))
	index := map[string]int{}
	for i, m := range months {
		// A cohort can only be followed up to this month.
		cohorts[i] = StatsCohort{Month: m, Retained: make([]int64, len(months)-i)}
		index[m] = i
	}
	for _, s := range sizes {
		if i, ok := index[s.Month]; ok {
			cohorts[i].Size = s.N
		}
	}
	for _, a := range active {
		if i, ok := index[a.Month]; ok && a.After >= 0 && a.After < len(cohorts[i].Retained) {
			cohorts[i].Retained[a.After] = a.N
		}
	}
	return cohorts, nil
}
