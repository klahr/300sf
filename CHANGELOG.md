# Changelog

All notable changes to 300 Svenska Fåglar will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
Versions are git tags `X.Y.Z`, cut with `make patch|minor|major`.

## [Unreleased]

### Added

- CI also publishes the bird sound identification service with every release,
  as `ghcr.io/klahr/300sf-analyzer`; `COMPOSE_PROFILES=analyzer` in
  `deploy/.env` runs it, and `make upgrade` pulls it along.

### Changed

- The species' songs and calls play in the app's own player, with play and
  pause, a bar to move in the recording and the time, one recording at a time,
  instead of the browser's controls, which made the species details scroll
  sideways on a phone.
- The landing page only shows bird sound identification when it is on
  (`site.analyze-endpoint`): the feature, its question and the example in the
  illustration.

### Fixed

- A checklist that got a new id when it was shared can still be opened,
  renamed, recoloured, archived and exported with its old id by its owner, on a
  device that hasn't caught up yet; for anyone else the old id stays closed.
- The admin statistics count days, weeks and months in the database's own time
  zone, so the newest signups are no longer left out around midnight.
- After a failed save, the share sheet no longer sends a choice made meanwhile
  behind the error message.
- Share pages no longer reload when a sighting's municipality is looked up but
  was already known.
- With `wikipedia-cache-duration` set to 0, a species picture that Wikipedia
  failed to give is asked for again within the hour instead of never.

## [0.2.0] - 2026-10-31

### Added

- `site.time-zone` in the config: the time zone share pages tell days in and
  competitions count them in (Europe/Stockholm when empty).
- Statistics for admins, under Admin → Statistics: accounts, signups, weekly
  and monthly active users, sightings, checklists and competitions created per
  week, retention by signup month, shared checklists and recent competitions'
  participants. Active means having reported a sighting.
- Admins set the donation page's progress, raised and goal, under
  Admin → Donation; it was fixed in the code before.
- API: `GET /v1/admin/stats` (admin), `GET /v1/donation` and `PUT /v1/donation` (admin).
- Prebuilt images: CI publishes `ghcr.io/klahr/300sf:X.Y.Z` and `:latest` for
  every release, `deploy/compose.prod.yaml` runs it, and `make upgrade` pulls
  and pins the latest release, so the server builds nothing.
- Welcome posts: a news post marked "Welcome new accounts" shows on the
  dashboard of accounts made in the last 30 days, however old the post is,
  and never as unread to older accounts (`welcome` on announcements).

### Changed

- The landing page links to the donation page, from the footer and the
  answer on what it costs, and explains why the site can be free. Signed out,
  the donation page shows no empty profile picture, and the logo leads back
  to the start page in the same language.
- The Swedish landing page is written in Swedish birders' own words
  (kryss, årslista, skådare) instead of translated from English.
- Joining a public competition asks you to agree to showing in which
  municipality you find species too, as its share page's map shows that.
- API: a competition's `latest` and `ended_at` days start at midnight in the
  site's time zone instead of UTC.

### Fixed

- Species' songs and calls play again: xeno-canto closed its API v2, so no
  species had recordings. The server uses API v3, which needs a key
  (`xeno-canto-key` in the config); without one there are no recordings, and
  the landing page's bird search no longer promises them.

- A checklist's share page counts sightings that have no start time, instead
  of showing their species as not seen.
- Share pages work without Redis, and a species' picture is looked up once,
  however many screens show the page.
- A competition's share page keeps updating when looking up who takes part
  fails for a moment, and many open pages no longer hold each other up.
- The map on a competition's share page leaves out sightings after the
  competition was won, like the standings, and loads faster.
- A competition's share page counts the days left in calendar days, and says
  "1 day left".
- Share pages show days in Swedish time, so a competition starting 1 October
  no longer shows 30 September, nor a sighting just after midnight the day before.
- Share pages update when a sighting's municipality is found after it was
  saved, and when the species list a list or competition follows changes.
- Choosing who can see a checklist while it is still being shared no longer
  fails and leaves it shared.
- A competition's share page offers joining only while it is open for joining.
- Sightings saved offline still land on a checklist that got a new id when it
  was shared since.
- A species' picture that Wikipedia failed to give is asked for again within
  the hour, and species without an English name no longer share one picture.
- A competition's share page and live updates query the database less.
- The admin user details show the user's language.
- The bird search on the landing page clears the last bird's picture as soon
  as another is chosen, shimmers while the new one loads and shows the 300sf
  mark when a species has no picture; its pictures are sharp, in a size that
  fits the screen, and its results are no longer covered by the text below them.

### Security

- Species search and species details are rate limited for anyone not signed
  in, also when an invalid token is sent; a valid token's API key is checked
  as before, including revoked keys and their limits.

## [0.1.0] - 2026-10-31

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
- Choose who can see a checklist's share page and comparison links: only you or anyone
  with the link.
- Share pages: every checklist that isn't private has one (`/share/<list>`), and every
  public competition (`/share/competition/<id>`) with its standings, species, newest
  finds and a map of where species were found. Made for a TV, they update themselves
  as sightings come in, and never show where anyone was or their notes.
- Creating or joining a competition asks you to agree to showing your display name, and in a
  public competition your finds by day, to the others.
- Find a bird on the landing page: search all species, read about them and hear
  their song without an account.
- API: `GET /v1/species/search` and `GET /v1/species/:uuid` work without signing in.
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
- Public checklists are now shared with a link: nothing listed them, so it was the
  same. Existing ones are changed by a migration, and a public visibility (2) is
  stored as with a link (1).

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
