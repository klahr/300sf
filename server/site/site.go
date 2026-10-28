// SPDX-License-Identifier: GPL-2.0-or-later

// Package site serves the websites from the same binary as the API, all on
// one host: the landing page on /, the app on /home and its other pages, and
// the API docs on /docs. Everything they need is embedded, so the server is a
// single file.
package site

import (
	"embed"
	"encoding/json"
	"html/template"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/server/config"
	"github.com/klahr/300sf/server/controllers"
	"github.com/klahr/300sf/server/models"
)

//go:embed static templates pages locales
var embedded embed.FS

// Client-side routes of the app; the page is chosen in the browser.
var appPages = []string{
	"home", "sign-up", "forgot", "settings", "checklists", "contribute", "donation",
	"contact", "tags", "reports", "set-password", "delete-account", "competitions",
	"admin", "experimental",
}

var appNames = map[string]string{
	"sv": "300 Svenska Fåglar",
	"en": "300 Swedish Birds",
	"de": "300 Schwedische Vögel",
}

type Site struct {
	files     fs.FS
	cfg       config.Site
	templates *template.Template
	landing   map[string]map[string]any
	// Overridable in tests.
	compare func(uuids []string) (*controllers.CompareResult, error)
}

// New serves the embedded site. Pass a directory (os.DirFS("site")) as files
// during development to pick up changes without rebuilding; nil uses the
// embedded copy.
func New(cfg config.Site, files fs.FS) (*Site, error) {
	if files == nil {
		files = embedded
	}
	s := &Site{files: files, cfg: cfg, landing: map[string]map[string]any{}, compare: controllers.CompareLists}

	funcs := template.FuncMap{
		"year": func() int { return time.Now().Year() },
		"translate": func(t models.Translation, lang string) string {
			if lang == "sv" && t.Sv.String != "" {
				return t.Sv.String
			}
			if lang == "de" && t.De.String != "" {
				return t.De.String
			}
			return t.En.String
		},
	}
	t, err := template.New("").Funcs(funcs).ParseFS(s.files, "templates/*.html")
	if err != nil {
		return nil, err
	}
	s.templates = t

	for _, lang := range []string{"en", "sv"} {
		raw, err := fs.ReadFile(s.files, "locales/landing-"+lang+".json")
		if err != nil {
			return nil, err
		}
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil {
			return nil, err
		}
		s.landing[lang] = m
	}
	return s, nil
}

// Register adds the website routes. API routes are registered separately and
// all live under /v1, /health and /api, so they never collide with these.
func (s *Site) Register(r *gin.Engine) {
	static, _ := fs.Sub(s.files, "static")
	entries, _ := fs.ReadDir(static, ".")
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			sub, _ := fs.Sub(static, name)
			r.StaticFS("/"+name, http.FS(sub))
		} else {
			r.StaticFileFS("/"+name, name, http.FS(static))
		}
	}

	r.GET("/", s.root)
	r.GET("/landing", s.landingPage)
	r.GET("/docs", s.docs)
	for _, page := range appPages {
		page := page
		r.GET("/"+page, func(c *gin.Context) {
			s.app(c, page, queryMap(c))
		})
	}
	r.GET("/list/:uuid", func(c *gin.Context) {
		s.app(c, "list", map[string]string{"uuid": c.Param("uuid")})
	})
	r.GET("/terms", s.page("pages/terms.html"))
	r.GET("/privacy-policy", s.page("pages/privacy-policy.html"))
	r.GET("/compare", s.comparePage)
	r.POST("/analyze", s.analyze())
}

// The app's session lives in cookies the page's JavaScript can't read:
//
//   - accessCookie: the access token, sent only to /api.
//   - refreshCookie: the refresh token, sent only to the sign-in endpoint.
//   - sessionCookie: not secret; tells the page that it is signed in and when
//     the access token runs out, so it can refresh in time.
//
// All are Secure and SameSite=Strict, so other sites can't send them along.
const (
	accessCookie  = "__Secure-sf_at"
	refreshCookie = "__Secure-sf_rt"
	sessionCookie = "sf_session"

	accessPath  = "/api"
	refreshPath = "/api/v1/auth"
)

func setCookie(c *gin.Context, name, value, path string, maxAge int, httpOnly bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		MaxAge:   maxAge,
		Secure:   true,
		HttpOnly: httpOnly,
		SameSite: http.SameSiteStrictMode,
	})
}

func clearSessionCookies(c *gin.Context) {
	setCookie(c, accessCookie, "", accessPath, -1, true)
	setCookie(c, refreshCookie, "", refreshPath, -1, true)
	setCookie(c, sessionCookie, "", "/", -1, false)
}

// AppSession lets the app call the API with its cookie instead of an
// Authorization header, and refuses cross-site writes: every state-changing
// request must carry X-Requested-With, which another site can only send after
// a CORS preflight, and /api answers no preflights.
func (s *Site) AppSession(c *gin.Context) {
	switch c.Request.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
	default:
		if c.GetHeader("X-Requested-With") == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, controllers.GeneralErrorResponse{Message: "missing X-Requested-With", Code: controllers.ErrorInvalidRequest})
			return
		}
	}
	if c.GetHeader("Authorization") == "" {
		if token, err := c.Cookie(accessCookie); err == nil && token != "" {
			c.Request.Header.Set("Authorization", "Bearer "+token)
		}
	}
	c.Next()
}

// AppSignIn signs the app in with a password or refreshes its session. The
// app's API key is added here, so it never reaches the browser, and the
// tokens go into cookies instead of the response.
func (s *Site) AppSignIn(c *gin.Context) {
	var r controllers.AuthRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, controllers.GeneralErrorResponse{Message: err.Error(), Code: controllers.ErrorParse})
		return
	}
	r.ClientId = s.cfg.ClientId
	r.ClientSecret = s.cfg.ClientSecret
	if r.GrantType == "refresh_token" {
		r.RefreshToken, _ = c.Cookie(refreshCookie)
	}

	tokens, authErr := controllers.Authenticate(r)
	if authErr != nil {
		if r.GrantType == "refresh_token" {
			clearSessionCookies(c)
		}
		c.JSON(authErr.Status, authErr.Body)
		return
	}

	// Without "stay signed in" the session ends when the browser closes.
	sessionAge := 0
	if tokens.StaySignedIn {
		sessionAge = int(config.AppConfig.RefreshTokenLifetime)
	}
	accessExpiresAt := time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second).UnixMilli()
	setCookie(c, accessCookie, tokens.AccessToken, accessPath, int(tokens.ExpiresIn), true)
	setCookie(c, refreshCookie, tokens.RefreshToken, refreshPath, sessionAge, true)
	setCookie(c, sessionCookie, strconv.FormatInt(accessExpiresAt, 10), "/", sessionAge, false)

	c.JSON(http.StatusOK, gin.H{
		"expires_in": tokens.ExpiresIn,
		"scope":      tokens.Scope,
		"lang":       tokens.Lang,
	})
}

// AppSignOut ends the session on the server and removes the cookies.
func (s *Site) AppSignOut(c *gin.Context) {
	if token, err := c.Cookie(refreshCookie); err == nil {
		controllers.RevokeRefreshToken(token)
	}
	clearSessionCookies(c)
	c.Status(http.StatusNoContent)
}

func hostname(c *gin.Context) string {
	host := strings.ToLower(c.Request.Host)
	if i := strings.LastIndex(host, ":"); i > 0 && !strings.Contains(host[i:], "]") {
		host = host[:i]
	}
	return host
}

// IsAPIPath reports whether path belongs to the API rather than the websites.
func IsAPIPath(path string) bool {
	return strings.HasPrefix(path, "/v1/") || strings.HasPrefix(path, "/api/") || path == "/health"
}

// Redirect sends pages on the old hosts (app.300sf.se, api-docs.300sf.se, ...)
// to the same path on the one public host.
func (s *Site) Redirect(c *gin.Context) {
	root, ok := s.cfg.RedirectHosts[hostname(c)]
	if !ok || s.cfg.Host == "" || IsAPIPath(c.Request.URL.Path) || (c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead) {
		c.Next()
		return
	}
	target := *c.Request.URL
	target.Scheme = "https"
	target.Host = s.cfg.Host
	if target.Path == "/" || target.Path == "" {
		target.Path = root
	}
	c.Redirect(http.StatusMovedPermanently, target.String())
	c.Abort()
}

func (s *Site) root(c *gin.Context) {
	s.render(c, http.StatusOK, "landing.html", s.landingData(c, true))
}

// language picks ?lang=, then the browser's preference, then the fallback.
func language(c *gin.Context, supported []string, fallback string) string {
	ok := func(l string) bool {
		for _, s := range supported {
			if s == l {
				return true
			}
		}
		return false
	}
	if l := c.Query("lang"); ok(l) {
		return l
	}
	for _, part := range strings.Split(c.GetHeader("Accept-Language"), ",") {
		l := strings.ToLower(strings.TrimSpace(strings.SplitN(part, ";", 2)[0]))
		if len(l) >= 2 && ok(l[:2]) {
			return l[:2]
		}
	}
	return fallback
}

func queryMap(c *gin.Context) map[string]string {
	m := map[string]string{}
	for k, v := range c.Request.URL.Query() {
		if len(v) > 0 {
			m[k] = v[0]
		}
	}
	return m
}

func (s *Site) render(c *gin.Context, status int, name string, data any) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(status)
	if err := s.templates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Error(err)
	}
}

func (s *Site) landingData(c *gin.Context, signedInGoesToApp bool) gin.H {
	lang := language(c, []string{"en", "sv"}, "en")
	return gin.H{"Lang": lang, "T": s.landing[lang], "SignedInGoesToApp": signedInGoesToApp}
}

// landingPage always shows the landing page, also to signed-in visitors.
func (s *Site) landingPage(c *gin.Context) {
	s.render(c, http.StatusOK, "landing.html", s.landingData(c, false))
}

func (s *Site) app(c *gin.Context, redirect string, query map[string]string) {
	lang := language(c, []string{"sv", "en", "de"}, "sv")
	s.render(c, http.StatusOK, "app.html", gin.H{
		"Lang":  lang,
		"Title": appNames[lang],
		// html/template encodes this as JSON that is safe inside <script>.
		"Boot": gin.H{"language": lang, "redirect": redirect, "queryParameters": query},
	})
}

func (s *Site) docs(c *gin.Context) {
	s.render(c, http.StatusOK, "docs.html", nil)
}

func (s *Site) page(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, err := fs.ReadFile(s.files, name)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", raw)
	}
}

func (s *Site) comparePage(c *gin.Context) {
	uuids := strings.Split(c.Query("uuids"), ",")
	if c.Query("uuids") == "" {
		c.Redirect(http.StatusFound, "/")
		return
	}
	result, err := s.compare(uuids)
	if err != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}
	lang := language(c, []string{"sv", "en", "de"}, "sv")
	s.render(c, http.StatusOK, "compare.html", gin.H{"Lang": lang, "Title": appNames[lang], "Data": result})
}

// analyze passes audio uploads to the identification service unchanged.
func (s *Site) analyze() gin.HandlerFunc {
	target, err := url.Parse(s.cfg.AnalyzeEndpoint)
	if s.cfg.AnalyzeEndpoint == "" || err != nil {
		return func(c *gin.Context) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"message": "identification is not available"})
		}
	}
	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.Out.URL = target
			r.Out.Host = target.Host
		},
	}
	return func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
