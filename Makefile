.PHONY: help test test-db image deploy major minor patch _bump latest

.DEFAULT_GOAL := help

# The release: the latest X.Y.Z tag, plus commits since and -dirty when the
# tree has changes. The app shows it and names its files by it, so browsers
# and the service worker pick up a new release. See CHANGELOG.md.
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
export VERSION

help:
	@echo "300sf — common targets"
	@echo ""
	@echo "    make test       Server tests and the app's typecheck"
	@echo "    make test-db    The API against a real database and Redis (Docker)"
	@echo "    make image      Build the Docker image, stamped with the version"
	@echo "    make deploy     Build and (re)start the production stack in deploy/"
	@echo ""
	@echo "  Release"
	@echo "    make patch|minor|major"
	@echo "                    Promote CHANGELOG [Unreleased], commit, tag"
	@echo "    make latest     Print the latest release tag"
	@echo ""
	@echo "  Version now: $(VERSION)"

test:
	cd server && go test ./...
	cd frontend && npm run typecheck

# The database tests (server/db_test.go) against throwaway MariaDB and Redis
# containers, removed afterwards, with the schema from server/db.sql.
TEST_DB_PORT ?= 33306
TEST_REDIS_PORT ?= 36379
test-db:
	@set -e; \
	docker rm -f sf-test-db sf-test-redis >/dev/null 2>&1 || true; \
	trap 'docker rm -f sf-test-db sf-test-redis >/dev/null 2>&1' EXIT; \
	docker run -d --name sf-test-db -e MARIADB_ROOT_PASSWORD=test -e MARIADB_DATABASE=300sf -p 127.0.0.1:$(TEST_DB_PORT):3306 mariadb:11 >/dev/null; \
	docker run -d --name sf-test-redis -p 127.0.0.1:$(TEST_REDIS_PORT):6379 redis:7-alpine >/dev/null; \
	echo "Waiting for MariaDB..."; \
	for i in $$(seq 1 60); do docker exec sf-test-db mariadb -uroot -ptest -e "SELECT 1" >/dev/null 2>&1 && break; sleep 1; done; \
	docker exec -i sf-test-db mariadb -uroot -ptest 300sf < server/db.sql; \
	cd server && TEST_DATA_SOURCE="root:test@tcp(127.0.0.1:$(TEST_DB_PORT))/300sf?charset=utf8mb4&parseTime=True&loc=Local" \
		TEST_REDIS="127.0.0.1:$(TEST_REDIS_PORT)" go test -run '^TestDB' -count=1 .

image:
	docker build --build-arg VERSION=$(VERSION) -t 300sf .

deploy:
	cd deploy && docker compose up -d --build

# --- Release -----------------------------------------------------------------

# major/minor/patch cut an annotated release tag, bumping from the latest
# release tag (0.0.0 if none). CHANGELOG.md's [Unreleased] entries move
# under a new [X.Y.Z] - date heading, an empty [Unreleased] is left behind,
# and the change is committed so the tag points at it. An EMPTY
# [Unreleased] aborts: nobody wrote down what ships.
#
#   make patch   0.1.0 -> 0.1.1
#   make minor   0.1.0 -> 0.2.0
#   make major   0.1.0 -> 1.0.0
#
# The tag is local; the recipe prints the push command.
major:
	@$(MAKE) --no-print-directory _bump BUMP=major
minor:
	@$(MAKE) --no-print-directory _bump BUMP=minor
patch:
	@$(MAKE) --no-print-directory _bump BUMP=patch

# set -e: a failed `git commit` must not fall through to `git tag`, which
# would tag the previous HEAD as the new release.
_bump:
	@set -e; \
	CUR=$$($(LATEST_TAG)); CUR=$${CUR:-0.0.0}; \
	MAJOR=$$(echo "$$CUR" | cut -d. -f1); \
	MINOR=$$(echo "$$CUR" | cut -d. -f2); \
	PATCH=$$(echo "$$CUR" | cut -d. -f3); \
	case "$(BUMP)" in \
		major) MAJOR=$$((MAJOR + 1)); MINOR=0; PATCH=0 ;; \
		minor) MINOR=$$((MINOR + 1)); PATCH=0 ;; \
		patch) PATCH=$$((PATCH + 1)) ;; \
		*) echo "Use: make major|minor|patch"; exit 1 ;; \
	esac; \
	NEW="$$MAJOR.$$MINOR.$$PATCH"; \
	if grep -q "^## \[$$NEW\]" CHANGELOG.md; then \
		echo "CHANGELOG.md already has a [$$NEW] heading — leaving it alone"; \
	else \
		if [ "$$(awk '/^## \[Unreleased\]/{f=1;next} /^## \[/{f=0} f' CHANGELOG.md | grep -c '[^[:space:]]')" -eq 0 ]; then \
			echo "REFUSING: CHANGELOG.md [Unreleased] is empty — nothing to release."; \
			echo "Write what ships under [Unreleased] first, or add a [$$NEW] heading by hand."; \
			exit 1; \
		fi; \
		awk -v h="## [$$NEW] - $$(date +%F)" '{print} /^## \[Unreleased\]$$/ && !d {print ""; print h; d=1}' \
			CHANGELOG.md > CHANGELOG.md.tmp && mv CHANGELOG.md.tmp CHANGELOG.md; \
		echo "CHANGELOG.md: [Unreleased] promoted to [$$NEW]"; \
	fi; \
	git add CHANGELOG.md; \
	git commit -q -m "Release $$NEW" -- CHANGELOG.md; \
	git tag -a "$$NEW" -m "Release $$NEW"; \
	echo "$$CUR -> $$NEW (committed CHANGELOG.md, tagged)"; \
	echo "Push with: git push origin main $$NEW"

# The newest RELEASE tag: only X.Y.Z, no pre-releases, ordered numerically
# (0.10.0 beats 0.9.0). Expanded by the recipes, so plain `make` runs no git.
LATEST_TAG = git tag -l '[0-9]*.[0-9]*.[0-9]*' --sort=-v:refname | grep -Ex '[0-9]+\.[0-9]+\.[0-9]+' | head -1

latest:
	@$(LATEST_TAG)
