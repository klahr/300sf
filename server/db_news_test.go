// SPDX-License-Identifier: AGPL-3.0-or-later
package main

import (
	"net/http"
	"testing"
	"time"
)

func newPost(t *testing.T, s *dbServer, admin *testUser, title string, extra map[string]any) string {
	t.Helper()
	body := map[string]any{
		"title": map[string]any{"en": title, "sv": title + " (sv)"},
		"body":  map[string]any{"en": "**Hello** [300sf](https://300sf.se)"},
		"color": "#1273a3", "icon": "campaign",
	}
	for k, v := range extra {
		body[k] = v
	}
	r := s.must(t, admin, http.StatusOK, "POST", "/v1/announcements", body)
	return get(r.Body, "data", "uuid").(string)
}

// mine keeps the posts made by this test, as others may be in the feed.
func mine(posts any, ours ...string) []string {
	var out []string
	for _, u := range uuids(posts) {
		if contains(ours, u) {
			out = append(out, u)
		}
	}
	return out
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestDBNewsFeed(t *testing.T) {
	s := testDB(t)
	admin, reader := s.signUpAdmin(t), s.signUp(t)
	day := 24 * time.Hour
	at := func(d time.Duration) string { return time.Now().Add(d).UTC().Format(time.RFC3339) }

	s.must(t, reader, http.StatusUnauthorized, "POST", "/v1/announcements", map[string]any{"title": map[string]any{"en": "x"}, "body": map[string]any{"en": "y"}, "color": "#1273a3"})
	s.must(t, admin, http.StatusBadRequest, "POST", "/v1/announcements", map[string]any{"title": map[string]any{"sv": "bara svenska"}, "body": map[string]any{"sv": "x"}, "color": "#1273a3"})
	s.must(t, admin, http.StatusBadRequest, "POST", "/v1/announcements", map[string]any{"title": map[string]any{"en": "x"}, "body": map[string]any{"en": "y"}, "color": "red;x"})
	s.must(t, admin, http.StatusBadRequest, "POST", "/v1/announcements", map[string]any{"title": map[string]any{"en": "x"}, "body": map[string]any{"en": "y"}, "color": "#1273a3", "from": at(day), "until": at(0)})

	published := newPost(t, s, admin, "Published", nil)
	draft := newPost(t, s, admin, "Draft", map[string]any{"draft": true})
	scheduled := newPost(t, s, admin, "Scheduled", map[string]any{"from": at(day)})
	ended := newPost(t, s, admin, "Ended", map[string]any{"from": at(-2 * day), "until": at(-day)})
	all := []string{published, draft, scheduled, ended}

	unread := func() []string {
		return mine(get(s.must(t, reader, http.StatusOK, "GET", "/v1/announcements/list", nil).Body, "data"), all...)
	}
	if got := unread(); !equal(got, []string{published}) {
		t.Errorf("unread = %v, want only the published post", got)
	}
	feed := s.must(t, reader, http.StatusOK, "GET", "/v1/announcements/feed", nil)
	if got := mine(feed.Body["data"], all...); !equal(got, []string{published}) {
		t.Errorf("feed = %v, want only the published post", got)
	}

	s.must(t, admin, http.StatusOK, "PUT", "/v1/announcements/"+draft+"/resume", nil)
	if got := unread(); !equal(got, []string{draft, published}) {
		t.Errorf("after publishing the draft, unread = %v", got)
	}

	s.must(t, reader, http.StatusOK, "PUT", "/v1/announcements/"+published+"/read", nil)
	s.must(t, reader, http.StatusOK, "PUT", "/v1/announcements/"+published+"/read", nil)
	if got := unread(); !equal(got, []string{draft}) {
		t.Errorf("after reading one, unread = %v", got)
	}
	s.must(t, reader, http.StatusOK, "PUT", "/v1/announcements/read_all", nil)
	if got := unread(); len(got) != 0 {
		t.Errorf("after reading all, unread = %v", got)
	}
	feed = s.must(t, reader, http.StatusOK, "GET", "/v1/announcements/feed", nil)
	for _, p := range items(feed.Body["data"]) {
		if contains(all, get(p, "uuid").(string)) && get(p, "read") != true {
			t.Errorf("read post not marked read in the feed: %v", p)
		}
	}

	adminList := s.must(t, admin, http.StatusOK, "GET", "/v1/announcements/all", nil)
	reads := map[string]float64{}
	for _, p := range items(get(adminList.Body, "data")) {
		reads[get(p, "uuid").(string)] = get(p, "reads").(float64)
	}
	if len(mine(get(adminList.Body, "data"), all...)) != 4 || reads[published] != 1 || reads[draft] != 1 {
		t.Errorf("admin list: reads %v", reads)
	}

	// Editing replaces the texts; a language left out is cleared.
	s.must(t, admin, http.StatusOK, "PUT", "/v1/announcements/"+published, map[string]any{
		"title": map[string]any{"en": "Edited"}, "body": map[string]any{"en": "New text"}, "color": "#157975", "icon": "info",
	})
	feed = s.must(t, reader, http.StatusOK, "GET", "/v1/announcements/feed", nil)
	for _, p := range items(feed.Body["data"]) {
		if get(p, "uuid") == published && (get(p, "title", "en") != "Edited" || get(p, "title", "sv") != "" || get(p, "icon") != "info") {
			t.Errorf("edited post = %v", p)
		}
	}

	s.must(t, admin, http.StatusOK, "PUT", "/v1/announcements/"+published+"/pause", nil)
	feed = s.must(t, reader, http.StatusOK, "GET", "/v1/announcements/feed", nil)
	if contains(uuids(feed.Body["data"]), published) {
		t.Error("taken down post still in the feed")
	}
	s.must(t, admin, http.StatusNotFound, "PUT", "/v1/announcements/nonexistent/resume", nil)
	s.must(t, admin, http.StatusOK, "DELETE", "/v1/announcements/"+draft, nil)
}

func TestDBNewsPaging(t *testing.T) {
	s := testDB(t)
	admin, reader := s.signUpAdmin(t), s.signUp(t)
	var ours []string
	for i := 0; i < 5; i++ {
		ours = append(ours, newPost(t, s, admin, "Page post", nil))
	}

	// Every post exactly once, newest first, however the pages fall.
	seen := map[string]int{}
	var order []string
	before := ""
	for page := 0; page < 1000; page++ {
		path := "/v1/announcements/feed?limit=2"
		if before != "" {
			path += "&before=" + before
		}
		r := s.must(t, reader, http.StatusOK, "GET", path, nil)
		got := uuids(r.Body["data"])
		for _, u := range got {
			seen[u]++
			if contains(ours, u) {
				order = append(order, u)
			}
		}
		if r.Body["more"] != true {
			break
		}
		before = got[len(got)-1]
	}
	for u, n := range seen {
		if n != 1 {
			t.Errorf("post %s came %d times", u, n)
		}
	}
	want := []string{ours[4], ours[3], ours[2], ours[1], ours[0]}
	if !equal(order, want) {
		t.Errorf("order = %v, want newest first %v", order, want)
	}

	// Posts from before an account was made don't count as unread for it.
	// Times are stored to the second, so let one pass.
	time.Sleep(1100 * time.Millisecond)
	late := s.signUp(t)
	if got := mine(get(s.must(t, late, http.StatusOK, "GET", "/v1/announcements/list", nil).Body, "data"), ours...); len(got) != 0 {
		t.Errorf("a new account has old posts unread: %v", got)
	}
}
