// SPDX-License-Identifier: AGPL-3.0-or-later
package controllers

import (
	"strings"
	"testing"
	"time"

	"github.com/klahr/300sf/server/models"
	"gopkg.in/guregu/null.v4/zero"
)

func TestNewsRequestValidate(t *testing.T) {
	ok := func() newsRequest {
		return newsRequest{
			Title: models.LocalizedText{En: zero.StringFrom(" New: offline reporting ")},
			Body:  models.LocalizedText{En: zero.StringFrom("Report without a connection.")},
			Color: "#1273a3",
		}
	}
	r := ok()
	if err := r.validate(); err != nil {
		t.Fatal(err)
	}
	if r.Title.En.String != "New: offline reporting" || r.Icon != "campaign" {
		t.Errorf("not tidied: title %q, icon %q", r.Title.En.String, r.Icon)
	}

	now := time.Now()
	for name, change := range map[string]func(*newsRequest){
		"no English title": func(r *newsRequest) { r.Title.En = zero.StringFrom("  ") },
		"no English text":  func(r *newsRequest) { r.Body.En = zero.String{} },
		"long title":       func(r *newsRequest) { r.Title.Sv = zero.StringFrom(strings.Repeat("å", maxNewsTitle+1)) },
		"long text":        func(r *newsRequest) { r.Body.De = zero.StringFrom(strings.Repeat("x", maxNewsBody+1)) },
		"icon markup":      func(r *newsRequest) { r.Icon = "<b>x</b>" },
		"color markup":     func(r *newsRequest) { r.Color = "red;background:url(x)" },
		"ends first": func(r *newsRequest) {
			r.From = zero.TimeFrom(now)
			r.Until = zero.TimeFrom(now.Add(-time.Hour))
		},
	} {
		r := ok()
		change(&r)
		if r.validate() == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

func TestNewsPostPublishedAt(t *testing.T) {
	made := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	a := models.Announcement{Model: models.Model{Timestamp: made}}
	if got := newsPost(a, false).PublishedAt; !got.Equal(made) {
		t.Errorf("without from: %v", got)
	}
	from := made.Add(48 * time.Hour)
	a.From = zero.TimeFrom(from)
	if got := newsPost(a, true); !got.PublishedAt.Equal(from) || !got.Read {
		t.Errorf("with from: %+v", got)
	}
}
