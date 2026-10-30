// SPDX-License-Identifier: AGPL-3.0-or-later
package main

// The tools /mcp offers (see mcp.go). Each calls the API and gives back what
// an assistant needs, in fewer words than the API: names spelled out next to
// ids, and nothing empty.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const mcpInstructions = `300 Svenska Fåglar ("300 Swedish Birds") is where birders in Sweden keep checklists of the species they have seen.
A sighting (report) is one species seen at a time and place. A checklist (list) collects sightings, often for a year or a region, and may have target species. Labels (tags) group sightings. Species have Swedish, English and Latin names.
Find a species' uuid with search_species before adding a sighting. Dates are YYYY-MM-DD or RFC 3339.`

// newMcpServer is the server for one request, with the tools its token may use.
func newMcpServer(a *apiCaller) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "300sf", Title: "300 Svenska Fåglar", Version: version}, &mcp.ServerOptions{Instructions: mcpInstructions})
	addReadTools(s, a)
	if a.canChange() {
		addChangeTools(s, a)
	}
	return s
}

type toolKind int

const (
	reads toolKind = iota
	changes
	deletes
)

// addTool adds a tool whose answer is JSON for the assistant to read.
func addTool[In any](s *mcp.Server, kind toolKind, name, title, description string, run func(context.Context, In) (any, error)) {
	no, yes := false, true
	annotations := &mcp.ToolAnnotations{Title: title, ReadOnlyHint: kind == reads, OpenWorldHint: &no}
	if kind != reads {
		annotations.DestructiveHint = &no
		if kind == deletes {
			annotations.DestructiveHint = &yes
			annotations.IdempotentHint = true
		}
	}
	mcp.AddTool(s, &mcp.Tool{Name: name, Title: title, Description: description, Annotations: annotations},
		func(ctx context.Context, _ *mcp.CallToolRequest, in In) (*mcp.CallToolResult, any, error) {
			out, err := run(ctx, in)
			if err != nil {
				return nil, nil, err
			}
			text, err := json.Marshal(out)
			if err != nil {
				return nil, nil, err
			}
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(text)}}}, nil, nil
		})
}

// Helpers for reading the API's decoded JSON.

func field(v any, path ...string) any {
	for _, k := range path {
		m, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		v = m[k]
	}
	return v
}

func str(v any, path ...string) string {
	s, _ := field(v, path...).(string)
	return s
}

func array(v any, path ...string) []any {
	a, _ := field(v, path...).([]any)
	return a
}

// nameOf spells out a translated name: Swedish / English / Latin.
func nameOf(t any) string {
	var parts []string
	for _, lang := range []string{"sv", "en", "la"} {
		if s := str(t, lang); s != "" && !slices.Contains(parts, s) {
			parts = append(parts, s)
		}
	}
	if len(parts) == 0 {
		return str(t, "de")
	}
	return strings.Join(parts, " / ")
}

// day is the date of an API time, empty when unset.
func day(t string) string {
	if t == "" || strings.HasPrefix(t, "0001-01-01") {
		return ""
	}
	if len(t) >= 10 {
		return t[:10]
	}
	return t
}

func when(t string) string {
	if day(t) == "" {
		return ""
	}
	return t
}

// slim leaves out what says nothing: empty values, zero times, map outlines
// and the ids of translations.
func slim(v any) any {
	switch t := v.(type) {
	case map[string]any:
		translation := true
		for k := range t {
			if !slices.Contains([]string{"uuid", "sv", "en", "de", "la"}, k) {
				translation = false
			}
		}
		out := map[string]any{}
		for k, val := range t {
			if k == "polygon" || k == "timestamp" || (translation && k == "uuid") {
				continue
			}
			if s, ok := val.(string); ok && (s == "" || strings.HasPrefix(s, "0001-01-01")) {
				continue
			}
			if val = slim(val); val != nil {
				out[k] = val
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	case []any:
		if len(t) == 0 {
			return nil
		}
		out := make([]any, 0, len(t))
		for _, val := range t {
			out = append(out, slim(val))
		}
		return out
	}
	return v
}

// sighting is a report in short.
func sighting(r any) map[string]any {
	out := map[string]any{
		"uuid":         str(r, "uuid"),
		"species":      nameOf(field(r, "species", "name")),
		"species_uuid": str(r, "species", "uuid"),
		"from":         when(str(r, "from")),
		"until":        when(str(r, "until")),
		"site":         str(r, "site"),
		"place":        nameOf(field(r, "zone", "name")),
		"comment":      str(r, "comment", "text"),
	}
	if q, _ := field(r, "quantity").(float64); q > 0 {
		out["quantity"] = q
	}
	lat, _ := field(r, "latitude").(float64)
	lon, _ := field(r, "longitude").(float64)
	if lat != 0 || lon != 0 {
		out["latitude"], out["longitude"] = lat, lon
	}
	for _, k := range []string{"age", "gender", "method", "activity"} {
		if n := nameOf(field(r, k, "name")); n != "" {
			out[k] = n
		}
	}
	var lists, tags []string
	for _, l := range array(r, "lists") {
		lists = append(lists, nameOf(field(l, "name"))+" ("+str(l, "uuid")+")")
	}
	for _, t := range array(r, "tags") {
		tags = append(tags, str(t, "name")+" ("+str(t, "uuid")+")")
	}
	if lists != nil {
		out["checklists"] = lists
	}
	if tags != nil {
		out["labels"] = tags
	}
	for k, v := range out {
		if v == "" {
			delete(out, k)
		}
	}
	return out
}

func checklist(l any) map[string]any {
	out := map[string]any{
		"uuid":           str(l, "uuid"),
		"name":           nameOf(field(l, "name")),
		"target_species": len(array(l, "species")),
	}
	if v, ok := field(l, "visibility").(float64); ok && v >= 0 && v <= 2 {
		out["visibility"] = []string{"private", "hidden", "public"}[int(v)]
	}
	if d := day(str(l, "from")); d != "" {
		out["from"] = d
	}
	if d := day(str(l, "until")); d != "" {
		out["until"] = d
	}
	return out
}

func speciesInShort(sp any) map[string]any {
	out := map[string]any{"uuid": str(sp, "uuid"), "name": nameOf(field(sp, "name"))}
	if o := nameOf(field(sp, "order", "name")); o != "" {
		out["order"] = o
	}
	return out
}

func listPath(uuid string) string {
	return url.PathEscape(uuid)
}

type uuidInput struct {
	Uuid string `json:"uuid" jsonschema:"the uuid"`
}

func addReadTools(s *mcp.Server, a *apiCaller) {
	addTool(s, reads, "whoami", "Who am I", "The signed-in user, and whether this connection may change things or only read.",
		func(ctx context.Context, _ struct{}) (any, error) {
			r, err := a.call(ctx, "GET", "/v1/users/whoami", nil, "")
			if err != nil {
				return nil, err
			}
			return map[string]any{
				"display_name": str(r, "data", "display_name"),
				"username":     str(r, "data", "username"),
				"language":     str(r, "data", "language", "code"),
				"can_change":   a.canChange(),
			}, nil
		})

	type listsInput struct {
		Archived bool `json:"archived,omitempty" jsonschema:"list the archived checklists instead of the active ones"`
	}
	addTool(s, reads, "list_checklists", "List checklists", "The user's checklists, with how many target species each has (0 means any species counts).",
		func(ctx context.Context, in listsInput) (any, error) {
			path := "/v1/lists/list"
			if in.Archived {
				path = "/v1/lists/archived"
			}
			r, err := a.call(ctx, "GET", path, nil, "")
			if err != nil {
				return nil, err
			}
			out := []any{}
			for _, l := range array(r, "result") {
				out = append(out, checklist(l))
			}
			return out, nil
		})

	addTool(s, reads, "get_checklist", "Checklist progress",
		"One checklist: each species seen on it (first and last date, number of sightings) and, when it has target species, those not seen yet. Works for the user's own lists and other users' shared ones (then without sightings).",
		func(ctx context.Context, in uuidInput) (any, error) {
			r, err := a.call(ctx, "GET", "/v1/lists/"+listPath(in.Uuid), nil, "")
			if err != nil {
				return nil, err
			}
			list := field(r, "result")
			out := map[string]any{"checklist": checklist(list)}

			reports, err := a.call(ctx, "GET", "/v1/lists/"+listPath(in.Uuid)+"/reports", nil, "")
			if err != nil {
				out["sightings"] = "not available: " + err.Error()
				return out, nil
			}
			type seen struct {
				Uuid      string `json:"species_uuid"`
				Name      string `json:"species"`
				First     string `json:"first_seen"`
				Last      string `json:"last_seen"`
				Sightings int    `json:"sightings"`
			}
			bySpecies := map[string]*seen{}
			for _, rep := range array(reports, "result") {
				id := str(rep, "species", "uuid")
				d := day(str(rep, "from"))
				e := bySpecies[id]
				if e == nil {
					e = &seen{Uuid: id, Name: nameOf(field(rep, "species", "name")), First: d, Last: d}
					bySpecies[id] = e
				}
				e.Sightings++
				if d != "" && (e.First == "" || d < e.First) {
					e.First = d
				}
				if d > e.Last {
					e.Last = d
				}
			}
			seenList := make([]*seen, 0, len(bySpecies))
			for _, e := range bySpecies {
				seenList = append(seenList, e)
			}
			sort.Slice(seenList, func(i, j int) bool { return seenList[i].First < seenList[j].First })
			out["species_seen"] = len(seenList)
			out["seen"] = seenList

			if targets := array(list, "species"); len(targets) > 0 {
				missing := []any{}
				for _, sp := range targets {
					if bySpecies[str(sp, "uuid")] == nil {
						missing = append(missing, speciesInShort(sp))
					}
				}
				out["not_seen_yet"] = missing
			}
			return out, nil
		})

	type sightingsInput struct {
		From      string `json:"from,omitempty" jsonschema:"only sightings on or after this date, YYYY-MM-DD"`
		Until     string `json:"until,omitempty" jsonschema:"only sightings on or before this date, YYYY-MM-DD"`
		Species   string `json:"species,omitempty" jsonschema:"only this species: a uuid, or part of its Swedish, English or Latin name"`
		Checklist string `json:"checklist_uuid,omitempty" jsonschema:"only sightings on this checklist"`
		Limit     int    `json:"limit,omitempty" jsonschema:"at most this many, newest first (default 50, at most 500)"`
	}
	addTool(s, reads, "list_sightings", "Find sightings", "The user's sightings, newest first, optionally only within dates, of a species or on a checklist.",
		func(ctx context.Context, in sightingsInput) (any, error) {
			r, err := a.call(ctx, "GET", "/v1/reports/list", nil, "")
			if err != nil {
				return nil, err
			}
			limit := in.Limit
			if limit <= 0 {
				limit = 50
			}
			limit = min(limit, 500)
			query := strings.ToLower(in.Species)
			found := []any{}
			total := 0
			for _, rep := range array(r, "result") {
				d := day(str(rep, "from"))
				if (in.From != "" && d < in.From) || (in.Until != "" && (d == "" || d > in.Until)) {
					continue
				}
				if query != "" && str(rep, "species", "uuid") != in.Species && !strings.Contains(strings.ToLower(nameOf(field(rep, "species", "name"))), query) {
					continue
				}
				if in.Checklist != "" && !slices.ContainsFunc(array(rep, "lists"), func(l any) bool { return str(l, "uuid") == in.Checklist }) {
					continue
				}
				total++
				if len(found) < limit {
					found = append(found, sighting(rep))
				}
			}
			return map[string]any{"matching": total, "returned": len(found), "sightings": found}, nil
		})

	addTool(s, reads, "get_sighting", "Get a sighting", "One of the user's sightings in full.",
		func(ctx context.Context, in uuidInput) (any, error) {
			r, err := a.call(ctx, "GET", "/v1/reports/"+url.PathEscape(in.Uuid), nil, "")
			if err != nil {
				return nil, err
			}
			return sighting(field(r, "result")), nil
		})

	type searchInput struct {
		Query string `json:"query" jsonschema:"part of the species' Swedish, English or Latin name"`
		Limit int    `json:"limit,omitempty" jsonschema:"at most this many (default 20)"`
	}
	addTool(s, reads, "search_species", "Search species", "Finds species by name, with their uuids; best matches first.",
		func(ctx context.Context, in searchInput) (any, error) {
			query := strings.ToLower(strings.TrimSpace(in.Query))
			if query == "" {
				return nil, fmt.Errorf("query is empty")
			}
			r, err := a.call(ctx, "GET", "/v1/species/list", nil, "")
			if err != nil {
				return nil, err
			}
			type match struct {
				rank int
				sp   any
			}
			var matches []match
			for _, sp := range array(r, "result") {
				best := -1
				for _, lang := range []string{"sv", "en", "la", "de"} {
					n := strings.ToLower(str(sp, "name", lang))
					rank := -1
					switch {
					case n == query:
						rank = 0
					case strings.HasPrefix(n, query):
						rank = 1
					case strings.Contains(n, query):
						rank = 2
					}
					if rank >= 0 && (best < 0 || rank < best) {
						best = rank
					}
				}
				if best >= 0 || str(sp, "uuid") == in.Query {
					matches = append(matches, match{max(best, 0), sp})
				}
			}
			sort.SliceStable(matches, func(i, j int) bool { return matches[i].rank < matches[j].rank })
			limit := in.Limit
			if limit <= 0 {
				limit = 20
			}
			out := []any{}
			for _, m := range matches[:min(limit, len(matches))] {
				out = append(out, speciesInShort(m.sp))
			}
			return out, nil
		})

	type speciesInput struct {
		Uuid     string `json:"uuid" jsonschema:"the species' uuid"`
		Language string `json:"language,omitempty" jsonschema:"the Wikipedia edition: sv, en or de (default en)"`
	}
	addTool(s, reads, "get_species", "Species details", "A species with a summary from Wikipedia and a few sound recordings from xeno-canto.",
		func(ctx context.Context, in speciesInput) (any, error) {
			lang := in.Language
			if lang == "" {
				lang = "en"
			}
			r, err := a.call(ctx, "GET", "/v1/species/"+url.PathEscape(in.Uuid), nil, lang)
			if err != nil {
				return nil, err
			}
			out := speciesInShort(field(r, "result"))
			if w := field(r, "wikipedia"); w != nil {
				out["wikipedia"] = map[string]any{"url": str(w, "url"), "summary": str(w, "summary")}
			}
			var recordings []any
			for _, rec := range array(r, "recordings") {
				if len(recordings) == 5 {
					break
				}
				recordings = append(recordings, slim(map[string]any{
					"url": str(rec, "url"), "type": str(rec, "type"), "sex": str(rec, "sex"),
					"country": str(rec, "country"), "location": str(rec, "location"), "recorder": str(rec, "recorder"),
				}))
			}
			if recordings != nil {
				out["recordings"] = recordings
			}
			return out, nil
		})

	addTool(s, reads, "list_competitions", "List competitions", "Competitions the user created, joined or was invited to.",
		func(ctx context.Context, _ struct{}) (any, error) {
			r, err := a.call(ctx, "GET", "/v1/competitions/list/me", nil, "")
			if err != nil {
				return nil, err
			}
			return slim(r), nil
		})

	addTool(s, reads, "get_competition", "Competition standings", "A competition with its rules and standings.",
		func(ctx context.Context, in uuidInput) (any, error) {
			r, err := a.call(ctx, "GET", "/v1/competitions/"+url.PathEscape(in.Uuid), nil, "")
			if err != nil {
				return nil, err
			}
			return slim(r), nil
		})

	type compareInput struct {
		Uuids []string `json:"checklist_uuids" jsonschema:"1 to 20 shared checklists; the first is the one compared against"`
	}
	addTool(s, reads, "compare_checklists", "Compare checklists",
		"Which species each of several shared (not private) checklists has, and which species of the first each one lacks.",
		func(ctx context.Context, in compareInput) (any, error) {
			r, err := a.call(ctx, "POST", "/v1/lists/compare", map[string]any{"uuids": in.Uuids}, "")
			if err != nil {
				return nil, err
			}
			entries := array(r, "result")
			if len(entries) == 0 {
				return nil, fmt.Errorf("none of the checklists were found, or they are private")
			}
			speciesOf := func(e any) map[string]string {
				m := map[string]string{}
				for _, rep := range array(e, "reports") {
					m[str(rep, "species", "uuid")] = nameOf(field(rep, "species", "name"))
				}
				return m
			}
			reference := speciesOf(entries[0])
			out := []any{}
			for _, e := range entries {
				own := speciesOf(e)
				var names, lacks []string
				for _, n := range own {
					names = append(names, n)
				}
				for id, n := range reference {
					if _, ok := own[id]; !ok {
						lacks = append(lacks, n)
					}
				}
				sort.Strings(names)
				sort.Strings(lacks)
				out = append(out, slim(map[string]any{
					"checklist": nameOf(field(e, "list", "name")), "uuid": str(e, "list", "uuid"),
					"owner": str(e, "user", "display_name"), "species_seen": len(own),
					"species": names, "lacks_from_first": lacks,
				}))
			}
			return out, nil
		})

	addTool(s, reads, "list_labels", "List labels", "The user's labels (tags), which group sightings.",
		func(ctx context.Context, _ struct{}) (any, error) {
			r, err := a.call(ctx, "GET", "/v1/tags/list", nil, "")
			if err != nil {
				return nil, err
			}
			return slim(field(r, "result")), nil
		})

	addTool(s, reads, "list_checklist_templates", "List checklist templates",
		"Species lists a new checklist can start from: the user's own and public ones.",
		func(ctx context.Context, _ struct{}) (any, error) {
			r, err := a.call(ctx, "GET", "/v1/list_templates/list", nil, "")
			if err != nil {
				return nil, err
			}
			out := []any{}
			for _, t := range array(r, "result") {
				out = append(out, slim(map[string]any{
					"uuid": str(t, "uuid"), "name": nameOf(field(t, "name")), "description": nameOf(field(t, "description")),
					"species": len(array(t, "species")),
				}))
			}
			return out, nil
		})

	addTool(s, reads, "list_sighting_options", "Sighting options",
		"The ages, sexes, methods and activities a sighting can have, with the uuids add_sighting and update_sighting take.",
		func(ctx context.Context, _ struct{}) (any, error) {
			out := map[string]any{}
			for kind, path := range map[string]string{"ages": "/v1/ages/list", "genders": "/v1/genders/list", "methods": "/v1/methods/list", "activities": "/v1/activities/list"} {
				r, err := a.call(ctx, "GET", path, nil, "")
				if err != nil {
					return nil, err
				}
				var options []any
				for _, o := range array(r, "result") {
					options = append(options, map[string]any{"uuid": str(o, "uuid"), "name": nameOf(field(o, "name"))})
				}
				out[kind] = options
			}
			return out, nil
		})
}

// sightingFields are what a sighting is made of, in add_sighting and
// update_sighting. Left out, a field is not set (or, for update, kept).
type sightingFields struct {
	From       *string   `json:"from,omitempty" jsonschema:"when: YYYY-MM-DD for a date alone, or RFC 3339 with time and zone"`
	Until      *string   `json:"until,omitempty" jsonschema:"when it ended, like from, for a sighting over a while"`
	Quantity   *int64    `json:"quantity,omitempty" jsonschema:"how many birds"`
	Latitude   *float64  `json:"latitude,omitempty"`
	Longitude  *float64  `json:"longitude,omitempty"`
	Site       *string   `json:"site,omitempty" jsonschema:"the place's name, in the user's words"`
	Comment    *string   `json:"comment,omitempty" jsonschema:"a private note"`
	Checklists *[]string `json:"checklist_uuids,omitempty" jsonschema:"the user's checklists it counts on"`
	Labels     *[]string `json:"label_uuids,omitempty" jsonschema:"the user's labels"`
	Age        *string   `json:"age_uuid,omitempty" jsonschema:"from list_sighting_options"`
	Gender     *string   `json:"gender_uuid,omitempty" jsonschema:"from list_sighting_options"`
	Method     *string   `json:"method_uuid,omitempty" jsonschema:"from list_sighting_options"`
	Activity   *string   `json:"activity_uuid,omitempty" jsonschema:"from list_sighting_options"`
}

// apply puts the fields that are given into body, the API's ReportFields.
func (f sightingFields) apply(body map[string]any) {
	set := func(key string, v any) {
		if s, ok := v.(string); ok && s == "" {
			delete(body, key)
			return
		}
		body[key] = v
	}
	if f.From != nil {
		set("from", *f.From)
	}
	if f.Until != nil {
		set("until", *f.Until)
	}
	if f.Quantity != nil {
		body["quantity"] = *f.Quantity
	}
	if f.Latitude != nil {
		body["latitude"] = *f.Latitude
	}
	if f.Longitude != nil {
		body["longitude"] = *f.Longitude
	}
	if f.Site != nil {
		set("site", *f.Site)
	}
	if f.Comment != nil {
		set("comment", *f.Comment)
	}
	if f.Checklists != nil {
		body["list_uuids"] = *f.Checklists
	}
	if f.Labels != nil {
		body["tag_uuids"] = *f.Labels
	}
	for key, v := range map[string]*string{"age_uuid": f.Age, "gender_uuid": f.Gender, "method_uuid": f.Method, "activity_uuid": f.Activity} {
		if v != nil {
			set(key, *v)
		}
	}
}

// reportFields turns a sighting from the API back into the ReportFields
// that make it, for editing.
func reportFields(r any) map[string]any {
	body := map[string]any{"species_uuid": str(r, "species", "uuid")}
	for _, k := range []string{"from", "until"} {
		if t := when(str(r, k)); t != "" {
			body[k] = t
		}
	}
	for _, k := range []string{"quantity", "latitude", "longitude"} {
		if v, ok := field(r, k).(float64); ok && v != 0 {
			body[k] = v
		}
	}
	if s := str(r, "site"); s != "" {
		body["site"] = s
	}
	if s := str(r, "comment", "text"); s != "" {
		body["comment"] = s
	}
	for _, k := range []string{"age", "gender", "method", "activity"} {
		if id := str(r, k, "uuid"); id != "" {
			body[k+"_uuid"] = id
		}
	}
	lists, tags := []string{}, []string{}
	for _, l := range array(r, "lists") {
		lists = append(lists, str(l, "uuid"))
	}
	for _, t := range array(r, "tags") {
		tags = append(tags, str(t, "uuid"))
	}
	body["list_uuids"], body["tag_uuids"] = lists, tags
	return body
}

func addChangeTools(s *mcp.Server, a *apiCaller) {
	type addInput struct {
		Species string `json:"species_uuid" jsonschema:"the species, from search_species"`
		sightingFields
	}
	addTool(s, changes, "add_sighting", "Add a sighting",
		"Adds a sighting of one species. Give at least the date (from); put it on checklists with checklist_uuids. The place (zone) is worked out from the position.",
		func(ctx context.Context, in addInput) (any, error) {
			if in.From == nil || *in.From == "" {
				today := time.Now().Format(time.DateOnly)
				in.From = &today
			}
			body := map[string]any{"species_uuid": in.Species, "silent": true}
			in.sightingFields.apply(body)
			r, err := a.call(ctx, "POST", "/v1/reports", body, "")
			if err != nil {
				return nil, err
			}
			return sighting(r), nil
		})

	type updateInput struct {
		Uuid    string  `json:"uuid" jsonschema:"the sighting"`
		Species *string `json:"species_uuid,omitempty" jsonschema:"another species, when it was misidentified"`
		sightingFields
	}
	addTool(s, changes, "update_sighting", "Change a sighting",
		"Changes the given fields of a sighting and keeps the rest. An empty string clears a text or uuid field; checklist_uuids and label_uuids replace the sighting's checklists and labels.",
		func(ctx context.Context, in updateInput) (any, error) {
			path := "/v1/reports/" + url.PathEscape(in.Uuid)
			r, err := a.call(ctx, "GET", path, nil, "")
			if err != nil {
				return nil, err
			}
			body := reportFields(field(r, "result"))
			if in.Species != nil && *in.Species != "" {
				body["species_uuid"] = *in.Species
			}
			in.sightingFields.apply(body)
			updated, err := a.call(ctx, "PUT", path, body, "")
			if err != nil {
				return nil, err
			}
			return sighting(updated), nil
		})

	addTool(s, deletes, "delete_sighting", "Delete a sighting", "Deletes a sighting for good, with its note. Ask the user first.",
		func(ctx context.Context, in uuidInput) (any, error) {
			if _, err := a.call(ctx, "DELETE", "/v1/reports/"+url.PathEscape(in.Uuid), nil, ""); err != nil {
				return nil, err
			}
			return map[string]any{"deleted": in.Uuid}, nil
		})

	type onListInput struct {
		Sighting  string `json:"sighting_uuid"`
		Checklist string `json:"checklist_uuid"`
	}
	for _, op := range []struct{ name, title, description, path string }{
		{"add_sighting_to_checklist", "Put a sighting on a checklist", "Puts one of the user's sightings on one of their checklists.", "add_list"},
		{"remove_sighting_from_checklist", "Take a sighting off a checklist", "Takes a sighting off a checklist; the sighting itself is kept.", "remove_list"},
	} {
		addTool(s, changes, op.name, op.title, op.description,
			func(ctx context.Context, in onListInput) (any, error) {
				if _, err := a.call(ctx, "POST", "/v1/reports/"+url.PathEscape(in.Sighting)+"/"+op.path, map[string]any{"list_uuid": in.Checklist}, ""); err != nil {
					return nil, err
				}
				return map[string]any{"done": true}, nil
			})
	}

	type createListInput struct {
		Name       string `json:"name" jsonschema:"the checklist's name"`
		From       string `json:"from,omitempty" jsonschema:"first day it counts sightings, YYYY-MM-DD"`
		Until      string `json:"until,omitempty" jsonschema:"last day, YYYY-MM-DD"`
		Template   string `json:"template_uuid,omitempty" jsonschema:"start with the target species of this template, from list_checklist_templates"`
		Visibility string `json:"visibility,omitempty" jsonschema:"private (default), hidden (anyone with the link) or public"`
	}
	addTool(s, changes, "create_checklist", "Create a checklist", "Makes a new checklist for the user.",
		func(ctx context.Context, in createListInput) (any, error) {
			visibility := slices.Index([]string{"private", "hidden", "public"}, in.Visibility)
			if in.Visibility == "" {
				visibility = 0
			}
			if visibility < 0 {
				return nil, fmt.Errorf("visibility is private, hidden or public")
			}
			body := map[string]any{
				"name":       map[string]string{"sv": in.Name, "en": in.Name, "de": in.Name},
				"visibility": visibility,
			}
			for k, v := range map[string]string{"from": in.From, "until": in.Until, "template": in.Template} {
				if v != "" {
					body[k] = v
				}
			}
			r, err := a.call(ctx, "POST", "/v1/lists", body, "")
			if err != nil {
				return nil, err
			}
			return checklist(r), nil
		})
}
