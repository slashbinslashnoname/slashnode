<pre>
      //
     //
    //
   //
  //
 //
</pre>

# SlashNode

**Orchestrator for self-hosted services driven by JSON manifests, built on
Docker.** Designed for Bitcoin/Lightning nodes and self-hosting:
plug-and-play (`exports`/`wiring`), App Store, light/dark theme.

> This is not a custom OS. It's a Go daemon (`slashnoded`) that installs on
> any existing Debian/Ubuntu with a single line.

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/slashbinslashnoname/slashnode/master/bootstrap.sh | bash
```

Crypto-conscious audience? Audit before running:

```bash
curl -fsSL https://raw.githubusercontent.com/slashbinslashnoname/slashnode/master/bootstrap.sh -o slashnode.sh
less slashnode.sh
bash slashnode.sh
```

The bootstrap stays minimal: it installs Docker + Node, drops the
`slashnoded` binary (checksum verified), deploys the front end and delegates all
init to `slashnoded init`. Once done, the node is reachable at
**http://slashnode.local:8080** (mDNS/Avahi).

## Access & password

During installation the bootstrap asks how the node should be reached:

- **local** (default) — reachable on your LAN as `slashnode.local`. Open by
  default; you can optionally set an admin password to require a login.
- **server** — a public address (e.g. `node.example.com`); always
  password-protected with a login page.

This is stored in `config.json` under `access` (`mode`, `address`,
`password_protected`). You can also set it non-interactively:

```bash
slashnoded init --access server --address node.example.com --password '…' --password-protect
slashnoded init --access local --password '…' --password-protect   # local + login
```

When password protection is on, the web UI redirects to a login page that
checks the admin password.

## Architecture

| Component       | Role                                                            |
|-----------------|-----------------------------------------------------------------|
| `slashnoded`    | Go daemon, single binary. Runs the local API **and** the front end. |
| Next.js front end | UI (light/dark theme, red primary), launched by the daemon.   |
| Local Go API    | `127.0.0.1`, Bearer token auth. Consumed by the front end.     |
| systemd         | `slashnoded.service` + `slashnoded-update.timer` (update check). |
| Avahi           | mDNS → `slashnode.local`.                                       |
| Docker          | Runtime for apps (JSON manifests, see `docs/app-manifest.md`).  |

```
slashnoded serve
 ├─ Go API        127.0.0.1:8081   (/api/v1/status, /api/v1/update[/apply])
 └─ next start    0.0.0.0:8080     (front end, proxies the API via the token)
```

## Commands

```
slashnoded init           # config + secrets + systemd + Avahi (idempotent)
slashnoded serve          # Go API + supervised Next.js front end
slashnoded status         # node status (--post-install: URL + credentials)
slashnoded update         # apply the latest binary update (--to <tag>)
slashnoded check-update   # check for an update (called by the timer; notify-only)
slashnoded uninstall      # remove service + binary (--purge: data too)
```

## Updates

**Notify-only** policy by default: the systemd timer checks the latest release
daily and persists the state (`/var/lib/slashnode/update.json`). The UI then
shows a banner with an **"Apply"** button that triggers
`slashnoded update` (download + checksum verification + atomic
replacement of the binary + restart). Configurable in `config.json`
(`update.policy`, `update.channel`).

## Development

```bash
# Front end
cd web && npm install && npm run build

# Daemon (tests without root: SLASHNODE_ROOT prefixes all paths)
export SLASHNODE_ROOT=/tmp/sn
go run ./cmd/slashnoded init
go run ./cmd/slashnoded serve --web-dir web   # front end on :8080, API on :8081
```

Build the release artifacts (amd64/arm64 + macOS binaries, web bundle):

```bash
./scripts/build.sh v0.1.0   # → dist/
```

## Apps (manifests)

Each app = a JSON manifest (Docker services, `inputs` entered by the
user → environment variables, `exports`/`wiring` for automatic
wiring). See **[docs/app-manifest.md](docs/app-manifest.md)** and the
examples in `examples/`.

## License

To be defined.
