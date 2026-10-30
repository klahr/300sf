// SPDX-License-Identifier: AGPL-3.0-or-later

// Package geocoder asks a Nominatim service which administrative areas (the
// country, its region and the municipality) contain a point, with their
// OpenStreetMap ids, names and simplified outlines.
//
// The public service at nominatim.openstreetmap.org is shared and runs on
// donations; its usage policy (https://operations.osmfoundation.org/policies/nominatim/)
// is why this package:
//
//   - sends at most one request a second for the whole server, and 4 a
//     minute for background work (Slowly);
//   - identifies the app in its User-Agent;
//   - asks only for what the user's own sighting needs, never a grid or a
//     list: /reverse, and /search for the one municipality /reverse names
//     but doesn't return (/details is for debugging, not for programs);
//   - is called only when no zone already covers the point, so answers are
//     kept: in the zones table.
//
// The data is OpenStreetMap's, under the ODbL; the app shows the attribution.
package geocoder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tidwall/geojson"
	"github.com/tidwall/geojson/geometry"
)

// Area is one administrative area around the point.
type Area struct {
	OsmId      string // "R54220": OpenStreetMap type letter and id
	AdminLevel int    // OpenStreetMap's admin_level: 2 is the country
	Name       string // in the local language
	Names      map[string]string
	Geometry   json.RawMessage // GeoJSON Polygon or MultiPolygon
}

// LocalName is the area's name in language, or its local name.
func (a Area) LocalName(language string) string {
	if n := a.Names["name:"+language]; n != "" {
		return n
	}
	return a.Name
}

// Areas around a point: any of them can be missing.
type Areas struct {
	Country      *Area
	Region       *Area // the first level below the country: a Swedish län, a German Land
	Municipality *Area // the most local level, from 6 to 8: a Swedish kommun, a German Kreis or Gemeinde
}

const (
	// A little over the policy's 1 a second, so clocks never make it more.
	pace = 1100 * time.Millisecond
	// The policy's limit for scripts that run at intervals.
	slowPace = 15 * time.Second
)

type Client struct {
	endpoint  string
	userAgent string
	http      *http.Client

	mu   sync.Mutex
	last time.Time
}

// New returns a client for the Nominatim at endpoint; userAgent names the
// app and how to reach whoever runs it.
func New(endpoint, userAgent string) *Client {
	return &Client{endpoint: endpoint, userAgent: userAgent, http: &http.Client{Timeout: 20 * time.Second}}
}

// Lookup finds the areas containing the point, for a sighting just made.
func (c *Client) Lookup(ctx context.Context, lat, lon float64) (Areas, error) {
	return c.lookup(ctx, lat, lon, pace)
}

// Slowly is Lookup for background work, at 4 requests a minute.
func (c *Client) Slowly(ctx context.Context, lat, lon float64) (Areas, error) {
	return c.lookup(ctx, lat, lon, slowPace)
}

// Each zoom asks for the area at one rank. Which rank holds which level
// differs by country, so ask from the most local up and sort the answers by
// their real admin_level, stopping as soon as all three are known.
var zooms = []struct {
	zoom      int
	threshold string // simplifies the outline, in degrees
}{
	{10, "0.001"},
	{8, "0.001"},
	{5, "0.005"},
	{3, "0.01"},
}

const localThreshold = "0.001"

func (c *Client) lookup(ctx context.Context, lat, lon float64, gap time.Duration) (Areas, error) {
	var areas Areas
	for _, z := range zooms {
		r, err := c.reverse(ctx, lat, lon, z.zoom, z.threshold, gap)
		if err != nil {
			return areas, err
		}
		if a := r.area(); a != nil {
			areas.add(*a)
		}
		// In a town /reverse answers with the town, not the municipality
		// around it (a Swedish kommun ranks between two zooms), but names it.
		if z.zoom == 10 && areas.Municipality == nil && r.Address["municipality"] != "" {
			a, err := c.municipality(ctx, lat, lon, r.Address, gap)
			if err != nil {
				return areas, err
			}
			if a != nil {
				areas.add(*a)
			}
		}
		// Past zoom 5 only the country is left to find.
		if areas.Country != nil && (areas.Region != nil && areas.Municipality != nil || z.zoom <= 5) {
			break
		}
	}
	return areas, nil
}

func (a *Areas) add(area Area) {
	switch {
	case area.AdminLevel == 2:
		a.Country = &area
	case area.AdminLevel >= 3 && area.AdminLevel <= 5:
		if a.Region == nil || area.AdminLevel < a.Region.AdminLevel {
			a.Region = &area
		}
	case area.AdminLevel >= 6 && area.AdminLevel <= 8:
		if a.Municipality == nil || area.AdminLevel > a.Municipality.AdminLevel {
			a.Municipality = &area
		}
	}
}

type reverseResponse struct {
	Error      string            `json:"error"`
	Address    map[string]string `json:"address"`
	OsmType    string            `json:"osm_type"`
	OsmId      int64             `json:"osm_id"`
	Category   string            `json:"category"`
	Type       string            `json:"type"`
	Name       string            `json:"name"`
	NameDetail map[string]string `json:"namedetails"`
	ExtraTags  map[string]string `json:"extratags"`
	GeoJSON    json.RawMessage   `json:"geojson"`
}

// reverse returns Nominatim's answer at zoom: an administrative area, or
// something else (a town, the sea) that area() turns down.
func (c *Client) reverse(ctx context.Context, lat, lon float64, zoom int, threshold string, gap time.Duration) (reverseResponse, error) {
	q := shapeQuery(threshold)
	q.Set("lat", strconv.FormatFloat(lat, 'f', 5, 64))
	q.Set("lon", strconv.FormatFloat(lon, 'f', 5, 64))
	q.Set("zoom", strconv.Itoa(zoom))
	var r reverseResponse
	err := c.get(ctx, "/reverse", q, gap, &r)
	return r, err
}

// municipality searches for the municipality /reverse named, and keeps the
// administrative area that really contains the point: names repeat.
func (c *Client) municipality(ctx context.Context, lat, lon float64, address map[string]string, gap time.Duration) (*Area, error) {
	var parts []string
	for _, k := range []string{"municipality", "county", "state", "country"} {
		if address[k] != "" {
			parts = append(parts, address[k])
		}
	}
	q := shapeQuery(localThreshold)
	q.Set("q", strings.Join(parts, ", "))
	q.Set("limit", "5")
	var results []reverseResponse
	if err := c.get(ctx, "/search", q, gap, &results); err != nil {
		return nil, err
	}
	point := geojson.NewPoint(geometry.Point{X: lon, Y: lat})
	for _, r := range results {
		a := r.area()
		if a == nil || a.AdminLevel < 6 || a.AdminLevel > 8 {
			continue
		}
		if shape, err := geojson.Parse(string(a.Geometry), nil); err == nil && shape.Contains(point) {
			return a, nil
		}
	}
	return nil, nil
}

// The parameters every request has: names in all languages and the outline.
func shapeQuery(threshold string) url.Values {
	q := url.Values{}
	q.Set("format", "jsonv2")
	q.Set("extratags", "1")
	q.Set("namedetails", "1")
	q.Set("polygon_geojson", "1")
	q.Set("polygon_threshold", threshold)
	return q
}

func (c *Client) get(ctx context.Context, path string, q url.Values, gap time.Duration, into any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint+path+"?"+q.Encode(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", c.userAgent)

	if err := c.wait(ctx, gap); err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("geocoder: %s", resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(into); err != nil {
		return fmt.Errorf("geocoder: %w", err)
	}
	return nil
}

func (r reverseResponse) area() *Area {
	if r.Error != "" || r.OsmType != "relation" || r.Category != "boundary" || r.Type != "administrative" {
		return nil
	}
	level, err := strconv.Atoi(r.ExtraTags["admin_level"])
	if err != nil || len(r.GeoJSON) == 0 {
		return nil
	}
	var shape struct {
		Type string `json:"type"`
	}
	if json.Unmarshal(r.GeoJSON, &shape) != nil || (shape.Type != "Polygon" && shape.Type != "MultiPolygon") {
		return nil
	}
	name := r.NameDetail["name"]
	if name == "" {
		name = r.Name
	}
	return &Area{OsmId: "R" + strconv.FormatInt(r.OsmId, 10), AdminLevel: level, Name: name, Names: r.NameDetail, Geometry: r.GeoJSON}
}

// wait holds a request until gap has passed since the last one, whoever made it.
func (c *Client) wait(ctx context.Context, gap time.Duration) error {
	c.mu.Lock()
	at := c.last.Add(gap)
	if now := time.Now(); at.Before(now) {
		at = now
	}
	c.last = at
	c.mu.Unlock()

	t := time.NewTimer(time.Until(at))
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Candidate is a region found by name, for someone to choose from.
type Candidate struct {
	OsmId   string
	Name    string
	Names   map[string]string
	Country string // in the local language
}

// LocalName is the region's name in language, or its local name.
func (c Candidate) LocalName(language string) string {
	return Area{Name: c.Name, Names: c.Names}.LocalName(language)
}

// Search finds regions (a Swedish län, a German Land) by name, for a person
// choosing one: called when they search, never as they type, which the
// policy forbids.
func (c *Client) Search(ctx context.Context, text string) ([]Candidate, error) {
	q := url.Values{}
	q.Set("format", "jsonv2")
	q.Set("q", text)
	q.Set("limit", "20")
	q.Set("extratags", "1")
	q.Set("namedetails", "1")
	q.Set("addressdetails", "1")
	var results []reverseResponse
	if err := c.get(ctx, "/search", q, pace, &results); err != nil {
		return nil, err
	}
	var found []Candidate
	for _, r := range results {
		if !r.isRegion() {
			continue
		}
		name := r.NameDetail["name"]
		if name == "" {
			name = r.Name
		}
		found = append(found, Candidate{OsmId: "R" + strconv.FormatInt(r.OsmId, 10), Name: name, Names: r.NameDetail, Country: r.Address["country"]})
	}
	return found, nil
}

func (r reverseResponse) isRegion() bool {
	level, err := strconv.Atoi(r.ExtraTags["admin_level"])
	return err == nil && r.OsmType == "relation" && r.Category == "boundary" && r.Type == "administrative" && level >= 3 && level <= 5
}

// Region fetches a region chosen from Search, with its outline, and the
// country around it.
func (c *Client) Region(ctx context.Context, osmId string) (Areas, error) {
	var areas Areas
	if len(osmId) < 2 || osmId[0] != 'R' {
		return areas, fmt.Errorf("geocoder: not a relation: %q", osmId)
	}
	if _, err := strconv.ParseInt(osmId[1:], 10, 64); err != nil {
		return areas, fmt.Errorf("geocoder: not a relation: %q", osmId)
	}
	q := shapeQuery("0.005")
	q.Set("osm_ids", osmId)
	var results []struct {
		reverseResponse
		Lat string `json:"lat"`
		Lon string `json:"lon"`
	}
	if err := c.get(ctx, "/lookup", q, pace, &results); err != nil {
		return areas, err
	}
	if len(results) == 0 || !results[0].isRegion() {
		return areas, fmt.Errorf("geocoder: %s is not a region", osmId)
	}
	areas.Region = results[0].area()
	if areas.Region == nil {
		return areas, fmt.Errorf("geocoder: %s has no outline", osmId)
	}

	// The region's own point (its centre or capital) lies in its country.
	lat, err1 := strconv.ParseFloat(results[0].Lat, 64)
	lon, err2 := strconv.ParseFloat(results[0].Lon, 64)
	if err1 != nil || err2 != nil {
		return areas, fmt.Errorf("geocoder: %s has no position", osmId)
	}
	r, err := c.reverse(ctx, lat, lon, 3, "0.01", pace)
	if err != nil {
		return areas, err
	}
	if a := r.area(); a != nil && a.AdminLevel == 2 {
		areas.Country = a
	}
	return areas, nil
}
