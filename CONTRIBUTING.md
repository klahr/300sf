# Contributing

Thanks for helping. Bug reports, translations and code are all welcome.

## Before you start

- **Bugs and ideas:** open an issue. For anything bigger than a small fix,
  describe it in an issue first, so we can agree on the approach.
- **Security problems:** don't open an issue; see [`SECURITY.md`](SECURITY.md).

## Set up

See [Local development](README.md#local-development) in the README: Go, Node,
MariaDB and Redis. `make help` lists the common tasks.

## Make a change

- Keep a pull request to one change, and say what it does and why.
- Match the code around you: its naming, comments and idioms. Go is
  formatted with `gofmt`.
- Add a test where the change can break, and run `make test` before you push.
  Changes to the API or the database also get a test in `server/db_*_test.go`,
  which runs the API against a real database: `make test-db` (needs Docker).
  CI runs both, plus the image build.
- Note user-facing changes under `[Unreleased]` in [`CHANGELOG.md`](CHANGELOG.md),
  one line each.
- Database changes go at the end of `server/db.sql` as a migration that works
  on existing databases.

## Translations

The app is in Swedish, English and German. Edit the texts in
`frontend/translations/locale/{sv,en,de}.json`, then run `npm run translations`
in `frontend/` to update `src/Translations.ts`; CI checks that they match. A new
text needs all three languages; if you only know one, say so in the pull request.

## License

300sf is licensed under the [AGPL-3.0-or-later](LICENSE). By contributing, you
agree that your contribution is licensed the same way.
