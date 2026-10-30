// SPDX-License-Identifier: AGPL-3.0-or-later

// Package site serves the websites from the same binary as the API, all on
// one host: the landing page on /, the app on /home and its other pages, and
// the API docs on /docs. Everything they need is embedded, so the server is a
// single file.
package site

import (
	"crypto/rand"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/klahr/300sf/server/config"
	"github.com/klahr/300sf/server/controllers"
	"github.com/klahr/300sf/server/models"
)

//go:embed static templates pages locales
var embedded embed.FS

// Version is the release being served, set by main from the build
// (git describe); "dev" when run from source. Asset links carry it, so a
// browser never mixes files from two releases, and so does the service
// worker, so a release is a new one browsers pick up.
var Version = "dev"

// Client-side routes of the app; the page is chosen in the browser.
var appPages = []string{
	"home", "sign-up", "forgot", "settings", "checklists", "contribute", "donation",
	"contact", "tags", "reports", "set-password", "delete-account", "competitions",
	"admin", "experimental", "news",
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
	stats   func() (*controllers.SiteStats, error)

	statsMu   sync.Mutex
	statsAt   time.Time
	statsLast *controllers.SiteStats
}

// New serves the embedded site. Pass a directory (os.DirFS("site")) as files
// during development to pick up changes without rebuilding; nil uses the
// embedded copy.
func New(cfg config.Site, files fs.FS) (*Site, error) {
	if files == nil {
		files = embedded
	}
	s := &Site{files: files, cfg: cfg, landing: map[string]map[string]any{}, compare: controllers.CompareLists, stats: controllers.GetSiteStats}

	funcs := template.FuncMap{
		"year":   func() int { return time.Now().Year() },
		"asset":  assetURL,
		"number": formatNumber,
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
	files := r.Group("", cacheHeaders)
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			sub, _ := fs.Sub(static, name)
			files.StaticFS("/"+name, http.FS(sub))
		} else if name != "sw.js" && name != "robots.txt" && name != "sitemap.xml" {
			files.StaticFileFS("/"+name, name, http.FS(static))
		}
	}
	r.GET("/sw.js", s.serviceWorker)
	r.GET("/robots.txt", s.robots)
	r.GET("/sitemap.xml", s.sitemap)

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
	r.GET("/terms", s.page("terms"))
	r.GET("/privacy-policy", s.page("privacy-policy"))
	r.GET("/compare", s.comparePage)
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
	r.ClientIP = c.ClientIP()
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

// ContentSecurityPolicy is the policy for every response. Pages may not be
// framed by other sites (clickjacking), and scripts only run from this site,
// from Stripe (the donation page's buy button) or, inline, when they carry
// the page's nonce; an injected script would carry none.
func ContentSecurityPolicy(nonce string) string {
	scripts := "'self' https://js.stripe.com"
	if nonce != "" {
		scripts += " 'nonce-" + nonce + "'"
	}
	return "frame-ancestors 'none'; base-uri 'self'; object-src 'none'; script-src " + scripts
}

func newNonce() string {
	b := make([]byte, 16)
	// crypto/rand.Read never fails, it fills b entirely or crashes the process.
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// render executes a template with data, plus Nonce for its inline scripts.
func (s *Site) render(c *gin.Context, status int, name string, data gin.H) {
	if data == nil {
		data = gin.H{}
	}
	data["Nonce"] = newNonce()
	c.Header("Content-Security-Policy", ContentSecurityPolicy(data["Nonce"].(string)))
	c.Header("Content-Type", "text/html; charset=utf-8")
	// Pages name the release's assets, so they must never outlive it.
	c.Header("Cache-Control", "no-cache")
	c.Status(status)
	if err := s.templates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Error(err)
	}
}

func (s *Site) landingData(c *gin.Context, signedInGoesToApp bool) gin.H {
	lang := language(c, []string{"en", "sv"}, "en")
	return gin.H{"Lang": lang, "T": s.landing[lang], "SignedInGoesToApp": signedInGoesToApp, "Stats": s.siteStats(), "SourceUrl": s.cfg.SourceUrl}
}

// siteStats are the landing page's community numbers, counted at most every 15 minutes. Nil when
// they could not be counted; the page then leaves them out.
func (s *Site) siteStats() *controllers.SiteStats {
	s.statsMu.Lock()
	defer s.statsMu.Unlock()
	if time.Since(s.statsAt) < 15*time.Minute {
		return s.statsLast
	}
	stats, err := s.stats()
	if err != nil {
		stats = nil
	}
	s.statsAt = time.Now()
	s.statsLast = stats
	return stats
}

// formatNumber groups thousands the way the language does: 20 401 in Swedish, 20,401 in English.
func formatNumber(n int64, lang string) string {
	sep := ","
	if lang == "sv" {
		sep = "\u00a0"
	}
	digits := strconv.FormatInt(n, 10)
	var b strings.Builder
	for i, d := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			b.WriteString(sep)
		}
		b.WriteRune(d)
	}
	return b.String()
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
		"Boot": gin.H{
			"language": lang, "redirect": redirect, "queryParameters": query, "version": Version,
			"mapTiles": s.cfg.MapTiles, "mapAttribution": s.cfg.MapAttribution, "sourceUrl": s.cfg.SourceUrl,
		},
	})
}

func (s *Site) docs(c *gin.Context) {
	s.render(c, http.StatusOK, "docs.html", nil)
}

// page serves pages/<name>.html, or its Swedish version pages/<name>-sv.html.
func (s *Site) page(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		file := "pages/" + name + ".html"
		if language(c, []string{"en", "sv"}, "en") == "sv" {
			file = "pages/" + name + "-sv.html"
		}
		raw, err := fs.ReadFile(s.files, file)
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
	// Same limit as the API's compare, so list ids can't be guessed here instead.
	if controllers.Throttled("compare/"+controllers.ClientNetwork(c.ClientIP()), 60, time.Minute) {
		c.String(http.StatusTooManyRequests, "Too many requests, try again in a minute.")
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

// assetURL names a static file as the current release has it.
func assetURL(path string) string {
	return path + "?v=" + url.QueryEscape(Version)
}

// cacheHeaders lets browsers keep a file asked for by the current release
// (its ?v=) for good: a later release asks for it under another name. Any
// other file is checked with the server each time, which answers 304 while
// the release is the same.
func cacheHeaders(c *gin.Context) {
	switch {
	case Version == "dev":
		c.Header("Cache-Control", "no-cache")
	case c.Query("v") == Version:
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
	default:
		c.Header("Cache-Control", "no-cache")
		// Embedded files have no modification time to compare against.
		c.Header("ETag", `"`+Version+`"`)
	}
	c.Next()
}

// serviceWorker serves sw.js with the release in it. Browsers install a
// service worker again only when its bytes change, so without this a
// release would go unnoticed.
func (s *Site) serviceWorker(c *gin.Context) {
	raw, err := fs.ReadFile(s.files, "static/sw.js")
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	body := strings.Replace(string(raw), `const VERSION = "dev";`, fmt.Sprintf("const VERSION = %q;", Version), 1)
	c.Header("Cache-Control", "no-cache")
	c.Data(http.StatusOK, "text/javascript; charset=utf-8", []byte(body))
}

// robots and sitemap name this installation's host, not 300sf.se.
func (s *Site) robots(c *gin.Context) {
	c.String(http.StatusOK, "User-agent: *\nAllow: /\nDisallow: /api/\nSitemap: https://%s/sitemap.xml\n", s.cfg.Host)
}

func (s *Site) sitemap(c *gin.Context) {
	c.Header("Content-Type", "application/xml; charset=utf-8")
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n" + `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	for _, path := range []string{"/", "/terms", "/privacy-policy", "/docs"} {
		fmt.Fprintf(&b, "\t<url><loc>https://%s%s</loc></url>\n", s.cfg.Host, path)
	}
	b.WriteString("</urlset>\n")
	c.String(http.StatusOK, b.String())
}
