# 300sf server

The REST API for 300 Svenska Fåglar (users, checklists, list templates,
reports, tags, zones and species) and the websites around it: the app, the
landing page and the API docs, embedded from `site/`. Go + gin, MariaDB via
gorm, Redis for caching and API rate limits, Mailjet for email.

The app's JavaScript (`site/static/js/app.js`) is built from `../frontend`;
run `npm run build` (or `npm run dev`) there before starting the server.

## Running locally

Requires Go (see `go.mod`), MariaDB and Redis.

```bash
mysql -e 'CREATE DATABASE `300sf`'
mysql 300sf < db.sql

cp config.yml.sample config.yml   # then fill in credentials and secret
go build -o bin/300sf
./bin/300sf -config ./config.yml
curl localhost:8090/health
```

A new database has no users and no API key for the website to sign in with.
Sign up on the website, then `./bin/300sf -config ./config.yml -bootstrap
you@example.com` makes you an admin and prints the key for `site.client-id` /
`site.client-secret`.

Logs go to `$LOG_FILE_LOCATION` (stderr if unset).

## Tests

```bash
go test ./...
```

## Authentication

API clients call `POST /v1/auth` with `grant_type: password` (plus their
`client_id`/`client_secret` API key, `email`, `password` and optionally
`scopes`, `["user"]` when left out) or
`grant_type: refresh_token`, and send the access token as `Authorization:
Bearer <token>`. `POST /v1/auth/logout` with the refresh token ends the
session. Scopes are `user` and `admin`. Users make their own API keys in the
app under Settings → API keys (or `POST /v1/api_keys`, or with email and
password on `POST /v1/api_keys/password` for a first key); a key gets the
`user` scope, and `admin` too when an admin asks for it. The full API is
documented at `/docs` (`site/static/swagger.yaml`).

- Tokens are HS256 JWTs signed with `token-secret` and typed: only access
  tokens are accepted on API calls. `token-secret` is separate from
  `hashing-secret` (which API keys are hashed with), so it can be rotated
  without breaking API keys.
- Refresh tokens are single use. Each refresh replaces the token (its id is
  kept in Redis), so a stolen and replayed token stops working, and logging
  out revokes it. Refreshing re-checks the API key and the user's scopes.
- After 10 failed sign-ins for an email address, sign-in is refused for 15
  minutes.

The app (`/api/v1/...`) never handles tokens. The server adds the app's API
key (`site.client-id`/`site.client-secret`) on sign-in and keeps the tokens in
`HttpOnly`, `Secure`, `SameSite=Strict` cookies: the access token is sent only
to `/api`, the refresh token only to `/api/v1/auth`. A readable `sf_session`
cookie tells the page it is signed in and when to refresh. Without "stay
signed in" the cookies end with the browser session. `/api` answers no CORS
requests, and writes must carry `X-Requested-With`, so other sites can't act
with a visitor's session.

Routes are declared in `routes.go`; the OpenAPI spec in
`site/static/swagger.yaml` documents every one of them, so keep it in step.

## Zones

Zones are GeoJSON polygons nested country → region → municipality
(`admin_level` 2, 4, 7; a country without regions has its municipalities
straight under it). A sighting is matched to the zones it lies in by
point-in-polygon (`utils/location.go`).

Zones are made as people report: a sighting in a place no zone covers is
looked up with OpenStreetMap's Nominatim after it is saved (`places/`,
`geocoder/`), and the country, region and municipality found become zones,
keyed by their OpenStreetMap relation (`osm_id`). The next sighting there needs
no lookup. Sightings left without a zone (geocoder down, or made before this)
are filled in by a background pass every 15 minutes.

The public Nominatim's [usage policy](https://operations.osmfoundation.org/policies/nominatim/)
allows at most one request a second for the whole site, an identifying
User-Agent (`300sf/<version> (+<domain>)`), cached answers and no use of
`/details`; the geocoder package keeps to it, and the background pass to the
4 a minute asked of scheduled work. Point `geocoder-endpoint` at your own
Nominatim for more; leave it empty to turn lookups off.

Zones can still be added by hand on the Admin page, from a GeoJSON file. A
looked-up area with the same name and level as a hand-made zone takes that
zone over instead of making a twin.
