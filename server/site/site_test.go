// SPDX-License-Identifier: AGPL-3.0-or-later
package site

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/server/config"
	"github.com/klahr/300sf/server/controllers"
	"github.com/klahr/300sf/server/models"
	"gopkg.in/guregu/null.v4/zero"
)

var cfg = config.Site{
	Host: "300sf.se",
	RedirectHosts: map[string]string{
		"www.300sf.se":      "/",
		"app.300sf.se":      "/home",
		"api-docs.300sf.se": "/docs",
	},
	ClientId:     "web-client",
	ClientSecret: "pfx.secret",
}

func newTestSite(t *testing.T, c config.Site) (*Site, *gin.Engine) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	s, err := New(c, nil)
	if err != nil {
		t.Fatal(err)
	}
	s.stats = func() (*controllers.SiteStats, error) {
		return &controllers.SiteStats{Birders: 386, Sightings: 20401, Species: 449}, nil
	}
	r := gin.New()
	s.Register(r)
	return s, r
}

func get(r *gin.Engine, host, path string, headers ...string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Host = host
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func bootData(t *testing.T, body string) map[string]any {
	t.Helper()
	m := regexp.MustCompile(`const boot = (.*);`).FindStringSubmatch(body)
	if m == nil {
		t.Fatalf("no boot data in page")
	}
	var boot map[string]any
	if err := json.Unmarshal([]byte(m[1]), &boot); err != nil {
		t.Fatalf("boot data is not JSON: %v (%s)", err, m[1])
	}
	return boot
}

func TestLandingPage(t *testing.T) {
	_, r := newTestSite(t, cfg)

	w := get(r, "300sf.se", "/")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	for _, want := range []string{`<html lang="en">`, "<title>300 Swedish Birds</title>", "Parus major", "/css/landing.css", "Every bird you see"} {
		if !strings.Contains(body, want) {
			t.Errorf("landing page is missing %q", want)
		}
	}

	if !strings.Contains(body, `location.replace("/home")`) {
		t.Error("/ should send signed-in visitors to the app")
	}

	sv := get(r, "300sf.se:443", "/?lang=sv").Body.String()
	if !strings.Contains(sv, `<html lang="sv">`) || !strings.Contains(sv, "Talgoxe") {
		t.Error("?lang=sv did not give the Swedish landing page")
	}
	if !strings.Contains(body, "20,401") || !strings.Contains(sv, "20\u00a0401") {
		t.Error("community numbers are missing or not grouped per language")
	}

	browser := get(r, "300sf.se", "/", "Accept-Language", "sv-SE,sv;q=0.9,en;q=0.8").Body.String()
	if !strings.Contains(browser, `<html lang="sv">`) {
		t.Error("Accept-Language sv was ignored")
	}

	landing := get(r, "localhost", "/landing").Body.String()
	if !strings.Contains(landing, "Parus major") || strings.Contains(landing, "location.replace") {
		t.Error("/landing should always show the landing page")
	}
	if !strings.Contains(landing, `href="/sign-up"`) || strings.Contains(landing, "300sf.se") {
		t.Error("landing links should stay on the same host")
	}
}

func TestAppShell(t *testing.T) {
	_, r := newTestSite(t, cfg)

	w := get(r, "300sf.se", "/home")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `src="/js/app.js"`) {
		t.Fatalf("app shell not served: %d", w.Code)
	}
	boot := bootData(t, w.Body.String())
	if boot["language"] != "sv" || boot["redirect"] != "home" {
		t.Errorf("boot = %v", boot)
	}

	boot = bootData(t, get(r, "300sf.se", "/checklists?lang=en&x=1").Body.String())
	if boot["language"] != "en" || boot["redirect"] != "checklists" || boot["queryParameters"].(map[string]any)["x"] != "1" {
		t.Errorf("boot = %v", boot)
	}

	boot = bootData(t, get(r, "300sf.se", "/list/abc123").Body.String())
	if boot["redirect"] != "list" || boot["queryParameters"].(map[string]any)["uuid"] != "abc123" {
		t.Errorf("boot = %v", boot)
	}
}

func TestAppShellQueryCannotBreakOutOfScript(t *testing.T) {
	_, r := newTestSite(t, cfg)
	payload := "x'-alert(1)-'</script><script>alert(2)</script>"
	body := get(r, "300sf.se", "/home?a="+strings.NewReplacer("'", "%27", "<", "%3C", ">", "%3E", "/", "%2F", "(", "%28", ")", "%29").Replace(payload)).Body.String()
	if strings.Contains(body, "</script><script>alert") {
		t.Fatal("query parameter broke out of the boot script")
	}
	if got := bootData(t, body)["queryParameters"].(map[string]any)["a"]; got != payload {
		t.Errorf("payload not preserved: %q", got)
	}
}

func TestDocs(t *testing.T) {
	_, r := newTestSite(t, cfg)
	for _, w := range []*httptest.ResponseRecorder{get(r, "300sf.se", "/docs"), get(r, "localhost", "/docs")} {
		if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "SwaggerUIBundle") {
			t.Errorf("docs not served: %d", w.Code)
		}
	}
	if w := get(r, "localhost", "/swagger.yaml"); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "/v1/auth") {
		t.Errorf("swagger.yaml not served: %d", w.Code)
	}
}

func TestStaticFilesAndPages(t *testing.T) {
	_, r := newTestSite(t, cfg)
	for _, path := range []string{"/css/base.css", "/css/landing.css", "/css/tokens.css", "/img/300sf.svg", "/robots.txt", "/sitemap.xml", "/manifest.json", "/sw.js", "/favicon.png", "/terms", "/privacy-policy"} {
		if w := get(r, "300sf.se", path); w.Code != http.StatusOK {
			t.Errorf("GET %s = %d", path, w.Code)
		}
	}
	if w := get(r, "300sf.se", "/nope"); w.Code != http.StatusNotFound {
		t.Errorf("unknown path = %d", w.Code)
	}
}

func TestComparePage(t *testing.T) {
	s, r := newTestSite(t, cfg)

	if w := get(r, "300sf.se", "/compare"); w.Code != http.StatusFound {
		t.Errorf("without uuids: %d", w.Code)
	}

	var gotUuids []string
	s.compare = func(uuids []string) (*controllers.CompareResult, error) {
		gotUuids = uuids
		return &controllers.CompareResult{
			ReferenceList: models.List{Name: models.Translation{Sv: zero.StringFrom("Lista"), En: zero.StringFrom("List")}},
			Species: []models.Species{{
				Model: models.Model{Uuid: "s1"},
				Name:  models.Translation{Sv: zero.StringFrom("Talgoxe"), La: zero.StringFrom("Parus major")},
			}},
			Result: []controllers.CompareEntry{{
				User:    controllers.CompareUser{DisplayName: "<script>alert(1)</script>"},
				Reports: []controllers.CompareReport{{Species: controllers.CompareSpecies{Uuid: "s1"}}},
			}},
		}, nil
	}
	w := get(r, "300sf.se", "/compare?uuids=a,b")
	body := w.Body.String()
	if w.Code != http.StatusOK || strings.Join(gotUuids, ",") != "a,b" {
		t.Fatalf("status %d, uuids %v", w.Code, gotUuids)
	}
	if !strings.Contains(body, "Talgoxe") || !strings.Contains(body, "check_circle") {
		t.Error("species or check mark missing")
	}
	if strings.Contains(body, "<script>alert") || !strings.Contains(body, "&lt;script&gt;") {
		t.Error("display name not escaped")
	}

	s.compare = func([]string) (*controllers.CompareResult, error) { return nil, errors.New("not found") }
	if w := get(r, "300sf.se", "/compare?uuids=a"); w.Code != http.StatusFound {
		t.Errorf("on error: %d", w.Code)
	}
}

func TestAppSessionUsesCookieAndBlocksCrossSiteWrites(t *testing.T) {
	s, _ := newTestSite(t, cfg)
	r := gin.New()
	api := r.Group("/api", s.AppSession)
	var seen string
	handler := func(c *gin.Context) {
		seen = c.GetHeader("Authorization")
		c.Status(http.StatusOK)
	}
	api.GET("/v1/x", handler)
	api.POST("/v1/x", handler)

	do := func(method string, headers map[string]string) int {
		req := httptest.NewRequest(method, "/api/v1/x", nil)
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		req.AddCookie(&http.Cookie{Name: accessCookie, Value: "tok"})
		w := httptest.NewRecorder()
		seen = ""
		r.ServeHTTP(w, req)
		return w.Code
	}

	if code := do(http.MethodGet, nil); code != http.StatusOK || seen != "Bearer tok" {
		t.Errorf("GET with cookie = %d, Authorization %q", code, seen)
	}
	if code := do(http.MethodPost, nil); code != http.StatusForbidden {
		t.Errorf("POST without X-Requested-With = %d, want 403", code)
	}
	if code := do(http.MethodPost, map[string]string{"X-Requested-With": "XMLHttpRequest"}); code != http.StatusOK || seen != "Bearer tok" {
		t.Errorf("POST from the app = %d", code)
	}
	if code := do(http.MethodGet, map[string]string{"Authorization": "Bearer header"}); code != http.StatusOK || seen != "Bearer header" {
		t.Errorf("explicit Authorization header should win, got %q", seen)
	}
}

func TestAppSignOutClearsCookies(t *testing.T) {
	s, _ := newTestSite(t, cfg)
	r := gin.New()
	r.POST("/api/v1/auth/logout", s.AppSignOut)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil))
	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d", w.Code)
	}
	cleared := map[string]*http.Cookie{}
	for _, c := range w.Result().Cookies() {
		cleared[c.Name] = c
	}
	for name, path := range map[string]string{accessCookie: accessPath, refreshCookie: refreshPath, sessionCookie: "/"} {
		c := cleared[name]
		if c == nil || c.MaxAge >= 0 || c.Path != path || !c.Secure || c.SameSite != http.SameSiteStrictMode {
			t.Errorf("%s not cleared properly: %+v", name, c)
		}
	}
	if !cleared[accessCookie].HttpOnly || !cleared[refreshCookie].HttpOnly || cleared[sessionCookie].HttpOnly {
		t.Error("token cookies must be HttpOnly, the session marker must not")
	}
}

func TestAppSignInRejectsBadBody(t *testing.T) {
	s, _ := newTestSite(t, cfg)
	r := gin.New()
	r.POST("/api/v1/auth", s.AppSignIn)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/auth", strings.NewReader("{")))
	if w.Code != http.StatusBadRequest || len(w.Result().Cookies()) != 0 {
		t.Errorf("bad body = %d with %d cookies", w.Code, len(w.Result().Cookies()))
	}
}

func TestServesFromDirectory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s, err := New(cfg, os.DirFS("."))
	if err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	s.Register(r)
	if w := get(r, "300sf.se", "/"); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Parus major") {
		t.Errorf("landing from directory = %d", w.Code)
	}
}

func TestOldHostsRedirect(t *testing.T) {
	s, _ := newTestSite(t, cfg)
	r := gin.New()
	r.Use(s.Redirect)
	s.Register(r)
	r.GET("/api/health", func(c *gin.Context) { c.String(http.StatusOK, "api") })

	cases := []struct{ host, path, want string }{
		{"app.300sf.se", "/", "https://300sf.se/home"},
		{"app.300sf.se", "/list/abc?x=1", "https://300sf.se/list/abc?x=1"},
		{"app.300sf.se", "/set-password?token=t1", "https://300sf.se/set-password?token=t1"},
		{"api-docs.300sf.se", "/", "https://300sf.se/docs"},
		{"WWW.300sf.se:443", "/?lang=sv", "https://300sf.se/?lang=sv"},
	}
	for _, c := range cases {
		w := get(r, c.host, c.path)
		if w.Code != http.StatusMovedPermanently || w.Header().Get("Location") != c.want {
			t.Errorf("%s%s = %d %q, want 301 %q", c.host, c.path, w.Code, w.Header().Get("Location"), c.want)
		}
	}

	// API calls and non-GET requests on old hosts are answered, not redirected.
	if w := get(r, "app.300sf.se", "/api/health"); w.Code != http.StatusOK {
		t.Errorf("API on old host = %d", w.Code)
	}
	req := httptest.NewRequest(http.MethodPost, "/analyze", nil)
	req.Host = "app.300sf.se"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusMovedPermanently {
		t.Error("POST on old host was redirected")
	}
	// The public host itself is never redirected.
	if w := get(r, "300sf.se", "/"); w.Code != http.StatusOK {
		t.Errorf("300sf.se/ = %d", w.Code)
	}
}
