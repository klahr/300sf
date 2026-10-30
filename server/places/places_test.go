// SPDX-License-Identifier: AGPL-3.0-or-later
package places

import (
	"testing"

	"github.com/klahr/300sf/server/geocoder"
	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/types"
)

func TestComplete(t *testing.T) {
	zone := func(level int64) models.Zone { return models.Zone{AdminLevel: level} }
	if Complete(nil) {
		t.Error("no zones is not complete")
	}
	if Complete([]models.Zone{zone(types.AdminLevelCountry), zone(types.AdminLevelCounty)}) {
		t.Error("down to the county is not complete")
	}
	if !Complete([]models.Zone{zone(types.AdminLevelCountry), zone(types.AdminLevelMunicipality)}) {
		t.Error("a municipality straight under its country is complete")
	}
}

func TestPlaceKey(t *testing.T) {
	if placeKey(59.8586, 17.6389) != placeKey(59.8612, 17.6411) {
		t.Error("points a few hundred metres apart should share a key")
	}
	if placeKey(59.86, 17.64) == placeKey(59.90, 17.64) {
		t.Error("points kilometres apart should not")
	}
}

func TestNames(t *testing.T) {
	area := geocoder.Area{Name: "Uppsala län", Names: map[string]string{"name": "Uppsala län", "name:en": "Uppsala County", "name:sv": "Uppsala län"}}
	got := names(area)
	if len(got) != 2 || got[0] != "Uppsala län" || got[1] != "Uppsala County" {
		t.Errorf("names = %v, want each name once", got)
	}
}
