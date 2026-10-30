// SPDX-License-Identifier: AGPL-3.0-or-later
package main

// Connecting an app the way Claude does: it registers, the user approves it
// on the website, it gets tokens with PKCE and uses the tools on /mcp.

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/klahr/300sf/server/utils"
)

// raw is a request with the body and headers as given.
func (s *dbServer) raw(t *testing.T, address, method, path, body string, headers ...string) response {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.RemoteAddr = address + ":40000"
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	r := response{Status: w.Code, Raw: w.Body.String()}
	_ = json.Unmarshal(w.Body.Bytes(), &r.Body)
	if r.Body == nil {
		r.Body = map[string]any{}
	}
	r.Body["_www_authenticate"] = w.Header().Get("WWW-Authenticate")
	return r
}

// oauthApp is a registered app partway through connecting.
type oauthApp struct {
	clientId, redirect, verifier string
}

func (s *dbServer) registerApp(t *testing.T, u *testUser) oauthApp {
	t.Helper()
	app := oauthApp{redirect: "https://claude.example/api/mcp/auth_callback", verifier: utils.RandomSha256() + utils.RandomSha256()[:20]}
	r := s.raw(t, u.address, "POST", "/oauth/register", `{"client_name":"Claude","redirect_uris":["`+app.redirect+`"],"token_endpoint_auth_method":"none"}`, "Content-Type", "application/json")
	if r.Status != http.StatusCreated || r.Body["token_endpoint_auth_method"] != "none" {
		t.Fatalf("register = %d: %s", r.Status, r.Raw)
	}
	app.clientId = r.Body["client_id"].(string)
	return app
}

func (a oauthApp) challenge() string {
	sum := sha256.Sum256([]byte(a.verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// approve answers the authorize page as the user, with their website
// session, and returns the address they are sent back to.
func (s *dbServer) approve(t *testing.T, u *testUser, app oauthApp, allow, allowChanges bool) *url.URL {
	t.Helper()
	body, _ := json.Marshal(map[string]any{
		"client_id": app.clientId, "redirect_uri": app.redirect, "response_type": "code", "state": "st4te",
		"code_challenge": app.challenge(), "code_challenge_method": "S256", "scope": "read user",
		"resource": "http://localhost/mcp", "allow": allow, "allow_changes": allowChanges,
	})
	r := s.raw(t, u.address, "POST", "/api/v1/oauth/approve", string(body),
		"Content-Type", "application/json", "Authorization", "Bearer "+u.Token, "X-Requested-With", "XMLHttpRequest")
	if r.Status != http.StatusOK {
		t.Fatalf("approve = %d: %s", r.Status, r.Raw)
	}
	back, err := url.Parse(r.Body["redirect"].(string))
	if err != nil || !strings.HasPrefix(back.String(), app.redirect+"?") || back.Query().Get("state") != "st4te" || back.Query().Get("iss") != "http://localhost" {
		t.Fatalf("sent back to %v", r.Body["redirect"])
	}
	return back
}

func (s *dbServer) token(t *testing.T, address string, form url.Values) response {
	t.Helper()
	return s.raw(t, address, "POST", "/oauth/token", form.Encode(), "Content-Type", "application/x-www-form-urlencoded")
}

// mcpCall is a JSON-RPC call on /mcp; it returns the result.
func (s *dbServer) mcpCall(t *testing.T, address, token, method string, params any) (response, map[string]any) {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 1, "method": method, "params": params})
	r := s.raw(t, address, "POST", "/mcp", string(body),
		"Content-Type", "application/json", "Accept", "application/json, text/event-stream", "Authorization", "Bearer "+token)
	result, _ := r.Body["result"].(map[string]any)
	return r, result
}

func toolNames(result map[string]any) []string {
	var names []string
	for _, tool := range items(result["tools"]) {
		names = append(names, tool.(map[string]any)["name"].(string))
	}
	return names
}

// callTool calls a tool and decodes the JSON it answers with.
func (s *dbServer) callTool(t *testing.T, address, token, name string, args map[string]any) any {
	t.Helper()
	r, result := s.mcpCall(t, address, token, "tools/call", map[string]any{"name": name, "arguments": args})
	if r.Status != http.StatusOK || result == nil || result["isError"] == true {
		t.Fatalf("%s = %d: %s", name, r.Status, r.Raw)
	}
	var out any
	text := items(result["content"])[0].(map[string]any)["text"].(string)
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		t.Fatalf("%s answered %q", name, text)
	}
	return out
}

func TestDBConnectAppWithOAuth(t *testing.T) {
	s := testDB(t)
	u := s.signUp(t)
	app := s.registerApp(t, u)

	// Saying no sends the user back with an error and no code.
	if back := s.approve(t, u, app, false, false); back.Query().Get("error") != "access_denied" || back.Query().Has("code") {
		t.Errorf("denied = %s", back)
	}

	back := s.approve(t, u, app, true, true)
	code := back.Query().Get("code")
	exchange := url.Values{"grant_type": {"authorization_code"}, "code": {code}, "client_id": {app.clientId}, "redirect_uri": {app.redirect}, "code_verifier": {app.verifier}}

	// The code is only good with the verifier, for the app it was given to.
	wrong := url.Values{}
	for k, v := range exchange {
		wrong[k] = v
	}
	wrong.Set("code_verifier", strings.Repeat("a", 43))
	if r := s.token(t, u.address, wrong); r.Status != http.StatusBadRequest || r.Body["error"] != "invalid_grant" {
		t.Errorf("wrong verifier = %d: %s", r.Status, r.Raw)
	}
	// ...and only once: the failed try used it up.
	if r := s.token(t, u.address, exchange); r.Status != http.StatusBadRequest {
		t.Errorf("a used code worked: %s", r.Raw)
	}

	exchange.Set("code", s.approve(t, u, app, true, true).Query().Get("code"))
	tokens := s.token(t, u.address, exchange)
	if tokens.Status != http.StatusOK || tokens.Body["scope"] != "user" || tokens.Body["token_type"] != "Bearer" {
		t.Fatalf("token = %d: %s", tokens.Status, tokens.Raw)
	}
	access := tokens.Body["access_token"].(string)

	// The connection is one of the user's keys, marked as such.
	keys := items(s.must(t, u, http.StatusOK, "GET", "/v1/api_keys/list", nil).Body["api_keys"])
	var prefix string
	for _, k := range keys {
		if k := k.(map[string]any); k["connected"] == true {
			prefix = k["prefix"].(string)
			if k["name"] != "Claude" || k["scopes"] != "user" {
				t.Errorf("connection = %v", k)
			}
		}
	}
	if prefix == "" {
		t.Fatalf("no connection among the keys: %v", keys)
	}
	// Approving again keeps the one connection.
	s.approve(t, u, app, true, true)
	if n := len(items(s.must(t, u, http.StatusOK, "GET", "/v1/api_keys/list", nil).Body["api_keys"])); n != len(keys) {
		t.Errorf("approving again made another connection: %d keys, had %d", n, len(keys))
	}

	// The tools, as the app sees them.
	r, result := s.mcpCall(t, u.address, access, "initialize", map[string]any{
		"protocolVersion": "2025-06-18", "capabilities": map[string]any{}, "clientInfo": map[string]any{"name": "test", "version": "1"},
	})
	if r.Status != http.StatusOK || get(result, "serverInfo", "name") != "300sf" {
		t.Fatalf("initialize = %d: %s", r.Status, r.Raw)
	}
	_, result = s.mcpCall(t, u.address, access, "tools/list", map[string]any{})
	if names := toolNames(result); !strings.Contains(strings.Join(names, ","), "add_sighting") || !strings.Contains(strings.Join(names, ","), "get_checklist") {
		t.Fatalf("tools = %v", names)
	}

	me := s.callTool(t, u.address, access, "whoami", nil)
	if get(me, "username") != u.Username || get(me, "can_change") != true {
		t.Errorf("whoami = %v", me)
	}
	found := items(s.callTool(t, u.address, access, "search_species", map[string]any{"query": "grus grus"}))
	if len(found) == 0 || !strings.Contains(get(found[0], "name").(string), "Grus grus") {
		t.Fatalf("search_species = %v", found)
	}
	crane := get(found[0], "uuid").(string)

	list := s.callTool(t, u.address, access, "create_checklist", map[string]any{"name": "Året 2026", "from": "2026-01-01", "until": "2026-12-31"})
	listUuid := get(list, "uuid").(string)
	added := s.callTool(t, u.address, access, "add_sighting", map[string]any{
		"species_uuid": crane, "from": "2026-04-02", "quantity": 12, "site": "Hornborgasjön", "checklist_uuids": []string{listUuid},
	})
	if get(added, "species_uuid") != crane || get(added, "quantity") != float64(12) {
		t.Errorf("add_sighting = %v", added)
	}
	// Changing one field keeps the others.
	changed := s.callTool(t, u.address, access, "update_sighting", map[string]any{"uuid": get(added, "uuid"), "quantity": 14})
	if get(changed, "quantity") != float64(14) || get(changed, "site") != "Hornborgasjön" || len(items(get(changed, "checklists"))) != 1 {
		t.Errorf("update_sighting = %v", changed)
	}
	progress := s.callTool(t, u.address, access, "get_checklist", map[string]any{"uuid": listUuid})
	if get(progress, "species_seen") != float64(1) || get(progress, "seen", 0, "first_seen") != "2026-04-02" {
		t.Errorf("get_checklist = %v", progress)
	}
	if found := s.callTool(t, u.address, access, "list_sightings", map[string]any{"from": "2026-04-01", "until": "2026-04-30", "species": "trana"}); get(found, "matching") != float64(1) {
		t.Errorf("list_sightings = %v", found)
	}
	// An API error comes back as a tool error, for the assistant to read.
	if _, result := s.mcpCall(t, u.address, access, "tools/call", map[string]any{"name": "get_sighting", "arguments": map[string]any{"uuid": "nope"}}); result["isError"] != true {
		t.Errorf("get_sighting of nothing = %v", result)
	}

	// Refreshing works for the app the token was given to, and no other.
	other := s.registerApp(t, u)
	refresh := url.Values{"grant_type": {"refresh_token"}, "refresh_token": {tokens.Body["refresh_token"].(string)}, "client_id": {other.clientId}}
	if r := s.token(t, u.address, refresh); r.Status != http.StatusBadRequest {
		t.Errorf("another app refreshed: %s", r.Raw)
	}
	refresh.Set("client_id", app.clientId)
	refreshed := s.token(t, u.address, refresh)
	if refreshed.Status != http.StatusOK || refreshed.Body["scope"] != "user" {
		t.Fatalf("refresh = %d: %s", refreshed.Status, refreshed.Raw)
	}
	// An ordinary API key's refresh token isn't taken here.
	if r := s.token(t, u.address, url.Values{"grant_type": {"refresh_token"}, "refresh_token": {u.Refresh}}); r.Status != http.StatusBadRequest {
		t.Errorf("an API key's refresh token worked: %s", r.Raw)
	}

	// Disconnecting in Settings ends it at once, with a challenge to reconnect.
	s.must(t, u, http.StatusNoContent, "POST", "/v1/api_keys/"+prefix+"/revoke", nil)
	r, _ = s.mcpCall(t, u.address, refreshed.Body["access_token"].(string), "tools/list", map[string]any{})
	if r.Status != http.StatusUnauthorized || !strings.Contains(r.Body["_www_authenticate"].(string), `resource_metadata="http://localhost/.well-known/oauth-protected-resource/mcp"`) {
		t.Errorf("after disconnecting /mcp = %d, %v", r.Status, r.Body["_www_authenticate"])
	}
	refresh.Set("refresh_token", refreshed.Body["refresh_token"].(string))
	if r := s.token(t, u.address, refresh); r.Status != http.StatusBadRequest {
		t.Errorf("refresh after disconnecting = %d: %s", r.Status, r.Raw)
	}
}

// Without "allow changes" the app only reads, and is offered no tools that change anything.
func TestDBConnectAppReadOnly(t *testing.T) {
	s := testDB(t)
	u := s.signUp(t)
	app := s.registerApp(t, u)
	back := s.approve(t, u, app, true, false)
	tokens := s.token(t, u.address, url.Values{"grant_type": {"authorization_code"}, "code": {back.Query().Get("code")}, "client_id": {app.clientId}, "redirect_uri": {app.redirect}, "code_verifier": {app.verifier}})
	if tokens.Status != http.StatusOK || tokens.Body["scope"] != "read" {
		t.Fatalf("token = %d: %s", tokens.Status, tokens.Raw)
	}
	access := tokens.Body["access_token"].(string)
	_, result := s.mcpCall(t, u.address, access, "tools/list", map[string]any{})
	names := strings.Join(toolNames(result), ",")
	if strings.Contains(names, "add_sighting") || strings.Contains(names, "delete_sighting") || !strings.Contains(names, "list_sightings") {
		t.Errorf("read-only tools = %s", names)
	}
	if me := s.callTool(t, u.address, access, "whoami", nil); get(me, "can_change") != false {
		t.Errorf("whoami = %v", me)
	}
	// The API refuses changes too, whatever the app tries.
	s.must(t, &testUser{address: u.address, Token: access}, http.StatusUnauthorized, "POST", "/v1/lists", map[string]any{"name": map[string]string{"en": "x"}})

	// An ordinary token works on /mcp as well; a website session can't approve
	// for another app's redirect.
	_, result = s.mcpCall(t, u.address, u.Token, "tools/list", map[string]any{})
	if !strings.Contains(strings.Join(toolNames(result), ","), "add_sighting") {
		t.Errorf("API key token tools = %v", toolNames(result))
	}
	body := `{"client_id":"` + app.clientId + `","redirect_uri":"https://evil.example/cb","response_type":"code","code_challenge":"x","code_challenge_method":"S256","allow":true}`
	if r := s.raw(t, u.address, "POST", "/api/v1/oauth/approve", body, "Content-Type", "application/json", "Authorization", "Bearer "+u.Token, "X-Requested-With", "x"); r.Status != http.StatusBadRequest {
		t.Errorf("approve to an unregistered redirect = %d: %s", r.Status, r.Raw)
	}
	// Approving needs the website's API, not the public one.
	if r := s.raw(t, u.address, "POST", "/v1/oauth/approve", body, "Content-Type", "application/json", "Authorization", "Bearer "+u.Token); r.Status != http.StatusNotFound {
		t.Errorf("/v1/oauth/approve = %d", r.Status)
	}
}
