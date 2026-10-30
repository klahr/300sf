// SPDX-License-Identifier: AGPL-3.0-or-later
package controllers

import (
	"net/url"
	"testing"
)

func TestValidRedirectUri(t *testing.T) {
	cases := map[string]bool{
		"https://claude.ai/api/mcp/auth_callback": true,
		"http://localhost:33418/callback":         true,
		"http://127.0.0.1/callback":               true,
		"http://[::1]:8080/cb":                    true,
		"http://example.com/callback":             false,
		"https://claude.ai/cb#fragment":           false,
		"https://user:pass@claude.ai/cb":          false,
		"javascript:alert(1)":                     false,
		"myapp://callback":                        false,
		"/relative":                               false,
		"":                                        false,
	}
	for uri, want := range cases {
		if got := validRedirectUri(uri); got != want {
			t.Errorf("validRedirectUri(%q) = %v, want %v", uri, got, want)
		}
	}
}

func TestMatchRedirectUri(t *testing.T) {
	registered := []string{"https://claude.ai/api/mcp/auth_callback", "http://localhost:5000/callback"}
	cases := []struct {
		given, want string
		ok          bool
	}{
		{"https://claude.ai/api/mcp/auth_callback", "https://claude.ai/api/mcp/auth_callback", true},
		{"https://claude.ai/api/mcp/auth_callback?x=1", "", false},
		{"https://evil.example/api/mcp/auth_callback", "", false},
		// A loopback port may differ, nothing else.
		{"http://localhost:6123/callback", "http://localhost:6123/callback", true},
		{"http://localhost:6123/other", "", false},
		{"http://127.0.0.1:5000/callback", "", false},
		{"", "", false},
	}
	for _, c := range cases {
		got, ok := matchRedirectUri(registered, c.given)
		if got != c.want || ok != c.ok {
			t.Errorf("matchRedirectUri(%q) = %q, %v; want %q, %v", c.given, got, ok, c.want, c.ok)
		}
	}
	if got, ok := matchRedirectUri(registered[:1], ""); !ok || got != registered[0] {
		t.Errorf("the only registered uri is not used when none is given: %q %v", got, ok)
	}
}

func TestPkceMatches(t *testing.T) {
	// From RFC 7636, appendix B.
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
	if !pkceMatches(verifier, challenge) {
		t.Error("RFC 7636 example does not match")
	}
	if pkceMatches(verifier+"x", challenge) || pkceMatches("short", challenge) || pkceMatches("", "") {
		t.Error("a wrong verifier matched")
	}
}

func TestRedirectWithKeepsQuery(t *testing.T) {
	got, _ := url.Parse(redirectWith("https://app.example/cb?keep=1", map[string]string{"code": "abc", "state": "x y", "empty": ""}))
	q := got.Query()
	if q.Get("keep") != "1" || q.Get("code") != "abc" || q.Get("state") != "x y" || q.Has("empty") {
		t.Errorf("redirectWith = %s", got)
	}
}
