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

### Fixed

- Browsers could keep running old code after a release.
- The health check reported the wrong version.
- robots.txt and the sitemap named 300sf.se on every installation.
- The API docs sent "Try it out" requests to 300sf.se.
- Translations are edited in the locale files again; the converter broke quotes.
- The 512 px app icon was missing from the web app manifest.
