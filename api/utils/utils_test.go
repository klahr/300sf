// SPDX-License-Identifier: GPL-2.0-or-later
package utils

import (
	"regexp"
	"testing"
	"time"

	"github.com/klahr/300sf/api/models"
	"gopkg.in/guregu/null.v4/zero"
)

func TestConversions(t *testing.T) {
	if got := Str2Int("42"); got != 42 {
		t.Errorf("Str2Int(42) = %d", got)
	}
	if got := Str2Int(""); got != 0 {
		t.Errorf("Str2Int(\"\") = %d, want 0 (missing redis counter)", got)
	}
	if got := Str2Float("1.5"); got != 1.5 {
		t.Errorf("Str2Float(1.5) = %v", got)
	}
	if got := Str2Time("2024-05-12"); !got.Equal(time.Date(2024, 5, 12, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("Str2Time = %v", got)
	}
	if got := Str2Time("not a date"); !got.IsZero() {
		t.Errorf("Str2Time(invalid) = %v, want zero", got)
	}
}

func TestPassword(t *testing.T) {
	var p Password
	if err := p.Create("hunter22"); err != nil {
		t.Fatal(err)
	}
	if len(p.Salt) != 40 {
		t.Errorf("salt length = %d, want 40", len(p.Salt))
	}
	if !p.IsMatching("hunter22", p.Value, p.Salt) {
		t.Error("correct password did not match")
	}
	if p.IsMatching("hunter23", p.Value, p.Salt) {
		t.Error("wrong password matched")
	}
	if p.IsMatching("hunter22", p.Value, RandomSha1()) {
		t.Error("password matched with wrong salt")
	}
}

func TestRandomHex(t *testing.T) {
	hex := regexp.MustCompile(`^[0-9a-f]+$`)
	a, b := RandomSha256(), RandomSha256()
	if len(a) != 64 || !hex.MatchString(a) {
		t.Errorf("RandomSha256 = %q", a)
	}
	if a == b {
		t.Error("RandomSha256 returned the same value twice")
	}
}

func TestComputeHmac256(t *testing.T) {
	// RFC 4231 test case 2.
	got := ComputeHmac256("what do ya want for nothing?", "Jefe")
	want := "5bdcc146bf60754e6a042426089575c75a003f089d2739839dec58b964ec3843"
	if got != want {
		t.Errorf("ComputeHmac256 = %s, want %s", got, want)
	}
}

func TestGenerateUuidWithFixedLength(t *testing.T) {
	valid := regexp.MustCompile(`^[abcdefghjkmnpqrstuvwxyz1-9]{6}$`)
	for i := 0; i < 100; i++ {
		if s := GenerateShortUuid(); !valid.MatchString(s) {
			t.Fatalf("GenerateShortUuid = %q", s)
		}
	}
	if s := GenerateUuidWithFixedLength(32); len(s) != 32 {
		t.Errorf("length = %d, want 32", len(s))
	}
}

func TestVerify(t *testing.T) {
	cases := []struct {
		name string
		fn   func(string) bool
		in   string
		want bool
	}{
		{"display name ok", VerifyDisplayName, "Kalle", true},
		{"display name short", VerifyDisplayName, "Kal", false},
		{"username ok", VerifyUsername, "kalle.anka", true},
		{"username short", VerifyUsername, "abc", false},
		{"username bad char", VerifyUsername, "kalle anka", false},
		{"username trailing dot", VerifyUsername, "kalle.", false},
		{"password ok", VerifyPassword, "abcd", true},
		{"password short", VerifyPassword, "abc", false},
	}
	for _, c := range cases {
		if got := c.fn(c.in); got != c.want {
			t.Errorf("%s: f(%q) = %v, want %v", c.name, c.in, got, c.want)
		}
	}
}

func TestTranslation2Value(t *testing.T) {
	tr := models.Translation{
		Sv: zero.StringFrom("Talgoxe"),
		En: zero.StringFrom("Great tit"),
		La: zero.StringFrom("Parus major"),
	}
	for lang, want := range map[string]string{"sv": "Talgoxe", "en": "Great tit", "la": "Parus major", "de": "", "fr": ""} {
		if got := Translation2Value(tr, lang); got != want {
			t.Errorf("Translation2Value(%s) = %q, want %q", lang, got, want)
		}
	}
}

func TestEmailTemplatesComplete(t *testing.T) {
	t.Chdir("..")
	for _, lang := range []string{"en", "sv", "de"} {
		for _, name := range []string{"welcome", "invite", "set_password", "delete_account"} {
			e, err := GetEmailTemplate(name, lang)
			if err != nil {
				t.Errorf("%s/%s: %v", lang, name, err)
				continue
			}
			if e.Subject == "" || (e.Txt == "" && e.Html == "") {
				t.Errorf("%s/%s: missing subject or body", lang, name)
			}
		}
	}
}
