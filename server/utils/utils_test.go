// SPDX-License-Identifier: AGPL-3.0-or-later
package utils

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/klahr/300sf/server/models"
	"golang.org/x/crypto/bcrypt"
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
	if !p.IsMatching("hunter22", p.Value, p.Salt) {
		t.Error("correct password did not match")
	}
	if p.IsMatching("hunter23", p.Value, p.Salt) {
		t.Error("wrong password matched")
	}
	if p.IsMatching("hunter22", p.Value, RandomSha1()) {
		t.Error("password matched with wrong salt")
	}

	long := strings.Repeat("correct horse battery staple ", 10)
	if err := p.Create(long); err != nil {
		t.Fatalf("long password: %v", err)
	}
	if !p.IsMatching(long, p.Value, p.Salt) || p.IsMatching(long[:len(long)-1], p.Value, p.Salt) {
		t.Error("long password not compared in full")
	}
}

// Hashes made before passwords were prehashed must keep working.
func TestPasswordLegacy(t *testing.T) {
	salt := RandomSha1()
	hash, err := bcrypt.GenerateFromPassword([]byte("hunter22"+salt), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	var p Password
	if !p.IsMatching("hunter22", string(hash), salt) {
		t.Error("legacy hash did not match")
	}
	if p.IsMatching("hunter23", string(hash), salt) {
		t.Error("legacy hash matched wrong password")
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
	valid := regexp.MustCompile(`^[abcdefghjkmnpqrstuvwxyz1-9]{10}$`)
	for i := 0; i < 100; i++ {
		if s := GenerateShortUuid(); !valid.MatchString(s) {
			t.Fatalf("GenerateShortUuid = %q", s)
		}
	}
	if s := GenerateUuidWithFixedLength(32); len(s) != 32 {
		t.Errorf("length = %d, want 32", len(s))
	}
}

func TestIsClientUuid(t *testing.T) {
	cases := map[string]bool{
		GenerateUuidWithFixedLength(16): true,
		GenerateUuidWithFixedLength(32): true,
		GenerateShortUuid():             false, // too short to be unguessable
		GenerateUuidWithFixedLength(33): false,
		"abcdefghjkmnpqr0":              false, // 0 is not one of the letters
		"ABCDEFGHJKMNPQRS":              false,
		"abcdefghjkmnpq'; --":           false,
	}
	for in, want := range cases {
		if got := IsClientUuid(in); got != want {
			t.Errorf("IsClientUuid(%q) = %v, want %v", in, got, want)
		}
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
		{"display name long", VerifyDisplayName, strings.Repeat("a", 65), false},
		{"display name markup", VerifyDisplayName, "<a href=x>hi</a>", false},
		{"display name accents", VerifyDisplayName, "Åsa Öberg", true},
		{"username ok", VerifyUsername, "kalle.anka", true},
		{"username short", VerifyUsername, "abc", false},
		{"username bad char", VerifyUsername, "kalle anka", false},
		{"username trailing dot", VerifyUsername, "kalle.", false},
		{"password ok", VerifyPassword, "abcdefgh", true},
		{"password short", VerifyPassword, "abcdefg", false},
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
	vars := map[string]string{"DISPLAY_NAME": "<b>Kalle</b>", "LINK": "https://300sf.se/set-password?token=abc123"}
	for _, lang := range []string{"en", "sv", "de"} {
		for _, name := range []string{"welcome", "invite", "set_password", "delete_account"} {
			e, err := GetEmailTemplate(name, lang, "https://300sf.se", vars)
			if err != nil {
				t.Errorf("%s/%s: %v", lang, name, err)
				continue
			}
			if e.Subject == "" || e.Email == "" || e.Name == "" || e.ReplyTo == "" {
				t.Errorf("%s/%s: missing subject, sender or reply-to", lang, name)
			}
			for part, body := range map[string]string{"html": e.Html, "text": e.Txt} {
				if strings.Contains(body, "${") {
					t.Errorf("%s/%s %s: unreplaced variable", lang, name, part)
				}
				if (name == "set_password" || name == "delete_account") && !strings.Contains(body, "token=abc123") {
					t.Errorf("%s/%s %s: link missing", lang, name, part)
				}
			}
			if strings.Contains(e.Html, "<b>Kalle</b>") {
				t.Errorf("%s/%s: display name not escaped", lang, name)
			}
			if !strings.Contains(e.Html, `lang="`+lang+`"`) {
				t.Errorf("%s/%s: rendered in the wrong language", lang, name)
			}
		}
	}
	if e, err := GetEmailTemplate("welcome", "../../etc", "https://300sf.se", nil); err != nil || !strings.Contains(e.Html, `lang="en"`) {
		t.Errorf("unknown language should fall back to English: %v", err)
	}
}

func TestVerifyPortrait(t *testing.T) {
	for in, want := range map[string]bool{"3.webp": true, "10.png": true, "": false, "../../evil.svg": false, "https://evil/x.png": false, "3.svg": false} {
		if got := VerifyPortrait(in); got != want {
			t.Errorf("VerifyPortrait(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestNewToken(t *testing.T) {
	token, hash := NewToken()
	if len(token) != 64 || hash == token || HashToken(token) != hash {
		t.Errorf("token %q hash %q", token, hash)
	}
	if other, _ := NewToken(); other == token {
		t.Error("tokens repeat")
	}
}
