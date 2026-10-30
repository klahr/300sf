// SPDX-License-Identifier: AGPL-3.0-or-later
package geocoder

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"
)

// Around every test point but the far one.
const (
	square = `{"type":"Polygon","coordinates":[[[0,40],[30,40],[30,70],[0,70],[0,40]]]}`
	far    = `{"type":"Polygon","coordinates":[[[100,0],[101,0],[101,1],[100,1],[100,0]]]}`
)

func boundary(id, level int, name string) string {
	return boundaryAt(id, level, name, square)
}

func boundaryAt(id, level int, name, shape string) string {
	return fmt.Sprintf(`{"osm_type":"relation","osm_id":%d,"category":"boundary","type":"administrative","name":%q,
		"namedetails":{"name":%q,"name:en":"%s (en)"},"extratags":{"admin_level":"%d"},"geojson":%s}`, id, name, name, name, level, shape)
}

// A fake Nominatim answering by zoom, counting requests.
func fake(t *testing.T, byZoom map[int]string) (*Client, *[]int) {
	t.Helper()
	var mu sync.Mutex
	var asked []int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/search" {
			mu.Lock()
			asked = append(asked, -1)
			mu.Unlock()
			w.Write([]byte(byZoom[-1]))
			return
		}
		if r.URL.Path != "/reverse" {
			t.Errorf("asked %s; only /reverse and /search may be used", r.URL.Path)
		}
		if r.Header.Get("User-Agent") != "300sf-test/1 (+https://example.org)" {
			t.Errorf("User-Agent = %q", r.Header.Get("User-Agent"))
		}
		zoom, _ := strconv.Atoi(r.URL.Query().Get("zoom"))
		mu.Lock()
		asked = append(asked, zoom)
		mu.Unlock()
		body, ok := byZoom[zoom]
		if !ok {
			body = `{"error":"Unable to geocode"}`
		}
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	c := New(srv.URL, "300sf-test/1 (+https://example.org)")
	return c, &asked
}

func TestLookupSweden(t *testing.T) {
	// In Uppsala zoom 10 is the town, which names its kommun; zoom 8 is the län and zoom 5
	// already the country.
	c, asked := fake(t, map[int]string{
		10: `{"osm_type":"node","osm_id":25735371,"category":"place","type":"city","name":"Uppsala",
			"address":{"city":"Uppsala","municipality":"Uppsala kommun","county":"Uppsala län","country":"Sverige"}}`,
		// Search finds a namesake elsewhere first; only the one around the point counts.
		-1: "[" + boundaryAt(1, 7, "Uppsala kommun", far) + "," + boundary(305455, 7, "Uppsala kommun") + "]",
		8:  boundary(54220, 4, "Uppsala län"),
		5:  boundary(52822, 2, "Sverige"),
		3:  boundary(52822, 2, "Sverige"),
	})
	areas, err := c.lookup(context.Background(), 59.86, 17.64, time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if areas.Country == nil || areas.Country.OsmId != "R52822" || areas.Region == nil || areas.Region.OsmId != "R54220" || areas.Municipality == nil || areas.Municipality.OsmId != "R305455" {
		t.Fatalf("areas = %+v", areas)
	}
	if fmt.Sprint(*asked) != "[10 -1 8 5]" {
		t.Errorf("asked %v (-1 is the search); the country came at zoom 5, so zoom 3 is not needed", *asked)
	}
	if got := areas.Region.LocalName("en"); got != "Uppsala län (en)" {
		t.Errorf("English name = %q", got)
	}
	if got := areas.Region.LocalName("sv"); got != "Uppsala län" {
		t.Errorf("no Swedish name falls back to the local one, got %q", got)
	}
}

func TestLookupGermany(t *testing.T) {
	// In Munich the city is a Kreis (6) at zooms 9 and 8, Bayern (4) at 5, and the country only at 3.
	c, asked := fake(t, map[int]string{
		10: boundary(62428, 6, "München"),
		8:  boundary(62428, 6, "München"),
		5:  boundary(2145268, 4, "Bayern"),
		3:  boundary(51477, 2, "Deutschland"),
	})
	areas, err := c.lookup(context.Background(), 48.14, 11.58, time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if areas.Country.OsmId != "R51477" || areas.Region.OsmId != "R2145268" || areas.Municipality.OsmId != "R62428" {
		t.Fatalf("areas = %+v", areas)
	}
	if len(*asked) != 4 {
		t.Errorf("asked zooms %v", *asked)
	}
}

func TestLookupSkipsWhatIsNotABoundary(t *testing.T) {
	// A town's point and the open sea are no areas.
	c, _ := fake(t, map[int]string{
		10: `{"osm_type":"node","osm_id":1,"category":"place","type":"city","name":"Uppsala"}`,
		5:  boundary(52822, 2, "Sverige"),
	})
	areas, err := c.lookup(context.Background(), 59.86, 17.64, time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if areas.Municipality != nil || areas.Region != nil || areas.Country == nil {
		t.Fatalf("areas = %+v", areas)
	}
}

func TestWaitKeepsThePace(t *testing.T) {
	c := New("", "")
	start := time.Now()
	for i := 0; i < 3; i++ {
		if err := c.wait(context.Background(), 50*time.Millisecond); err != nil {
			t.Fatal(err)
		}
	}
	// The first goes at once, the other two 50 ms apart.
	if d := time.Since(start); d < 100*time.Millisecond {
		t.Errorf("3 requests in %v, want at least 100ms", d)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := c.wait(ctx, time.Hour); err == nil {
		t.Error("a cancelled wait should give up")
	}
}

func TestSearchKeepsRegions(t *testing.T) {
	c, _ := fake(t, map[int]string{
		// Searching "Uppsala" finds the town, the kommun and the län; only the län is a region.
		-1: `[` + boundary(305455, 7, "Uppsala kommun") + `,
			{"osm_type":"node","osm_id":1,"category":"place","type":"city","name":"Uppsala"},
			{"osm_type":"relation","osm_id":54220,"category":"boundary","type":"administrative","name":"Uppsala län",
			 "namedetails":{"name":"Uppsala län","name:en":"Uppsala County"},"extratags":{"admin_level":"4"},
			 "address":{"county":"Uppsala län","country":"Sverige"}}]`,
	})
	found, err := c.Search(context.Background(), "Uppsala")
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 1 || found[0].OsmId != "R54220" || found[0].Country != "Sverige" || found[0].LocalName("en") != "Uppsala County" {
		t.Fatalf("found %+v", found)
	}
}

func TestRegionRefusesOtherIds(t *testing.T) {
	c := New("http://127.0.0.1:1", "")
	for _, id := range []string{"", "N123", "R12x", "R"} {
		if _, err := c.Region(context.Background(), id); err == nil {
			t.Errorf("Region(%q) should refuse it", id)
		}
	}
}
