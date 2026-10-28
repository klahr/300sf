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

Logs go to `$LOG_FILE_LOCATION` (stderr if unset).

## Tests

```bash
go test ./...
```

## Authentication

API clients call `POST /v1/auth` with `grant_type: password` (plus their
`client_id`/`client_secret` API key, `email`, `password`, `scopes`) or
`grant_type: refresh_token`, and send the access token as `Authorization:
Bearer <token>`. `POST /v1/auth/logout` with the refresh token ends the
session. Scopes are `user` and `admin`.

- Tokens are HS256 JWTs signed with `hashing-secret` and typed: only access
  tokens are accepted on API calls.
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
`site/static/swagger.yaml` is partial.

## Zones

Zones are GeoJSON polygons nested country → county → municipality
(`admin_level` 2, 4, 7). Reports are matched to zones by point-in-polygon
(`utils/location.go`). Zones are inserted with `300sf zone --insert` from the
CLI.
