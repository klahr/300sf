// SPDX-License-Identifier: AGPL-3.0-or-later
package controllers

import (
	"testing"

	"github.com/klahr/300sf/server/models"
	"gopkg.in/guregu/null.v4/zero"
)

func TestTranslationTextOnlyChangesSentText(t *testing.T) {
	sent := models.Translation{Model: models.Model{Uuid: "someone-elses"}, En: zero.StringFrom("Garden")}
	got := translationText(sent)
	if len(got) != 1 || got["en"] != zero.StringFrom("Garden") {
		t.Errorf("translationText = %v, want only en", got)
	}
	if _, ok := got["uuid"]; ok {
		t.Error("uuid must not be updated")
	}
}
