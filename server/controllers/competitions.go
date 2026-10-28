package controllers

import (
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/table"
	"github.com/klahr/300sf/server/utils"
	"gopkg.in/guregu/null.v4"
	"gopkg.in/guregu/null.v4/zero"
	"gorm.io/gorm"
)

// Competitions: who sees the most species in a period. A participant's score is the number of
// different species they have reported with a sighting date inside the period, counting only
// species on the competition's list (all species when it has none) and, when it is limited to
// some counties, sightings placed in them. With end_on_reports set, the first participant to
// reach that many species wins and the competition ends at that sighting.
//
// Public competitions are listed for everyone. Private ones are reached through their link, whose
// uuid is long and random; either kind can also have a password, which invited users skip.

const (
	competitionStatusUpcoming = "upcoming"
	competitionStatusOngoing  = "ongoing"
	competitionStatusFinished = "finished"
)

type CompetitionPerson struct {
	DisplayName string `json:"display_name"`
	Username    string `json:"username"`
	Portrait    string `json:"portrait"`
}

type CompetitionStanding struct {
	Rank   int               `json:"rank"`
	User   CompetitionPerson `json:"user"`
	Count  int               `json:"count"`
	Latest null.Time         `json:"latest"`
	Me     bool              `json:"me"`
}

type CompetitionInfo struct {
	Uuid           string               `json:"uuid"`
	Name           models.Translation   `json:"name"`
	Description    models.Translation   `json:"description"`
	Color          string               `json:"color"`
	From           time.Time            `json:"from"`
	Until          time.Time            `json:"until"`
	JoinUntil      time.Time            `json:"join_until"`
	EndOnReports   zero.Int             `json:"end_on_reports"`
	MaxCompetitors zero.Int             `json:"max_competitors"`
	Visibility     int64                `json:"visibility"`
	HasPassword    bool                 `json:"has_password"`
	ListName       *models.Translation  `json:"list_name"`
	ListSize       int64                `json:"list_size"`
	Zones          []models.Translation `json:"zones"`
	Creator        CompetitionPerson    `json:"creator"`
	Participants   int                  `json:"participants"`
	Status         string               `json:"status"`
	EndedAt        null.Time            `json:"ended_at"`
	Joined         bool                 `json:"joined"`
	IsCreator      bool                 `json:"is_creator"`
	Invited        bool                 `json:"invited"`
	CanJoin        bool                 `json:"can_join"`
	MyRank         int                  `json:"my_rank"`
	MyCount        int                  `json:"my_count"`
	// Nil when the viewer may not see who takes part.
	Standings []CompetitionStanding `json:"standings"`
}

func competitionError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"code":    -10001,
		"message": message,
	})
}

func currentUser(c *gin.Context) (*models.User, error) {
	var user models.User
	if err := models.DB.Where("uuid=?", c.GetString("uuid")).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func loadCompetition(uuid string) (*models.Competition, error) {
	var competition models.Competition
	err := models.DB.Where("uuid=? AND archived IS NULL", uuid).
		Preload("Name").Preload("Description").Preload("User").
		Preload("CompetitionZones", func(db *gorm.DB) *gorm.DB { return db.Select("id", "uuid", "name_id") }).
		Preload("CompetitionZones.Name").
		First(&competition).Error
	if err != nil {
		return nil, err
	}
	return &competition, nil
}

func competitionParticipants(competitionId int64) ([]models.User, error) {
	var users []models.User
	err := models.DB.Table("users").
		Joins("JOIN competition_users cu ON cu.user_id=users.id").
		Where("cu.competition_id=? AND users.deleted_at IS NULL", competitionId).
		Find(&users).Error
	return users, err
}

func pendingInvite(competitionId, userId int64) bool {
	var n int64
	models.DB.Model(&models.CompetitionInvite{}).
		Where("competition_id=? AND user_id=? AND accepted_at IS NULL AND rejected_at IS NULL", competitionId, userId).
		Count(&n)
	return n > 0
}

func person(u models.User) CompetitionPerson {
	return CompetitionPerson{DisplayName: u.DisplayName, Username: u.Username, Portrait: u.Portrait.String}
}

// competitionSighting is a participant's first sighting of a species that counts.
type competitionSighting struct {
	UserId    int64
	SpeciesId int64
	At        time.Time
}

// competitionStandings ranks the participants and tells when the competition was won, if it ended
// by someone reaching end_on_reports species. Ties go to whoever got there first.
func competitionStandings(competition *models.Competition, participants []models.User) ([]CompetitionStanding, null.Time, error) {
	if len(participants) == 0 {
		return []CompetitionStanding{}, null.Time{}, nil
	}

	ids := make([]int64, len(participants))
	for i, p := range participants {
		ids[i] = p.Id
	}
	at := "COALESCE(r.`from`, r.timestamp)"
	q := models.DB.Table("reports AS r").
		Select("r.user_id, r.species_id, MIN("+at+") AS at").
		Where("r.archived IS NULL AND r.user_id IN ?", ids).
		Where(at+" >= ? AND "+at+" < ?", competition.From, competition.Until)
	if competition.ListTemplateId.Valid {
		q = q.Where("r.species_id IN (SELECT species_id FROM "+table.ListTemplateSpecies+" WHERE list_template_id=?)", competition.ListTemplateId.Int64)
	}
	if len(competition.CompetitionZones) > 0 {
		zones := make([]int64, len(competition.CompetitionZones))
		for i, z := range competition.CompetitionZones {
			zones[i] = z.Id
		}
		q = q.Joins("LEFT JOIN zones z ON z.id=r.zone_id").Where("r.zone_id IN ? OR z.parent_zone_id IN ?", zones, zones)
	}
	var sightings []competitionSighting
	if err := q.Group("r.user_id, r.species_id").Scan(&sightings).Error; err != nil {
		return nil, null.Time{}, err
	}
	standings, wonAt := rankCompetition(participants, sightings, competition.EndOnReports.Int64)
	return standings, wonAt, nil
}

// rankCompetition turns each participant's first sightings of the species into standings. With
// endOnReports above zero, the first to reach that many species wins at that moment and later
// sightings do not count.
func rankCompetition(participants []models.User, sightings []competitionSighting, endOnReports int64) ([]CompetitionStanding, null.Time) {
	standings := []CompetitionStanding{}
	times := map[int64][]time.Time{}
	for _, s := range sightings {
		times[s.UserId] = append(times[s.UserId], s.At)
	}
	for _, t := range times {
		sort.Slice(t, func(i, j int) bool { return t[i].Before(t[j]) })
	}

	var wonAt null.Time
	if endOnReports > 0 {
		n := int(endOnReports)
		for _, t := range times {
			if len(t) >= n && (!wonAt.Valid || t[n-1].Before(wonAt.Time)) {
				wonAt = null.TimeFrom(t[n-1])
			}
		}
	}

	for _, p := range participants {
		s := CompetitionStanding{User: person(p)}
		for _, t := range times[p.Id] {
			if wonAt.Valid && t.After(wonAt.Time) {
				break
			}
			s.Count++
			s.Latest = null.TimeFrom(t)
		}
		standings = append(standings, s)
	}
	sort.SliceStable(standings, func(i, j int) bool {
		a, b := standings[i], standings[j]
		if a.Count != b.Count {
			return a.Count > b.Count
		}
		if a.Latest.Valid && b.Latest.Valid && !a.Latest.Time.Equal(b.Latest.Time) {
			return a.Latest.Time.Before(b.Latest.Time)
		}
		return strings.ToLower(a.User.DisplayName) < strings.ToLower(b.User.DisplayName)
	})
	for i := range standings {
		standings[i].Rank = i + 1
		if i > 0 && standings[i].Count == standings[i-1].Count && standings[i].Latest == standings[i-1].Latest {
			standings[i].Rank = standings[i-1].Rank
		}
	}
	return standings, wonAt
}

// competitionInfo describes a competition as the viewer (nil when not signed in) may see it.
func competitionInfo(competition *models.Competition, viewer *models.User, withStandings bool) (*CompetitionInfo, error) {
	participants, err := competitionParticipants(competition.Id)
	if err != nil {
		return nil, err
	}
	standings, wonAt, err := competitionStandings(competition, participants)
	if err != nil {
		return nil, err
	}

	info := CompetitionInfo{
		Uuid:           competition.Uuid,
		Name:           competition.Name,
		Description:    competition.Description,
		Color:          competition.Color,
		From:           competition.From,
		Until:          competition.Until,
		JoinUntil:      competition.JoinUntil,
		EndOnReports:   competition.EndOnReports,
		MaxCompetitors: competition.MaxCompetitors,
		Visibility:     competition.Visibility,
		HasPassword:    competition.Password.Valid && competition.Password.String != "",
		Zones:          []models.Translation{},
		Participants:   len(participants),
	}
	for _, z := range competition.CompetitionZones {
		info.Zones = append(info.Zones, z.Name)
	}
	if competition.User != nil {
		info.Creator = person(*competition.User)
	}
	if competition.ListTemplateId.Valid {
		var template models.ListTemplate
		if err := models.DB.Preload("Name").First(&template, competition.ListTemplateId.Int64).Error; err == nil {
			info.ListName = &template.Name
			models.DB.Table(table.ListTemplateSpecies).Where("list_template_id=?", template.Id).Count(&info.ListSize)
		}
	}

	now := time.Now()
	switch {
	case wonAt.Valid:
		info.Status = competitionStatusFinished
		info.EndedAt = wonAt
	case !now.Before(competition.Until):
		info.Status = competitionStatusFinished
		info.EndedAt = null.TimeFrom(competition.Until)
	case now.Before(competition.From):
		info.Status = competitionStatusUpcoming
	default:
		info.Status = competitionStatusOngoing
	}

	if viewer != nil {
		for _, p := range participants {
			if p.Id == viewer.Id {
				info.Joined = true
			}
		}
		info.IsCreator = competition.UserId.Valid && competition.UserId.Int64 == viewer.Id
		info.Invited = !info.Joined && pendingInvite(competition.Id, viewer.Id)
		info.CanJoin = !info.Joined && competitionJoinable(competition, &info, now) == nil
		for i := range standings {
			if standings[i].User.Username == viewer.Username {
				standings[i].Me = true
				info.MyRank = standings[i].Rank
				info.MyCount = standings[i].Count
			}
		}
	}

	// Who takes part is shown in public competitions and to those who are in one.
	if withStandings && (competition.Visibility == models.VisibilityPublic || info.Joined || info.IsCreator || info.Invited) {
		info.Standings = standings
	}
	return &info, nil
}

func competitionJoinable(competition *models.Competition, info *CompetitionInfo, now time.Time) error {
	if info.Status == competitionStatusFinished {
		return errors.New("the competition has finished")
	}
	if now.Before(competition.JoinFrom) || !now.Before(competition.JoinUntil) {
		return errors.New("the competition is not open for joining")
	}
	if competition.MaxCompetitors.Valid && competition.MaxCompetitors.Int64 > 0 && int64(info.Participants) >= competition.MaxCompetitors.Int64 {
		return errors.New("the competition is full")
	}
	return nil
}

func translationIsEmpty(t models.Translation) bool {
	return strings.TrimSpace(t.Sv.String) == "" && strings.TrimSpace(t.En.String) == "" && strings.TrimSpace(t.De.String) == ""
}

func Competitions(c *gin.Context) {
	type request struct {
		Name                  models.Translation `json:"name"`
		Description           models.Translation `json:"description"`
		From                  time.Time          `json:"from"`
		Until                 time.Time          `json:"until"`
		ListTemplateUuid      string             `json:"list_template_uuid"`
		EndOnReports          int64              `json:"end_on_reports"`
		MaxCompetitors        int64              `json:"max_competitors"`
		Visibility            int64              `json:"visibility"`
		Color                 string             `json:"color"`
		CompetitionZonesUuids []string           `json:"competition_zones_uuids"`
		Password              string             `json:"password"`
		CreatorJoin           bool               `json:"creator_join"`
	}

	var r request
	if err := c.ShouldBindJSON(&r); err != nil {
		competitionError(c, http.StatusBadRequest, err.Error())
		return
	}
	switch {
	case translationIsEmpty(r.Name):
		competitionError(c, http.StatusBadRequest, "name required")
		return
	case !r.From.Before(r.Until):
		competitionError(c, http.StatusBadRequest, "the competition must end after it starts")
		return
	case !r.Until.After(time.Now()):
		competitionError(c, http.StatusBadRequest, "the competition must end in the future")
		return
	case r.Until.Sub(r.From) > 5*366*24*time.Hour:
		competitionError(c, http.StatusBadRequest, "the competition can last at most five years")
		return
	case r.Visibility != models.VisibilityPublic && r.Visibility != models.VisibilityPrivate:
		competitionError(c, http.StatusBadRequest, "invalid visibility")
		return
	case r.EndOnReports < 0 || r.MaxCompetitors < 0 || r.MaxCompetitors == 1:
		competitionError(c, http.StatusBadRequest, "invalid limits")
		return
	case len(r.Password) > 64 || len(r.Color) > 16:
		competitionError(c, http.StatusBadRequest, "invalid request")
		return
	}

	user, err := currentUser(c)
	if err != nil {
		competitionError(c, http.StatusBadRequest, err.Error())
		return
	}

	var uuid string
	err = models.DB.Transaction(func(tx *gorm.DB) error {
		name := models.Translation{Sv: r.Name.Sv, En: r.Name.En, De: r.Name.De}
		name.Timestamp = time.Now()
		name.Uuid = utils.GenerateShortUuid()
		if err := tx.Create(&name).Error; err != nil {
			return err
		}
		description := models.Translation{Sv: r.Description.Sv, En: r.Description.En, De: r.Description.De}
		description.Timestamp = time.Now()
		description.Uuid = utils.GenerateShortUuid()
		if err := tx.Create(&description).Error; err != nil {
			return err
		}

		competition := models.Competition{
			Model: models.Model{
				Timestamp: time.Now(),
				// The link is the way into a private competition, so it must not be guessable.
				Uuid: utils.RandomSha1()[:24],
			},
			NameId:        name.Id,
			DescriptionId: description.Id,
			UserId:        zero.IntFrom(user.Id),
			From:          r.From,
			Until:         r.Until,
			Visibility:    r.Visibility,
			JoinFrom:      time.Now(),
			JoinUntil:     r.Until,
			Icon:          "trophy",
			Color:         r.Color,
		}
		if r.ListTemplateUuid != "" {
			var template models.ListTemplate
			if err := tx.Where("uuid=? AND archived IS NULL", r.ListTemplateUuid).Where("user_id=? OR visibility=?", user.Id, models.VisibilityPublic).First(&template).Error; err != nil {
				return err
			}
			competition.ListTemplateId = zero.IntFrom(template.Id)
		}
		if r.EndOnReports > 0 {
			competition.EndOnReports = zero.IntFrom(r.EndOnReports)
		}
		if r.MaxCompetitors > 0 {
			competition.MaxCompetitors = zero.IntFrom(r.MaxCompetitors)
		}
		if r.Password != "" {
			p := utils.Password{}
			if err := p.Create(r.Password); err != nil {
				return err
			}
			competition.Password = zero.StringFrom(p.Value)
			competition.PasswordSalt = zero.StringFrom(p.Salt)
		}
		if len(r.CompetitionZonesUuids) > 0 {
			if err := tx.Select("id").Where("uuid IN ? AND admin_level=?", r.CompetitionZonesUuids, models.ZoneTypeCounty).Find(&competition.CompetitionZones).Error; err != nil {
				return err
			}
			if len(competition.CompetitionZones) != len(r.CompetitionZonesUuids) {
				return gorm.ErrRecordNotFound
			}
		}
		if err := tx.Omit("Name", "Description", "User", "ListTemplate", "Participants", "CompetitionZones.*").Create(&competition).Error; err != nil {
			return err
		}
		if r.CreatorJoin {
			if err := tx.Table("competition_users").Create(&models.CompetitionUsers{CompetitionId: competition.Id, UserId: user.Id}).Error; err != nil {
				return err
			}
		}
		uuid = competition.Uuid
		return nil
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		competitionError(c, http.StatusBadRequest, "unknown list or county")
		return
	}
	if err != nil {
		competitionError(c, http.StatusInternalServerError, err.Error())
		return
	}

	var info *CompetitionInfo
	competition, err := loadCompetition(uuid)
	if err == nil {
		info, err = competitionInfo(competition, user, true)
	}
	if err != nil {
		competitionError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, info)
}

// CompetitionsList is the public competitions that can still be joined.
func CompetitionsList(c *gin.Context) {
	user, err := currentUser(c)
	if err != nil {
		competitionError(c, http.StatusBadRequest, err.Error())
		return
	}
	var uuids []string
	if err := models.DB.Model(&models.Competition{}).
		Where("archived IS NULL AND visibility=? AND until>? AND join_until>?", models.VisibilityPublic, time.Now(), time.Now()).
		Order("`from`").Limit(50).Pluck("uuid", &uuids).Error; err != nil {
		competitionError(c, http.StatusInternalServerError, err.Error())
		return
	}
	competitionsResponse(c, uuids, user)
}

// CompetitionsListMe is the competitions the user takes part in, created or is invited to.
func CompetitionsListMe(c *gin.Context) {
	user, err := currentUser(c)
	if err != nil {
		competitionError(c, http.StatusBadRequest, err.Error())
		return
	}
	var uuids []string
	if err := models.DB.Model(&models.Competition{}).
		Where("archived IS NULL").
		Where("user_id=? OR id IN (SELECT competition_id FROM competition_users WHERE user_id=?) OR id IN (SELECT competition_id FROM competition_invites WHERE user_id=? AND accepted_at IS NULL AND rejected_at IS NULL)", user.Id, user.Id, user.Id).
		Order("until DESC").Limit(100).Pluck("uuid", &uuids).Error; err != nil {
		competitionError(c, http.StatusInternalServerError, err.Error())
		return
	}
	competitionsResponse(c, uuids, user)
}

func competitionsResponse(c *gin.Context, uuids []string, user *models.User) {
	result := []CompetitionInfo{}
	for _, uuid := range uuids {
		competition, err := loadCompetition(uuid)
		if err != nil {
			continue
		}
		info, err := competitionInfo(competition, user, false)
		if err != nil {
			competitionError(c, http.StatusInternalServerError, err.Error())
			return
		}
		result = append(result, *info)
	}
	c.JSON(http.StatusOK, gin.H{"result": result})
}

func CompetitionsGet(c *gin.Context) {
	user, err := currentUser(c)
	if err != nil {
		competitionError(c, http.StatusBadRequest, err.Error())
		return
	}
	competition, err := loadCompetition(c.Param("uuid"))
	if err != nil {
		competitionError(c, http.StatusNotFound, "not found")
		return
	}
	info, err := competitionInfo(competition, user, true)
	if err != nil {
		competitionError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, info)
}

// CompetitionsResult is the standings of a public competition, for anyone.
func CompetitionsResult(c *gin.Context) {
	competition, err := loadCompetition(c.Param("uuid"))
	if err != nil || competition.Visibility != models.VisibilityPublic {
		competitionError(c, http.StatusNotFound, "not found")
		return
	}
	info, err := competitionInfo(competition, nil, true)
	if err != nil {
		competitionError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, info)
}

func CompetitionsJoin(c *gin.Context) {
	type request struct {
		Password string `json:"password"`
	}
	var r request
	c.ShouldBindJSON(&r)

	user, err := currentUser(c)
	if err != nil {
		competitionError(c, http.StatusBadRequest, err.Error())
		return
	}
	competition, err := loadCompetition(c.Param("uuid"))
	if err != nil {
		competitionError(c, http.StatusNotFound, "not found")
		return
	}
	info, err := competitionInfo(competition, user, false)
	if err != nil {
		competitionError(c, http.StatusInternalServerError, err.Error())
		return
	}
	if info.Joined {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	if err := competitionJoinable(competition, info, time.Now()); err != nil {
		competitionError(c, http.StatusConflict, err.Error())
		return
	}
	if info.HasPassword && !info.Invited {
		p := utils.Password{}
		if !p.IsMatching(r.Password, competition.Password.String, competition.PasswordSalt.String) {
			c.JSON(http.StatusForbidden, gin.H{"code": ErrorCompetitionPassword, "message": "wrong password"})
			return
		}
	}
	err = models.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("competition_users").Create(&models.CompetitionUsers{CompetitionId: competition.Id, UserId: user.Id}).Error; err != nil {
			return err
		}
		return tx.Model(&models.CompetitionInvite{}).
			Where("competition_id=? AND user_id=? AND accepted_at IS NULL AND rejected_at IS NULL", competition.Id, user.Id).
			Update("accepted_at", time.Now()).Error
	})
	if err != nil {
		competitionError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func CompetitionsLeave(c *gin.Context) {
	user, err := currentUser(c)
	if err != nil {
		competitionError(c, http.StatusBadRequest, err.Error())
		return
	}
	competition, err := loadCompetition(c.Param("uuid"))
	if err != nil {
		competitionError(c, http.StatusNotFound, "not found")
		return
	}
	if err := models.DB.Table("competition_users").Where("competition_id=? AND user_id=?", competition.Id, user.Id).Delete(&models.CompetitionUsers{}).Error; err != nil {
		competitionError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// CompetitionsInvite invites users by username. Anyone taking part, and the creator, may invite.
func CompetitionsInvite(c *gin.Context) {
	type request struct {
		Usernames []string `json:"usernames"`
	}
	var r request
	if err := c.ShouldBindJSON(&r); err != nil || len(r.Usernames) == 0 || len(r.Usernames) > 50 {
		competitionError(c, http.StatusBadRequest, "usernames required")
		return
	}
	user, err := currentUser(c)
	if err != nil {
		competitionError(c, http.StatusBadRequest, err.Error())
		return
	}
	competition, err := loadCompetition(c.Param("uuid"))
	if err != nil {
		competitionError(c, http.StatusNotFound, "not found")
		return
	}
	info, err := competitionInfo(competition, user, false)
	if err != nil {
		competitionError(c, http.StatusInternalServerError, err.Error())
		return
	}
	if !info.Joined && !info.IsCreator {
		competitionError(c, http.StatusForbidden, "only participants can invite")
		return
	}
	if info.Status == competitionStatusFinished {
		competitionError(c, http.StatusConflict, "the competition has finished")
		return
	}

	invited, notFound := []string{}, []string{}
	for _, username := range r.Usernames {
		username = strings.TrimPrefix(strings.TrimSpace(username), "@")
		if username == "" {
			continue
		}
		var invitee models.User
		if err := models.DB.Where("username=? AND deleted_at IS NULL", username).First(&invitee).Error; err != nil {
			notFound = append(notFound, username)
			continue
		}
		var n int64
		models.DB.Table("competition_users").Where("competition_id=? AND user_id=?", competition.Id, invitee.Id).Count(&n)
		if n == 0 && !pendingInvite(competition.Id, invitee.Id) {
			invite := models.CompetitionInvite{
				Model:         models.Model{Timestamp: time.Now(), Uuid: utils.GenerateShortUuid()},
				UserId:        invitee.Id,
				CompetitionId: competition.Id,
			}
			if err := models.DB.Omit("User").Create(&invite).Error; err != nil {
				competitionError(c, http.StatusInternalServerError, err.Error())
				return
			}
		}
		invited = append(invited, invitee.Username)
	}
	c.JSON(http.StatusOK, gin.H{"invited": invited, "not_found": notFound})
}

func CompetitionsDecline(c *gin.Context) {
	user, err := currentUser(c)
	if err != nil {
		competitionError(c, http.StatusBadRequest, err.Error())
		return
	}
	competition, err := loadCompetition(c.Param("uuid"))
	if err != nil {
		competitionError(c, http.StatusNotFound, "not found")
		return
	}
	if err := models.DB.Model(&models.CompetitionInvite{}).
		Where("competition_id=? AND user_id=? AND accepted_at IS NULL AND rejected_at IS NULL", competition.Id, user.Id).
		Update("rejected_at", time.Now()).Error; err != nil {
		competitionError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// CompetitionsArchive removes a competition; only its creator may.
func CompetitionsArchive(c *gin.Context) {
	user, err := currentUser(c)
	if err != nil {
		competitionError(c, http.StatusBadRequest, err.Error())
		return
	}
	competition, err := loadCompetition(c.Param("uuid"))
	if err != nil {
		competitionError(c, http.StatusNotFound, "not found")
		return
	}
	if !competition.UserId.Valid || competition.UserId.Int64 != user.Id {
		competitionError(c, http.StatusForbidden, "only the creator can remove the competition")
		return
	}
	if err := models.DB.Model(&models.Competition{}).Where("id=?", competition.Id).Update("archived", time.Now()).Error; err != nil {
		competitionError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// CompetitionsBasis is what the form for a new competition chooses from: the counties and the
// species lists the user may use.
func CompetitionsBasis(c *gin.Context) {
	type listTemplate struct {
		Uuid        string             `json:"uuid"`
		Name        models.Translation `json:"name"`
		Description models.Translation `json:"description"`
		Size        int64              `json:"size"`
	}
	type response struct {
		Zones         []models.Zone  `json:"zones"`
		ListTemplates []listTemplate `json:"list_templates"`
	}

	user, err := currentUser(c)
	if err != nil {
		competitionError(c, http.StatusBadRequest, err.Error())
		return
	}

	var r response
	if err := models.DB.Select("id", "uuid", "name_id", "admin_level").Where("admin_level=?", models.ZoneTypeCounty).Preload("Name").Find(&r.Zones).Error; err != nil {
		competitionError(c, http.StatusInternalServerError, "could not find zones")
		return
	}

	var templates []models.ListTemplate
	if err := models.DB.Table(table.ListTemplates).Where("user_id=? OR visibility=?", user.Id, models.VisibilityPublic).Where("archived IS NULL").Preload("Name").Preload("Description").Find(&templates).Error; err != nil {
		competitionError(c, http.StatusInternalServerError, err.Error())
		return
	}
	r.ListTemplates = []listTemplate{}
	for _, t := range templates {
		l := listTemplate{Uuid: t.Uuid, Name: t.Name, Description: t.Description}
		models.DB.Table(table.ListTemplateSpecies).Where("list_template_id=?", t.Id).Count(&l.Size)
		r.ListTemplates = append(r.ListTemplates, l)
	}

	c.JSON(http.StatusOK, r)
}
