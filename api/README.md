# 300sf API

REST API for 300 Svenska Fåglar: users, checklists, list templates, reports
(observations), tags, competitions, zones and species. Go + gin, MariaDB via
gorm, Redis for caching and API rate limits, Mailjet for email.

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

`POST /v1/auth` with `grant_type: password` (plus a `client_id`/`client_secret`
API key, `email`, `password`, `scopes`) or `grant_type: refresh_token`. Returns
HS256 JWTs signed with `hashing-secret`; send them as `Authorization: Bearer
<token>`. Scopes are `user` and `admin`. Routes are declared in `routes.go`;
the OpenAPI spec in `swagger.yaml` is partial.

## Zones

Zones are GeoJSON polygons nested country → county → municipality
(`admin_level` 2, 4, 7). Reports are matched to zones by point-in-polygon
(`utils/location.go`). Zones are inserted with `300sf zone --insert` from the
CLI.
