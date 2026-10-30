// SPDX-License-Identifier: AGPL-3.0-or-later
package main

// Tests against a real database and Redis: the whole API, through the same
// router the server runs, with the schema from db.sql. They are skipped
// unless both are given:
//
//	TEST_DATA_SOURCE="root:test@tcp(127.0.0.1:33306)/300sf?charset=utf8mb4&parseTime=True&loc=Local"
//	TEST_REDIS="127.0.0.1:36379"
//
// `make test-db` starts both in Docker and runs them; CI does the same. The
// tests write to the database and empty Redis, so never point them at one
// that matters: only addresses on this machine are accepted.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/server/auth"
	"github.com/klahr/300sf/server/config"
	"github.com/klahr/300sf/server/models"
	"github.com/klahr/300sf/server/seed"
	"github.com/klahr/300sf/server/site"
	"github.com/klahr/300sf/server/utils"
)

type dbServer struct {
	router       *gin.Engine
	clientId     string
	clientSecret string
}

var (
	dbOnce   sync.Once
	dbShared *dbServer
	dbErr    error
	// Each test user calls from an address of their own, as people do, so
	// the per-address limits (sign-up, sign-in) don't run out.
	nextAddress atomic.Int64
)

// testDB is the server running against the test database, set up once for
// all tests; the test is skipped when there is none.
func testDB(t *testing.T) *dbServer {
	t.Helper()
	dataSource, redisHost := os.Getenv("TEST_DATA_SOURCE"), os.Getenv("TEST_REDIS")
	if dataSource == "" || redisHost == "" {
		t.Skip("TEST_DATA_SOURCE and TEST_REDIS not set; see db_test.go")
	}
	dbOnce.Do(func() {
		dbShared, dbErr = startDBServer(dataSource, redisHost)
	})
	if dbErr != nil {
		t.Fatal(dbErr)
	}
	return dbShared
}

func local(address string) bool {
	return strings.Contains(address, "127.0.0.1") || strings.Contains(address, "localhost")
}

func startDBServer(dataSource, redisHost string) (*dbServer, error) {
	if !local(dataSource) || !local(redisHost) {
		return nil, fmt.Errorf("the database tests only run against services on this machine")
	}
	gin.SetMode(gin.TestMode)
	config.AppConfig = &config.App{
		HashingSecret:        "test-hashing-secret-0123456789abcdef",
		TokenSecret:          "test-token-secret-0123456789abcdef",
		AccessTokenLifetime:  600,
		RefreshTokenLifetime: 3600,
		DefaultApiSoftLimit:  100000,
		DefaultApiHardLimit:  100000,
		ApiCallCounterTTL:    3600,
		Domain:               "http://localhost",
		Mode:                 gin.TestMode,
	}
	if err := models.SetupDB(dataSource, 1); err != nil {
		return nil, err
	}
	if err := models.SetupCache(redisHost, ""); err != nil {
		return nil, err
	}
	// Rate limits and sessions from an earlier run.
	if err := models.Cache.FlushDB().Err(); err != nil {
		return nil, err
	}
	if _, _, err := seed.Load(models.DB); err != nil {
		return nil, err
	}
	web, err := site.New(config.Site{}, nil)
	if err != nil {
		return nil, err
	}
	s := &dbServer{router: newRouter(web)}

	// An API key to sign in with, owned by an account of its own.
	owner := models.User{
		Model:    models.Model{Timestamp: time.Now(), Uuid: utils.GenerateShortUuid()},
		Email:    "key-owner-" + utils.GenerateUuidWithFixedLength(8) + "@example.com",
		Username: "keyowner" + utils.GenerateUuidWithFixedLength(8), DisplayName: "Key owner",
		Password: "-", PasswordSalt: "-", Scopes: auth.AuthScopeUser, LanguageId: 1,
	}
	if err := models.DB.Create(&owner).Error; err != nil {
		return nil, err
	}
	secret := utils.RandomSha256()
	key := models.ApiKey{
		Model:    models.Model{Timestamp: time.Now(), Uuid: utils.GenerateShortUuid()},
		Name:     "Tests",
		ClientId: utils.GenerateUuidWithFixedLength(32),
		Prefix:   utils.GenerateUuidWithFixedLength(7),
		Hash:     utils.ComputeHmac256(secret, config.AppConfig.HashingSecret),
		UserId:   owner.Id, Scopes: auth.AuthScopeUser + " " + auth.AuthScopeAdmin,
		SoftLimit: 100000, HardLimit: 100000,
	}
	if err := models.DB.Create(&key).Error; err != nil {
		return nil, err
	}
	s.clientId, s.clientSecret = key.ClientId, key.Prefix+"."+secret
	return s, nil
}

// testUser is a signed-up account and its session.
type testUser struct {
	Uuid     string
	Email    string
	Username string
	Password string
	Token    string
	Refresh  string
	address  string
}

// response is an answer from the API, with its JSON decoded.
type response struct {
	Status int
	Body   map[string]any
	Raw    string
}

func (s *dbServer) do(t *testing.T, address, token, method, path string, body any) response {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	req.RemoteAddr = address + ":40000"
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-App-Language", "en")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	r := response{Status: w.Code, Raw: w.Body.String()}
	var decoded any
	if json.Unmarshal(w.Body.Bytes(), &decoded) == nil {
		if m, ok := decoded.(map[string]any); ok {
			r.Body = m
		} else {
			r.Body = map[string]any{"": decoded}
		}
	}
	return r
}

// call is a request as the user.
func (s *dbServer) call(t *testing.T, u *testUser, method, path string, body any) response {
	t.Helper()
	return s.do(t, u.address, u.Token, method, path, body)
}

// must is call, failing the test unless the answer has the status.
func (s *dbServer) must(t *testing.T, u *testUser, status int, method, path string, body any) response {
	t.Helper()
	r := s.call(t, u, method, path, body)
	if r.Status != status {
		t.Fatalf("%s %s = %d, want %d: %s", method, path, r.Status, status, r.Raw)
	}
	return r
}

func (s *dbServer) signIn(t *testing.T, u *testUser, scopes ...string) response {
	t.Helper()
	return s.do(t, u.address, "", "POST", "/v1/auth", map[string]any{
		"grant_type": "password", "client_id": s.clientId, "client_secret": s.clientSecret,
		"email": u.Email, "password": u.Password, "scopes": scopes,
	})
}

// signUp makes a new account, signed in with the user scope.
func (s *dbServer) signUp(t *testing.T) *testUser {
	t.Helper()
	n := nextAddress.Add(1)
	id := utils.GenerateUuidWithFixedLength(10)
	u := &testUser{
		Email:    "user-" + id + "@example.com",
		Username: "user" + id,
		Password: "correct-horse-" + id,
		address:  fmt.Sprintf("10.%d.%d.%d", (n>>16)&255, (n>>8)&255, n&255),
	}
	r := s.do(t, u.address, "", "POST", "/v1/users", map[string]any{
		"email": u.Email, "display_name": "Test " + id, "username": u.Username,
		"password": u.Password, "accepted_terms": true,
	})
	if r.Status != http.StatusOK {
		t.Fatalf("sign up = %d: %s", r.Status, r.Raw)
	}
	u.Uuid = r.Body["uuid"].(string)
	s.session(t, u, auth.AuthScopeUser)
	return u
}

// signUpAdmin makes a new account with the admin scope, signed in with it.
func (s *dbServer) signUpAdmin(t *testing.T) *testUser {
	t.Helper()
	u := s.signUp(t)
	if err := models.DB.Model(&models.User{}).Where("uuid=?", u.Uuid).Update("scopes", "user admin").Error; err != nil {
		t.Fatal(err)
	}
	s.session(t, u, auth.AuthScopeUser, auth.AuthScopeAdmin)
	return u
}

func (s *dbServer) session(t *testing.T, u *testUser, scopes ...string) {
	t.Helper()
	r := s.signIn(t, u, scopes...)
	if r.Status != http.StatusOK {
		t.Fatalf("sign in = %d: %s", r.Status, r.Raw)
	}
	u.Token = r.Body["access_token"].(string)
	u.Refresh = r.Body["refresh_token"].(string)
}

// get walks into decoded JSON: get(body, "data", 0, "uuid").
func get(v any, path ...any) any {
	for _, p := range path {
		switch k := p.(type) {
		case string:
			m, ok := v.(map[string]any)
			if !ok {
				return nil
			}
			v = m[k]
		case int:
			a, ok := v.([]any)
			if !ok || k >= len(a) {
				return nil
			}
			v = a[k]
		}
	}
	return v
}

func items(v any) []any {
	a, _ := v.([]any)
	return a
}

// uuids are the "uuid" of each item.
func uuids(v any) []string {
	var out []string
	for _, e := range items(v) {
		if u, ok := get(e, "uuid").(string); ok {
			out = append(out, u)
		}
	}
	return out
}

func contains(list []string, s string) bool {
	for _, e := range list {
		if e == s {
			return true
		}
	}
	return false
}
