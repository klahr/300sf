# Deploy

One container serves everything on one domain, `300sf.se` (`site.host`):

| Path | Serves |
|---|---|
| `/` | landing page (signed-in visitors go on to `/home`) |
| `/home`, `/checklists`, `/list/…` and the other app pages | the app |
| `/api/v1/…` | the API, as used by the app |
| `/v1/…` | the API, for the CLI and other API key users |
| `/docs` | API docs (Swagger UI) |

The old domains in `site.redirect-hosts` redirect permanently to the same path
on `300sf.se` (`app.300sf.se/` → `/home`, `api-docs.300sf.se/` → `/docs`), so
bookmarks and links in old emails keep working. API requests on them are
answered instead of redirected, so `api.300sf.se/v1/…` keeps working for
existing API clients.

Everyone signs in once more after this version is deployed: sessions now live
in cookies on `300sf.se` instead of `localStorage` on `app.300sf.se`, and
tokens issued before it are no longer accepted. API key users (such as the
CLI) sign in again with their password. Redis now holds the sessions, so
flushing it signs everyone out.

## Run

```bash
sudo mkdir -p /etc/300sf
sudo cp ../server/config.yml.sample /etc/300sf/300sf.yml   # fill in secrets
docker compose up -d --build
docker compose logs -f
```

`CONFIG_FILE` overrides the config path, `PORT` the port used by the health
check. The container uses host networking so it reaches MariaDB and Redis on
the host. Logs go to `docker compose logs` (rotated, 5 × 10 MB) instead of
files.

`site.client-id` / `site.client-secret` are the API key the app signs in
with (previously `ClientId`/`ClientSecret` in the webapp's
`routes/config.js`). `site.analyze-endpoint` is the bird sound identification
service: the `analyzer` container in `docker-compose.yml` (BirdNET-Analyzer,
built from `../analyzer`) answers on `http://127.0.0.1:8095/analyze`. Leave the
setting empty to disable identification; the app then says it is unavailable.
Only signed-in users can reach it (`/api/v1/analyze`). BirdNET's model is
licensed CC BY-NC-SA 4.0, so it may only be used non-commercially.

Apply the migrations at the end of `server/db.sql` that are newer than your
database before starting a new version.

## Reverse proxy

TLS terminates on the host. Send `300sf.se` and the old domains to the same
port and pass the `Host` header through (the server needs it to redirect the
old domains). The certificate must still cover the old domains for their
redirects to work. For nginx:

```nginx
upstream sf {
    server 127.0.0.1:8090;
}

server {
    listen 443 ssl;
    http2 on;
    # 300sf.se plus the old domains, which the server redirects.
    server_name 300sf.se www.300sf.se app.300sf.se api.300sf.se api-docs.300sf.se;

    # ssl_certificate / ssl_certificate_key as today

    client_max_body_size 20m;   # bird sound uploads

    location / {
        proxy_pass http://sf;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

This replaces the previous upstreams (webapp on `:3000`, landing page on
`:8080`, swagger-ui on `:8070`, and the two API containers).

## Moving from the old setup

1. Merge the API config (`300sf-albatross.conf`) and the webapp's
   `routes/config.js` into `/etc/300sf/300sf.yml` (see `site:` in the sample).
   Point the CLI's `api-endpoint` at `https://300sf.se`.
2. Stop the old containers: `docker rm -f albatross blackbird yellowhammer woodpecker swagger-ui`.
3. `docker compose up -d --build`, then reload nginx with the config above.
4. Set `domain: "https://300sf.se"` in the config so emails link to the new
   address. Keep the existing `hashing-secret`: it signs sessions and API
   keys are stored hashed with it, so changing it signs everyone out and
   invalidates every API key. The server refuses to start without one.
   Set `address: "127.0.0.1:8090"` so only nginx can reach the server; it
   trusts `X-Forwarded-For` from loopback only, for per-client rate limits.
5. If importing from the old site should keep working, keep routing
   `/legacy/` on 300sf.se to it in nginx; everything else goes to the
   container.
6. Check `https://300sf.se/`, signing in on `https://300sf.se/home`,
   `https://300sf.se/docs`, `https://300sf.se/health`, and that
   `https://app.300sf.se/` redirects.

## Zero-downtime deploys

The old setup ran two API containers. With one container, `up --build` has a
few seconds of downtime while it restarts. If that matters, run a second copy
with `address: "127.0.0.1:8091"` in its own config and list both in the nginx
`upstream`; nginx retries the other while one restarts.
