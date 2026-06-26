# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

SlashNode is a self-hosting "OS" for a personal server: a single Go binary (`slashnoded`) that installs on Debian/Ubuntu and runs a web App Store to deploy Docker apps (Bitcoin Core, LND, Electrs, Mempool, Supabase, MinIO, Jitsi, SlashSlack…). Apps are JSON manifests that embed a docker-compose document.

## Build / verify

There is **no automated test suite** and the canonical checks are build + vet + typecheck. Go and Node may not be installed locally — build in containers when needed:

```bash
# Go (daemon)
go build ./... && go vet ./...
# or, without a local toolchain:
docker run --rm -v "$PWD":/src -w /src golang:1.24 bash -c "go build ./... && go vet ./..."

# Web (Next.js front)
cd web && npm install && npx tsc --noEmit            # typecheck (the main web gate)
docker run --rm -v "$PWD/web":/web -w /web node:22 bash -c "npm ci && npx tsc --noEmit"

# Release artifacts (linux/darwin amd64+arm64 binaries, web bundle, app catalog tarball)
./scripts/build.sh v0.1.0   # → dist/
```

Tests are written ad-hoc when verifying a change (e.g. `go test ./internal/apps/ -run TestX`); there are no committed `_test.go` files. Always build+vet Go and `tsc` the web before committing.

Run the daemon locally without root (prefixes all paths under `$SLASHNODE_ROOT`):
```bash
export SLASHNODE_ROOT=/tmp/sn
go run ./cmd/slashnoded init
go run ./cmd/slashnoded serve --web-dir web   # front :8080, API :8081
```
Env toggles: `SLASHNODE_NO_DOCKER`, `SLASHNODE_NO_TOR`, `SLASHNODE_NO_CADDY` (force those subsystems off), `SLASHNODE_LATEST` (short-circuit the update check).

## Architecture (big picture)

**Two servers in one `serve`** (`internal/cli/serve.go`):
- **Go API** on `127.0.0.1:8081`, Bearer-token auth (`bearer()`), called only server-side by the front. The token never reaches the browser.
- **Public front** on `0.0.0.0:8080`: serves `/__console` (container-shell WebSocket, in Go) and reverse-proxies everything else to a supervised `next start` subprocess (internal port `Port+10000`). The Next API routes (`web/app/api/**/route.ts`) proxy to the Go API, injecting the token.

**Apps** (`internal/apps`, `apps/<id>/slashnode-app.json`): each manifest carries a real **docker-compose YAML** in its `compose` field. `installOne` (`orchestrate.go`) templates `${input.X}`/`${secret.X}`/`${app.exports.key}` (+ synthetic `${node.exports.host}` / `${self.exports.url}`) into the compose via `templateRefs`/`composeEscape`, applies per-service image-tag overrides, writes `docker-compose.yml`, and runs `docker compose -p slashnode-<id> -f`. Dependencies auto-install (post-order `installPlan`) and wire via the registry (an app's `exports` → `registry.json` → consumers reference `${dep.exports.*}`). The compose `image:` is parsed by a **line-based** parser (`orchestrator.ParseComposeImages`), so the `compose` field must be YAML, not JSON.

**State** lives under `/var/lib/slashnode` (or `$SLASHNODE_ROOT/...`): `config.json` (`internal/config`), `secrets.json`, `apps.json` (install state incl. `image_tags`, `subdomain`, `migration_version`), `app-secrets.json` (per-app secrets, plaintext mode 0600), `registry.json` (exports), `update.json`, `schema.json` (migration versions). Paths are centralized in `internal/paths`.

**Reverse proxy + Tor**: `internal/apps/proxy.go` regenerates a Caddyfile routing `<subdomain>.<host>` → each app's web port (subdomain defaults to the app id, overridable per app). `internal/apps/tor.go` + `internal/tor` generate a torrc giving the UI and each app a `.onion` (web on `:80`, plus declared endpoint ports). On `serve` startup the daemon reconciles both against installed apps so onions/routes get provisioned.

**Auth model**: the web UI **always requires admin login**. Login verifies the password and gets a per-login, expiring **HMAC session token** (`exp.hmac`, keyed by `SessionSecret`) issued by the Go API, verified by `web/middleware.ts` (Web Crypto) and the Go console. Never reintroduce an "open mode".

**Updates + migrations**: `internal/updater` self-updates the binary + web/apps bundles from GitHub releases (checksum-verified) and restarts. `internal/migrate` runs ordered, snapshotted **node-state migrations** on `init` and `serve` startup; manifests can declare **per-app declarative migrations** (`migrations` block, run during `Reapply`). Add migrations append-only; never renumber.

## Conventions / gotchas

- **All user input that reaches compose/config/Caddyfile is validated/escaped** — `validateInput` (control chars, select/number), `composeEscape` (`\` `"` `$`), `validImageTag`, `validHost`, DNS-label subdomains. Keep new sinks behind these.
- Manifest `command:` must be **YAML list form** when it templates a value (docker-compose shlex-splits a string command, breaking values with spaces).
- Volume names are explicit (`slashnode-<id>_<id>_<vol>`) so data survives upgrades and cross-app volume sharing works.
- The compose `image:` parser and tag overrides operate on the rendered YAML; image tags must match the docker tag charset.
- Git workflow used here: work on `develop`, then `--no-ff` merge `develop` → `master`. **Pushing `master` triggers the release workflow** (`.github/workflows/release.yml`), refreshing the rolling `latest` release.
