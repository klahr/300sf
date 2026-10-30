// SPDX-License-Identifier: AGPL-3.0-or-later
package main

import (
	"net/http"
	"testing"
)

func TestDBSignInAndSessions(t *testing.T) {
	s := testDB(t)
	u := s.signUp(t)

	s.must(t, u, http.StatusOK, "GET", "/v1/users/whoami", nil)

	wrong := *u
	wrong.Password = "not-the-password"
	if r := s.signIn(t, &wrong, "user"); r.Status != http.StatusBadRequest {
		t.Errorf("wrong password = %d: %s", r.Status, r.Raw)
	}

	// Scopes the user doesn't have are left out, not granted.
	r := s.signIn(t, u, "user", "admin")
	if r.Status != http.StatusOK || r.Body["scope"] != "user" {
		t.Errorf("asking for admin = %d, scope %v", r.Status, r.Body["scope"])
	}
	other := *u
	other.Token = r.Body["access_token"].(string)
	s.must(t, &other, http.StatusUnauthorized, "GET", "/v1/users/list", nil)
	// Asking only for scopes it can't have fails at once, not on the next call.
	if r := s.signIn(t, u, "admin"); r.Status != http.StatusBadRequest || get(r.Body, "code") != float64(3005) {
		t.Errorf("asking only for admin = %d: %s", r.Status, r.Raw)
	}

	// Refresh tokens are single use.
	refresh := func(token string) response {
		return s.do(t, u.address, "", "POST", "/v1/auth", map[string]any{"grant_type": "refresh_token", "refresh_token": token})
	}
	first := refresh(u.Refresh)
	if first.Status != http.StatusOK {
		t.Fatalf("refresh = %d: %s", first.Status, first.Raw)
	}
	if again := refresh(u.Refresh); again.Status != http.StatusUnauthorized {
		t.Errorf("reused refresh token = %d", again.Status)
	}
	newRefresh := first.Body["refresh_token"].(string)

	// A refresh token is no access token.
	refreshAsAccess := *u
	refreshAsAccess.Token = newRefresh
	s.must(t, &refreshAsAccess, http.StatusUnauthorized, "GET", "/v1/users/whoami", nil)

	// Signing out ends the session.
	if out := s.do(t, u.address, "", "POST", "/v1/auth/logout", map[string]any{"refresh_token": newRefresh}); out.Status != http.StatusNoContent {
		t.Errorf("logout = %d", out.Status)
	}
	if after := refresh(newRefresh); after.Status != http.StatusUnauthorized {
		t.Errorf("refresh after logout = %d", after.Status)
	}
}

func TestDBPasswordGuessingIsLimited(t *testing.T) {
	s := testDB(t)
	u := s.signUp(t)
	wrong := *u
	wrong.Password = "guess"
	for i := 0; i < 10; i++ {
		s.signIn(t, &wrong, "user")
	}
	if r := s.signIn(t, u, "user"); r.Status != http.StatusTooManyRequests {
		t.Errorf("right password after 10 wrong from the same network = %d", r.Status)
	}
	// From somewhere else the owner still gets in: a few wrong tries by
	// someone else must not lock the account.
	elsewhere := *u
	elsewhere.address = "10.250.0.1"
	if r := s.signIn(t, &elsewhere, "user"); r.Status != http.StatusOK {
		t.Errorf("right password from another network = %d: %s", r.Status, r.Raw)
	}
}

func TestDBRevokedApiKey(t *testing.T) {
	s := testDB(t)
	u := s.signUp(t)

	key := s.must(t, u, http.StatusOK, "POST", "/v1/api_keys", map[string]any{"name": "script"})
	secret := key.Body["api_key"].(string)
	list := s.must(t, u, http.StatusOK, "GET", "/v1/api_keys/list", nil)
	keys := items(list.Body["api_keys"])
	if len(keys) != 1 {
		t.Fatalf("keys = %v", list.Raw)
	}
	clientId := get(keys[0], "client_id").(string)
	prefix := get(keys[0], "prefix").(string)

	// Sign in with the user's own key.
	r := s.do(t, u.address, "", "POST", "/v1/auth", map[string]any{
		"grant_type": "password", "client_id": clientId, "client_secret": secret,
		"email": u.Email, "password": u.Password, "scopes": []string{"user"},
	})
	if r.Status != http.StatusOK {
		t.Fatalf("sign in with own key = %d: %s", r.Status, r.Raw)
	}
	viaKey := *u
	viaKey.Token = r.Body["access_token"].(string)
	s.must(t, &viaKey, http.StatusOK, "GET", "/v1/lists/list", nil)

	s.must(t, u, http.StatusNoContent, "POST", "/v1/api_keys/"+prefix+"/revoke", nil)

	// Tokens from it stop working, also for making new keys.
	s.must(t, &viaKey, http.StatusUnauthorized, "GET", "/v1/lists/list", nil)
	s.must(t, &viaKey, http.StatusUnauthorized, "POST", "/v1/api_keys", map[string]any{"name": "again"})
}
