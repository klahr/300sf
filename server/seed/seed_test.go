// SPDX-License-Identifier: AGPL-3.0-or-later
package seed

import "testing"

func TestStarterData(t *testing.T) {
	orders, err := Orders()
	if err != nil {
		t.Fatal(err)
	}
	species, err := Species()
	if err != nil {
		t.Fatal(err)
	}
	if len(species) < 500 {
		t.Errorf("%d species, want every species observed in Sweden", len(species))
	}

	known := map[string]bool{}
	for _, o := range orders {
		if o.La == "" || o.Sv == "" || o.En == "" || o.De == "" {
			t.Errorf("order without a name: %+v", o)
		}
		known[o.La] = true
	}
	seen := map[string]bool{}
	for _, s := range species {
		// German names are missing for a few rarities; the app falls back to English.
		if s.La == "" || s.Sv == "" || s.En == "" {
			t.Errorf("species without a name: %+v", s)
		}
		if !known[s.Order] {
			t.Errorf("%s: unknown order %q", s.La, s.Order)
		}
		if seen[s.La] {
			t.Errorf("%s twice", s.La)
		}
		seen[s.La] = true
	}
	if !seen["Parus major"] || !seen["Cygnus olor"] {
		t.Error("common species missing")
	}
}
