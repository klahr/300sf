# Self-hosting

The bare minimum to run your own 300sf. For the details (nginx, old domains,
the bird sound analyzer) see [`deploy/README.md`](deploy/README.md).

## You need

- Docker with Compose
- MariaDB (10.6 or later) and Redis on the same host
- A domain with HTTPS. Sessions use `Secure` cookies, so signing in only
  works over HTTPS (or on `http://localhost`).

## 1. Database

The schema uses the database name `300sf`.

```bash
mysql -e 'CREATE DATABASE `300sf` CHARACTER SET utf8mb4'
mysql 300sf < server/db.sql
```

## 2. Config

```bash
sudo mkdir -p /etc/300sf
sudo cp server/config.yml.sample /etc/300sf/300sf.yml
```

Set at least:

| Key | Value |
|---|---|
| `data-source` | your MariaDB user, password and host |
| `hashing-secret` | `openssl rand -hex 32`; changing it later signs everyone out and breaks API keys |
| `domain` | `https://your.domain`, used in email links |
| `site.host` | `your.domain` |
| `site.redirect-hosts` | `{}`, unless you have old domains to redirect |

Leave `site.client-id` and `site.client-secret` empty for now.

## 3. Start

```bash
make deploy        # builds the image and starts it on 127.0.0.1:8090
```

This also builds the bird sound analyzer, which is large. To leave it out,
run `cd deploy && VERSION=$(git describe --tags --always) docker compose up -d --build 300sf`
instead.

Point your HTTPS reverse proxy at `127.0.0.1:8090` and pass the `Host` header
through ([nginx example](deploy/README.md#reverse-proxy)).

## 4. First admin

Nobody can sign in yet: the website signs in with an API key, and API keys
are made by signed-in users. So:

1. Sign up at `https://your.domain/sign-up`.
2. Add the starter birds, make yourself an admin and create the website's
   key:

   ```bash
   cd deploy
   docker compose run --rm 300sf ./300sf -config /config/300sf.yml -seed -bootstrap you@example.com
   ```

3. Copy the printed `client-id` and `client-secret` into `site:` in the
   config, then `docker compose restart 300sf`.
4. Sign in. The Admin page is in the account menu.

## 5. Fill it

`-seed` added the 542 species observed in Sweden and their orders, named in
Swedish, English, German and Latin ([sources](server/seed/README.md)). Add
more species on the Admin page. People build checklists from list templates;
make those on the Admin page too.

Zones (country, region, municipality) make themselves: the first sighting in a
new place is looked up on OpenStreetMap, anywhere in the world. That is
`geocoder-endpoint` in the config, OpenStreetMap's public Nominatim by default.
Its [usage policy](https://operations.osmfoundation.org/policies/nominatim/)
fits a birding site; for heavy use, run your own Nominatim. Map tiles are
OpenStreetMap's too, under their [tile policy](https://operations.osmfoundation.org/policies/tiles/);
`site.map-tiles` moves them to another service.

Optional:

- **Email** (password resets, invitations): set the Mailjet keys and
  `send-email-interval` (seconds, for example `60`). Without email, people can
  still sign up and sign in, but can't reset a forgotten password.
- **Bird sound identification**: see [`deploy/README.md`](deploy/README.md).
  BirdNET's model is licensed CC BY-NC-SA 4.0, for non-commercial use only.

## Make it yours

Some things still say 300sf.se. Change them before you invite people:

- the sender addresses in `server/email-templates/*/*.yml`
- `contact-email` in the config, for the app's contact form (off when empty)
- `site.source-url`, if you change the code: the AGPL requires offering your
  users its source
- the terms and privacy policy in `server/site/pages/`. These describe the
  300sf.se service, not yours.

## Upgrading

```bash
git fetch --tags && git checkout <version>     # see CHANGELOG.md
# apply the migrations at the end of server/db.sql that are newer than your database
make deploy
```

Back up the database first.
