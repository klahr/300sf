# Deploy

One container serves everything on one domain (`site.host`):

| Path | Serves |
|---|---|
| `/` | landing page (signed-in visitors go on to `/home`) |
| `/home`, `/checklists`, `/list/…` and the other app pages | the app |
| `/api/v1/…` | the API, as used by the app |
| `/v1/…` | the API, for API key users |
| `/docs` | API docs (Swagger UI) |

For a first installation, start with [`SELFHOSTING.md`](../SELFHOSTING.md);
this page has the details.

## Run

```bash
sudo mkdir -p /etc/300sf
sudo cp ../server/config.yml.sample /etc/300sf/300sf.yml   # fill in secrets
make -C .. deploy    # docker compose up -d --build, with VERSION from git describe
docker compose logs -f
```

Build through `make` (or set `VERSION` yourself): a plain `docker compose up
--build` stamps the release as `dev`, and browsers then aren't told that a
new release is out.

`CONFIG_FILE` overrides the config path, `PORT` the port used by the health
check. The container uses host networking so it reaches MariaDB and Redis on
the host. Logs go to `docker compose logs` (rotated, 5 × 10 MB) instead of
files. Set `address: "127.0.0.1:8090"` so only the reverse proxy can reach the
server; it trusts `X-Forwarded-For` from loopback only, for per-client rate
limits.

`site.client-id` / `site.client-secret` are the API key the app signs in
with (`300sf -bootstrap` makes it). `site.analyze-endpoint` is the bird sound
identification service: the `analyzer` container in `docker-compose.yml`
(BirdNET-Analyzer, built from `../analyzer`) answers on
`http://127.0.0.1:8095/analyze`. Leave the setting empty to disable
identification; the app then says it is unavailable. Only signed-in users can
reach it (`/api/v1/analyze`). BirdNET's model is licensed CC BY-NC-SA 4.0, so
it may only be used non-commercially.

Keep `hashing-secret`: API keys are stored hashed with it, so changing it
invalidates every API key. `token-secret` signs sessions; change it (`openssl
rand -hex 32`) whenever it may have leaked, which only signs everyone out.
Redis holds the sessions, so flushing it signs everyone out too; give it a
password (`requirepass`, and `redis-password` in the config) and bind it to
`127.0.0.1`. MariaDB, too, should listen on `127.0.0.1` only.

## Upgrading

1. Back up the database.
2. Check out the new release (`git fetch --tags && git checkout <version>`) and
   read its section in [`CHANGELOG.md`](../CHANGELOG.md).
3. Apply the migrations at the end of `server/db.sql` that are newer than your
   database.
4. Add any new settings from `server/config.yml.sample` you want.
5. `make -C .. deploy`.

## Reverse proxy

TLS terminates on the host. Send your domain, and any old domains in
`site.redirect-hosts`, to the same port and pass the `Host` header through
(the server needs it to redirect old domains). The certificate must cover the
old domains for their redirects to work. For nginx:

```nginx
upstream sf {
    server 127.0.0.1:8090;
}

server {
    listen 443 ssl;
    http2 on;
    # Your domain, plus any old domains the server redirects.
    server_name your.domain;

    # ssl_certificate / ssl_certificate_key

    client_max_body_size 20m;   # bird sound uploads

    location / {
        proxy_pass http://sf;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

`site.redirect-hosts` redirects old domains permanently to the same path on
`site.host` (on 300sf.se, `app.300sf.se/` → `/home` and `api-docs.300sf.se/`
→ `/docs`), so bookmarks and links in old emails keep working. API requests on
them are answered instead of redirected, so old API clients keep working.

## Zero-downtime deploys

With one container, `up --build` has a few seconds of downtime while it
restarts. If that matters, run a second copy with `address: "127.0.0.1:8091"`
in its own config and list both in the nginx `upstream`; nginx retries the
other while one restarts.

## 300sf.se

Notes for the installation at 300sf.se.

**Next release** (offline reporting, zones from OpenStreetMap):

1. Apply the migrations that add `zones.osm_id` and turn announcements into
   news posts (end of `server/db.sql`).
2. Add `token-secret` with a new `openssl rand -hex 32`. Sessions were signed
   with `hashing-secret` until now; the server refuses to start with the two
   the same. Everyone signs in once more.
3. Set `requirepass` in Redis and the same as `redis-password`.
4. Importing from the old app is gone: remove the `/legacy/` route from
   nginx, and the old site behind it.
5. Add `geocoder-endpoint: "https://nominatim.openstreetmap.org"` to turn on
   zones from OpenStreetMap, and `contact-email: "hej@300sf.se"` to keep the
   contact form on. Check `domain` is `https://300sf.se`: it names the server
   to OpenStreetMap.

**Moving from the old setup** (one container instead of five, done once):

1. Merge the API config (`300sf-albatross.conf`) and the webapp's
   `routes/config.js` into `/etc/300sf/300sf.yml` (see `site:` in the sample).
2. Stop the old containers: `docker rm -f albatross blackbird yellowhammer woodpecker swagger-ui`.
3. `make -C .. deploy`, then reload nginx with `server_name 300sf.se
   www.300sf.se app.300sf.se api.300sf.se api-docs.300sf.se;`.
4. Set `domain: "https://300sf.se"`, keep the existing `hashing-secret`.
5. Check `https://300sf.se/`, signing in on `https://300sf.se/home`,
   `https://300sf.se/docs`, `https://300sf.se/health`, and that
   `https://app.300sf.se/` redirects.

Everyone signed in once more after the move: sessions live in cookies on
`300sf.se` instead of `localStorage` on `app.300sf.se`.
