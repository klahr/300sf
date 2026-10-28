# 300 Svenska Fåglar

A bird checklist service: keep species lists, report sightings, compete with
friends and see rarity per region.

| Directory | What | Runs as |
|---|---|---|
| [`api/`](api/) | Go REST API (gin, MariaDB, Redis, Mailjet) | `api.300sf.se` |
| [`webapp/`](webapp/) | Single-page app (TypeScript) served by Express, which also proxies `/api/*` to the API and adds the client credentials on login | `app.300sf.se` |
| [`web/`](web/) | Landing page (Express + EJS) | `300sf.se` |
| [`cli/`](cli/) | Python command line client, also used for admin tasks | – |
| [`deploy/`](deploy/) | `docker-compose.yml` for the production stack and swagger-ui | – |

## Local development

```bash
# API, see api/README.md for database setup
cd api && cp config.yml.sample config.yml && go run . -config config.yml

# Webapp on :3000, talking to the API on :8090
cd webapp && npm install
npm run dev     # compiles TypeScript in watch mode; leave running
API_ENDPOINT=http://localhost:8090 CLIENT_ID=... CLIENT_SECRET=... npm start

# Landing page on :8080
cd web && npm install && npm start
```

## Tests

```bash
(cd api && go test ./...)
(cd webapp && npm test)
(cd web && npm test)
(cd cli && python -m unittest discover -s tests)
```

CI runs all of them on every push and pull request.

## Deploy

```bash
cd deploy && docker compose up -d --build
```

The API image needs its config file (`300sf-albatross.conf`,
`300sf-blackbird.conf`) in `api/`, and the webapp needs `webapp/routes/config.js`
(see `config.js.template`) or the equivalent environment variables. These hold
secrets and are not committed.

## License

The API is GPL-2.0-or-later (`api/COPYING`); the CLI is MIT (`cli/LICENSE`).
