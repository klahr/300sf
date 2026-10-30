# Changelog

All notable changes to 300 Svenska Fåglar will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
Versions are git tags `X.Y.Z`, cut with `make patch|minor|major`.

## [Unreleased]

### Added

- Report sightings offline; they are sent when the connection is back.
- More accurate GPS position when reporting, with its accuracy shown.
- The report map starts where you last were.
- Zones are created automatically from OpenStreetMap, anywhere in the world.
- Competition counties can be chosen on a map or found by name.
- The version is shown in Settings.
- An open app offers to reload when a new version is released.
- Starter data with the 542 species observed in Sweden (`300sf -seed`).
- `300sf -bootstrap` sets up the first admin on a new installation.
- A guide for self-hosting (`SELFHOSTING.md`).
- The map tile service can be changed in the config.
- A link to the source code, as the AGPL requires (`site.source-url`).
- The contact form's address is set in the config (`contact-email`).
- Licenses of the bundled libraries, fonts and data (`THIRD_PARTY_NOTICES.md`).
- A contributing guide, issue templates and Dependabot.
- Tests of the API against a real database and Redis (`make test-db`), also in CI.
- API clients can let a new account get its first API key with its email and
  password (`POST /v1/api_keys/password`), without the website.
- `GET /v1/config` tells API clients the server's map tiles, source link and
  which optional features (sound identification, contact, region search) are on.
- Comparing checklists names each list (`result[].list`), in the order asked
  for, so two lists of one person can be told apart; the compare page shows it.
- Inviting to a competition lists who already takes part (`already_in`)
  instead of reporting them as invited.
- `POST /v1/reports/{uuid}/add_list` and `remove_list` move a sighting between
  checklists without resending all of it.
- Sightings take a bare date (`"2026-05-03"`) for a date-only time.
- The CSV export is named after its checklist (`Content-Disposition`).
- A news feed: announcements are posts with a title, kept under the bell in
  the top bar (News in the sidebar), with unread ones counted and shown on
  the home page. New accounts don't start with old posts unread.
- Admins write posts in English, Swedish and German with simple formatting
  and a preview, edit them, schedule them, publish and take them down, and
  see how many have read each one.
- Everyone can make API keys for other apps and scripts, under Settings → API keys;
  the client ID they sign in with is shown with each key. Admins can give a key admin access.
- Read-only API keys, which can look but not change anything (the `read` scope).
- Connect Claude and other AI assistants: add `https://300sf.se/mcp` as a connector, approve it
  on the site, and ask about your checklists and sightings or add new ones. Choose whether it may
  only read. Connected apps are listed, and can be disconnected, under Settings → API keys.
- API: an MCP endpoint (`/mcp`) and OAuth 2.1 with dynamic registration and PKCE
  (`/oauth/register`, `/oauth/authorize`, `/oauth/token`, `/oauth/revoke`, `/.well-known/oauth-*`).
- Complete API documentation at `/docs`.
- Edit a sighting, including which bird it was, from the sighting's page.
- Choose who can see a checklist on comparison links: only you, anyone with the link, or everyone.
- Change your display name, username and email in your profile; the old address is told
  about a new one.
- API: the same, as `PUT /v1/reports/:uuid`, `POST /v1/lists/:uuid/visibility` and
  `POST /v1/users/display_name`, `/username` and `/email`.

### Changed

- API: signing in without `scopes` gives the `user` scope instead of none.
- API: `POST /v1/api_keys` also returns the key's `client_id`, `prefix` and `scopes`, and takes `scopes`.
- API: `POST /v1/reports` returns the sighting with its species, place and other details,
  and `GET /v1/reports/:uuid` includes the sex.
- API: a checklist's `from` and `until` take a date or an RFC 3339 time, and anything
  else is refused instead of ignored.
- Signing up with a username that is taken says so (409, code 4006) instead of failing.

### Removed

- Importing sightings from the old 300 Svenska Fåglar app (`/v1/users/migrate`),
  and `imported_at` on users.

### Security

- Sessions are signed with a new `token-secret`, separate from
  `hashing-secret`, so it can be rotated without breaking API keys. The
  server refuses to start without one, with one shorter than 32 characters,
  or with one equal to `hashing-secret`. Everyone signs in once more.
- Wrong passwords are counted per account and network, per network, and per
  account from anywhere, so a few tries no longer lock someone out and
  passwords can't be sprayed across accounts from one address. IPv6
  addresses are limited per /64.
- Tokens from a revoked API key no longer work on admin and API key routes.
- Lists and private competitions with old 6-character ids no longer open
  for anyone but their owners and members: those ids could be guessed.
- Competition passwords, invites and sound identification are rate limited;
  joining can no longer go past the competition's maximum.
- Competitions can start at most a year back, and standings show the day of
  a sighting, not its time.
- Pages only run scripts from this site, Stripe, or inline with the page's
  nonce (Content-Security-Policy `script-src`).
- Redis can be given a password (`redis-password`).

### Fixed

- Sightings came with their zone's whole outline, which made sighting lists
  much larger than needed.
- A checklist from an unknown or someone else's private template is refused
  with 400 (code 1001), not a server error, and leaves no name behind.
- Resuming a paused announcement did nothing.
- Text too long for the database (a name, label, site or note) is refused with
  400 (code 2001, "… is too long"), and a duplicate with 400 (code 1002),
  instead of a server error with the database's message.
- Setting a checklist's visibility without one made it private; a checklist or
  template with a visibility other than 0, 1 or 2 was stored.
- Labels could be empty, have any color text, or be put on a sighting while
  archived (and then not show); adding one twice gave a database error.
- Sightings with a negative quantity, a position off the globe or an end
  before their start were stored.
- Signing in asking only for scopes the account or key lacks gave a token no
  route accepts; it is now refused (code 3005).
- Parse errors on the display name, username and template species routes have
  code 2001; an unknown language no longer says "could not find species".

- Browsers could keep running old code after a release.
- The health check reported the wrong version.
- robots.txt and the sitemap named 300sf.se on every installation.
- The API docs sent "Try it out" requests to 300sf.se.
- Translations are edited in the locale files again; the converter broke quotes.
- The 512 px app icon was missing from the web app manifest.
