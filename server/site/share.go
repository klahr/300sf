// SPDX-License-Identifier: AGPL-3.0-or-later
package site

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/server/config"
	"github.com/klahr/300sf/server/controllers"
	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/wikipedia"
)

// The share page, /share/<list>, shows a list's progress to anyone with the
// link, big enough for a screen on the wall. It reloads itself now and then.

var shareTexts = map[string]map[string]string{
	"sv": {
		"Competition":      "Tävling",
		"Participants":     "deltagare",
		"DaysLeft":         "{0} dagar kvar",
		"LastDay":          "Sista dagen",
		"StartsOn":         "Startar {0}",
		"FinishedOn":       "Avslutad {0}",
		"Winner":           "Vinnare",
		"FirstTo":          "Först till {0} arter vinner",
		"LatestFinds":      "Senaste fynden",
		"Where":            "Var arterna hittades",
		"AreaSpecies":      "{0} arter",
		"NoFindsYet":       "Inga fynd än.",
		"Join":             "Var med i tävlingen",
		"NoParticipants":   "Ingen är med än.",
		"TakenBy":          "Hittad av {0} av {1}",
		"SearchListen":     "Lyssna",
		"SearchPause":      "Pausa",
		"SearchReadMore":   "Läs mer på Wikipedia",
		"SearchReport":     "Rapportera fynd",
		"SearchReportHint": "Skapa ett gratis konto så väntar fyndet där, redo att sparas.",
		"SearchFailed":     "Kunde inte hämta arten. Försök igen om en stund.",
		"Close":            "Stäng",
		"Language":         "Språk",
		"Checklist":        "Checklista",
		"Of":               "av",
		"Species":          "arter",
		"ToGo":             "kvar",
		"Newest":           "Senaste art",
		"NoneYet":          "Inga arter än. Den första är nära.",
		"NotSharedTitle":   "Listan delas inte",
		"NotSharedText":    "Den finns inte, eller så har den som äger den gjort den privat.",
	},
	"en": {
		"Competition":      "Competition",
		"Participants":     "participants",
		"DaysLeft":         "{0} days left",
		"LastDay":          "Last day",
		"StartsOn":         "Starts {0}",
		"FinishedOn":       "Finished {0}",
		"Winner":           "Winner",
		"FirstTo":          "First to {0} species wins",
		"LatestFinds":      "Latest finds",
		"Where":            "Where species were found",
		"AreaSpecies":      "{0} species",
		"NoFindsYet":       "No finds yet.",
		"Join":             "Join the competition",
		"NoParticipants":   "Nobody has joined yet.",
		"TakenBy":          "Found by {0} of {1}",
		"SearchListen":     "Listen",
		"SearchPause":      "Pause",
		"SearchReadMore":   "Read more on Wikipedia",
		"SearchReport":     "Report a sighting",
		"SearchReportHint": "Create a free account and the sighting is waiting there, ready to save.",
		"SearchFailed":     "Couldn't load the species. Try again in a moment.",
		"Close":            "Close",
		"Language":         "Language",
		"Checklist":        "Checklist",
		"Of":               "of",
		"Species":          "species",
		"ToGo":             "to go",
		"Newest":           "Newest species",
		"NoneYet":          "No species yet. The first one is close.",
		"NotSharedTitle":   "This list isn't shared",
		"NotSharedText":    "It doesn't exist, or its owner has made it private.",
	},
	"de": {
		"Competition":      "Wettbewerb",
		"Participants":     "Teilnehmende",
		"DaysLeft":         "Noch {0} Tage",
		"LastDay":          "Letzter Tag",
		"StartsOn":         "Beginnt am {0}",
		"FinishedOn":       "Beendet am {0}",
		"Winner":           "Gewinner",
		"FirstTo":          "Wer zuerst {0} Arten hat, gewinnt",
		"LatestFinds":      "Neueste Funde",
		"Where":            "Wo Arten gefunden wurden",
		"AreaSpecies":      "{0} Arten",
		"NoFindsYet":       "Noch keine Funde.",
		"Join":             "Mitmachen",
		"NoParticipants":   "Noch niemand dabei.",
		"TakenBy":          "{0} von {1} haben sie gefunden",
		"SearchListen":     "Anhören",
		"SearchPause":      "Pause",
		"SearchReadMore":   "Mehr auf Wikipedia",
		"SearchReport":     "Beobachtung melden",
		"SearchReportHint": "Erstelle ein kostenloses Konto, dann wartet die Beobachtung dort, bereit zum Speichern.",
		"SearchFailed":     "Die Art konnte nicht geladen werden. Versuche es gleich noch einmal.",
		"Close":            "Schließen",
		"Language":         "Sprache",
		"Checklist":        "Checkliste",
		"Of":               "von",
		"Species":          "Arten",
		"ToGo":             "fehlen noch",
		"Newest":           "Neueste Art",
		"NoneYet":          "Noch keine Arten. Die erste ist nah.",
		"NotSharedTitle":   "Diese Liste wird nicht geteilt",
		"NotSharedText":    "Es gibt sie nicht, oder sie wurde privat gemacht.",
	},
}

var shareMonths = map[string][12]string{
	"sv": {"jan", "feb", "mar", "apr", "maj", "jun", "jul", "aug", "sep", "okt", "nov", "dec"},
	"en": {"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"},
	"de": {"Jan.", "Feb.", "März", "Apr.", "Mai", "Juni", "Juli", "Aug.", "Sep.", "Okt.", "Nov.", "Dez."},
}

type shareSpecies struct {
	Uuid   string
	Name   models.Translation
	Seen   bool
	Recent bool
	Date   string
}

type shareGroup struct {
	Order   models.Translation
	Seen    int
	Species []shareSpecies
}

func (s *Site) sharePage(c *gin.Context) {
	lang := language(c, []string{"sv", "en", "de"}, "sv")
	data := gin.H{"Lang": lang, "Title": appNames[lang], "T": shareTexts[lang]}
	list, err := s.share(c.Param("uuid"))
	if err != nil {
		s.render(c, http.StatusNotFound, "share.html", data)
		return
	}

	date := func(t time.Time) string {
		return fmt.Sprintf("%d %s", t.Day(), shareMonths[lang][t.Month()-1])
	}
	recent := map[int64]bool{}
	for _, l := range list.Latest {
		recent[l.Name.Id] = true
	}
	var groups []*shareGroup
	for _, sp := range list.Species {
		if len(groups) == 0 || groups[len(groups)-1].Order.Id != sp.Order.Id {
			groups = append(groups, &shareGroup{Order: sp.Order})
		}
		g := groups[len(groups)-1]
		e := shareSpecies{Uuid: sp.Uuid, Name: sp.Name, Seen: !sp.Seen.IsZero(), Recent: recent[sp.Name.Id]}
		if e.Seen {
			e.Date = date(sp.Seen)
			g.Seen++
		}
		g.Species = append(g.Species, e)
	}

	total := len(list.Species)
	percent := 0
	if total > 0 {
		percent = list.Seen * 100 / total
	}
	data["List"] = list
	data["Total"] = total
	data["Percent"] = percent
	data["ToGo"] = total - list.Seen
	data["Groups"] = groups
	if len(list.Latest) > 0 {
		newest := list.Latest[0]
		data["Newest"] = shareSpecies{Uuid: newest.Uuid, Name: newest.Name, Seen: true, Date: date(newest.Seen)}
		data["Photo"] = s.photo(newest.Name)
		var more []shareSpecies
		for _, l := range list.Latest[1:] {
			more = append(more, shareSpecies{Uuid: l.Uuid, Name: l.Name, Seen: true, Date: date(l.Seen)})
		}
		data["More"] = more
	}
	s.render(c, http.StatusOK, "share.html", data)
}

// speciesPhoto is a large Wikipedia picture of a species, or "" while there is
// none. Wikipedia is slow, so a picture not yet looked up is fetched in the
// background and shows when the page next reloads.
func speciesPhoto(name models.Translation) string {
	key := "share/photo/" + name.En.String
	if url, err := models.Cache.Get(key).Result(); err == nil {
		return url
	}
	go func() {
		models.Cache.Set(key, wikipedia.Photo("en", name, 1280), time.Second*time.Duration(config.AppConfig.WikipediaCacheDuration))
	}()
	return ""
}

// shareDay is a day and month, for lists where the year is clear from the page.
func shareDay(t time.Time, lang string) string {
	return fmt.Sprintf("%d %s", t.Day(), shareMonths[lang][t.Month()-1])
}

// shareDate is a day the way the share pages write it, with the year when it isn't this one.
func shareDate(t time.Time, lang string) string {
	d := fmt.Sprintf("%d %s", t.Day(), shareMonths[lang][t.Month()-1])
	if t.Year() != time.Now().Year() {
		d += " " + strconv.Itoa(t.Year())
	}
	return d
}

type shareStanding struct {
	controllers.CompetitionStanding
	Percent int
	Latest  string
}

type shareRaceSpecies struct {
	Uuid    string
	Name    models.Translation
	Finders int
	// Share is the percentage of participants who found it; Tip says how many.
	Share  int
	Tip    string
	Recent bool
	Date   string
}

type shareRaceGroup struct {
	Order   models.Translation
	Found   int
	Species []shareRaceSpecies
}

type shareFind struct {
	controllers.SharedFind
	Date string
}

// The share page of a public competition, /share/competition/<id>: standings, where species are found
// and every species on its list.
func (s *Site) shareCompetitionPage(c *gin.Context) {
	lang := language(c, []string{"sv", "en", "de"}, "sv")
	t := shareTexts[lang]
	data := gin.H{"Lang": lang, "Title": appNames[lang], "T": t}
	race, err := s.shareRace(c.Param("uuid"))
	if err != nil {
		s.render(c, http.StatusNotFound, "share.html", data)
		return
	}
	info := race.Info
	now := time.Now()
	fill := func(text string, v any) string { return strings.Replace(text, "{0}", fmt.Sprint(v), 1) }

	switch info.Status {
	case "upcoming":
		data["Status"] = fill(t["StartsOn"], shareDate(info.From, lang))
	case "finished":
		// A competition that ran its time ends as its last day does.
		ended := info.EndedAt.Time
		if !ended.Before(info.Until) {
			ended = info.Until.Add(-time.Second)
		}
		data["Status"] = fill(t["FinishedOn"], shareDate(ended, lang))
		if len(info.Standings) > 0 && info.Standings[0].Count > 0 {
			data["Winner"] = info.Standings[0].User
		}
	default:
		days := int(info.Until.Sub(now).Hours() / 24)
		if days < 1 {
			data["Status"] = t["LastDay"]
		} else {
			data["Status"] = fill(t["DaysLeft"], days)
		}
	}
	last := info.Until.Add(-time.Second)
	if info.From.Year() == last.Year() {
		data["Period"] = fmt.Sprintf("%d %s – %d %s %d", info.From.Day(), shareMonths[lang][info.From.Month()-1], last.Day(), shareMonths[lang][last.Month()-1], last.Year())
	} else {
		data["Period"] = shareDate(info.From, lang) + " – " + shareDate(last, lang)
	}
	if info.EndOnReports.Int64 > 0 {
		data["Goal"] = fill(t["FirstTo"], info.EndOnReports.Int64)
	}
	data["CanJoin"] = info.Status != "finished" && now.Before(info.JoinUntil) && (info.MaxCompetitors.Int64 == 0 || int64(info.Participants) < info.MaxCompetitors.Int64)

	// Bars measure against the winning count, the list, or the leader.
	target := int(info.EndOnReports.Int64)
	if target == 0 {
		target = int(info.ListSize)
	}
	if target == 0 && len(info.Standings) > 0 {
		target = info.Standings[0].Count
	}
	var standings []shareStanding
	for _, st := range info.Standings {
		e := shareStanding{CompetitionStanding: st}
		if target > 0 {
			e.Percent = min(100, st.Count*100/target)
		}
		if st.Latest.Valid {
			e.Latest = shareDate(st.Latest.Time, lang)
		}
		standings = append(standings, e)
	}
	data["Standings"] = standings

	recent := map[string]bool{}
	var finds []shareFind
	for _, f := range race.Latest {
		recent[f.Uuid] = true
		finds = append(finds, shareFind{SharedFind: f, Date: shareDay(f.At, lang)})
	}
	data["Finds"] = finds
	var groups []*shareRaceGroup
	found := 0
	for _, sp := range race.Species {
		if len(groups) == 0 || groups[len(groups)-1].Order.Id != sp.Order.Id {
			groups = append(groups, &shareRaceGroup{Order: sp.Order})
		}
		g := groups[len(groups)-1]
		e := shareRaceSpecies{Uuid: sp.Uuid, Name: sp.Name, Finders: sp.Finders, Recent: recent[sp.Uuid]}
		if info.Participants > 0 {
			e.Share = sp.Finders * 100 / info.Participants
		}
		e.Tip = strings.NewReplacer("{0}", strconv.Itoa(sp.Finders), "{1}", strconv.Itoa(info.Participants)).Replace(t["TakenBy"])
		if sp.Finders > 0 {
			e.Date = shareDay(sp.First, lang)
			g.Found++
			found++
		}
		g.Species = append(g.Species, e)
	}
	data["Groups"] = groups
	data["Found"] = found
	data["Total"] = len(race.Species)
	data["Race"] = race
	data["Color"] = info.Color
	data["Join"] = "/competitions?c=" + info.Uuid
	data["Tiles"] = s.cfg.MapTiles
	data["TilesAttribution"] = s.cfg.MapAttribution
	s.render(c, http.StatusOK, "share-competition.html", data)
}

// shareCompetitionAreas is the map's data: zones with how many species were found there.
func (s *Site) shareCompetitionAreas(c *gin.Context) {
	areas, err := s.areas(c.Param("uuid"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "not found"})
		return
	}
	// Fetched again on every live update, so never from a cache.
	c.Header("Cache-Control", "no-cache")
	c.JSON(http.StatusOK, gin.H{"areas": areas})
}
