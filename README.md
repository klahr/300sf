# 300 Svenska Fåglar

A bird checklist service: keep species lists, report sightings, compete with
friends and see rarity per region.

One Go server runs everything on one domain, 300sf.se — the landing page on
`/`, the app on `/home`, the API on `/api/v1` and `/v1`, and its docs on
`/docs` — and ships as a single container.

| Directory | What |
|---|---|
| [`server/`](server/) | Go server: the REST API (gin, MariaDB, Redis, Mailjet) and the websites, embedded from `server/site/` |
| [`frontend/`](frontend/) | TypeScript source of the app; compiled into `server/site/static/js/` at build time |
| [`cli/`](cli/) | Python command line client, also used for admin tasks |
| [`deploy/`](deploy/) | `docker-compose.yml` and the reverse proxy setup |

## Local development

Needs Go, Node (only to compile the app), MariaDB and Redis.

```bash
(cd frontend && npm install && npm run dev)     # compiles the app in watch mode; leave running
cd server
cp config.yml.sample config.yml                # see server/README.md for the database
go run . -config config.yml
```

With `mode: debug` the server reads `server/site/` from disk, so CSS and app
changes show up on reload. Open <http://localhost:8090> for the landing page,
<http://localhost:8090/home> for the app and <http://localhost:8090/docs> for
the API docs.

## Tests

```bash
(cd server && go test ./...)
(cd frontend && npm run typecheck)
(cd cli && python -m unittest discover -s tests)
```

CI runs these and builds the image on every push and pull request.

## Deploy

```bash
docker build -t 300sf .
cd deploy && docker compose up -d --build
```

See [`deploy/README.md`](deploy/README.md) for the config file and the nginx
setup.

## License

The server is GPL-2.0-or-later (`server/COPYING`); the CLI is MIT (`cli/LICENSE`).
