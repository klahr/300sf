// SPDX-License-Identifier: GPL-2.0-or-later
package wikipedia

import "testing"

func TestLanguage(t *testing.T) {
	for in, want := range map[string]string{
		"sv": "sv", "en": "en", "de": "de",
		"": "en", "la": "en", "fr": "en",
		"internal-host/x?": "en",
	} {
		if got := Language(in); got != want {
			t.Errorf("Language(%q) = %q, want %q", in, got, want)
		}
	}
}
