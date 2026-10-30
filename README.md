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
| [`deploy/`](deploy/) | `docker-compose.yml` and the reverse proxy setup |

To run your own, see [`SELFHOSTING.md`](SELFHOSTING.md). To help, see
[`CONTRIBUTING.md`](CONTRIBUTING.md). To report a security problem, see
[`SECURITY.md`](SECURITY.md).

## Local development

Needs Go, Node (only to compile the app), MariaDB and Redis.

```bash
(cd frontend && npm install && npm run dev)     # compiles the app in watch mode; leave running
cd server
cp config.yml.sample config.yml                # see server/README.md for the database
sed -i "s/change-me/$(openssl rand -hex 32)/; s/\"release\"/\"debug\"/" config.yml
go run . -config config.yml
```

The server won't start until `hashing-secret` is set. To sign in, sign up
at <http://localhost:8090/sign-up>, run
`go run . -config config.yml -seed -bootstrap you@example.com` (`-seed` adds
the starter birds), put the printed key
under `site:` in `config.yml` and restart. With `mode: debug` it
reads `server/site/` from disk, so CSS and app changes show up on reload. Open <http://localhost:8090> for the landing page,
<http://localhost:8090/home> for the app and <http://localhost:8090/docs> for
the API docs.

## Tests

```bash
(cd server && go test ./...)
(cd frontend && npm run typecheck)
```

CI runs these and builds the image on every push and pull request.

## Releases

```bash
make patch    # or minor / major: promotes CHANGELOG.md's [Unreleased], commits and tags
git push origin main <version>
```

Write what ships under `[Unreleased]` in [`CHANGELOG.md`](CHANGELOG.md) first;
an empty section stops the release. The build stamps the version from
`git describe`; the app shows it in Settings, names its CSS and JavaScript by
it, and puts it in the service worker, so browsers drop the old release's
files and an open app offers to reload.

## Deploy

```bash
make image    # docker build, stamped with the version
make deploy   # cd deploy && docker compose up -d --build, stamped likewise
```

A plain `docker build` works too, but the version is then `dev`, and browsers
won't be told about the new release.

See [`deploy/README.md`](deploy/README.md) for the config file and the nginx
setup.

## License

Copyright © 2022–2026 Angels' Ware.

300sf (server, web app, analyzer and artwork: logo, icons and illustrations)
is licensed under the GNU Affero General Public License v3.0 or later; see
[`LICENSE`](LICENSE). If you run a
modified version as a service, its users must be able to get its source. Set
`site.source-url` to where they can; the app links it in its footer and Settings.

The libraries, fonts and data it ships with are under their own licenses; see
[`THIRD_PARTY_NOTICES.md`](THIRD_PARTY_NOTICES.md).

Sound identification uses the BirdNET model, which is licensed separately
under CC BY-NC-SA 4.0 (non-commercial use only).
