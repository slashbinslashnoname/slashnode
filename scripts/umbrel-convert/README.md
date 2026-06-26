# umbrel-convert

Best-effort converter from [getumbrel/umbrel-apps](https://github.com/getumbrel/umbrel-apps)
manifests to SlashNode app manifests, **for review** — output lands in
`../../converted/`, never directly in the shipped `apps/` catalog.

```bash
npm install
npm start                       # convert the built-in "essentials" allowlist
npm start -- jellyfin navidrome # convert specific app ids
```

Each app yields `converted/<id>/slashnode-app.json`, plus a `converted/REPORT.md`
summarizing what converted cleanly (✓), what still needs human work (△), and a
per-app **security review** (🔒).

## What it translates

- `umbrel-app.yml` metadata → manifest fields (id/name/version/category/…).
- the `app_proxy` service → a `web` block + publishing the web port **on
  `127.0.0.1` only**, so SlashNode's Caddy (admin login + HTTPS) and Tor are the
  only ways in — never a public bind.
- `${APP_DATA_DIR}/…` bind mounts → named volumes (`slashnode-<id>_<svc>_<key>`).
- the shared external `slashnode` network + `container_name` per service.
- `image@sha256` digests stripped; a few known env vars remapped
  (`${DEVICE_DOMAIN}` → `${node.exports.host}`, `${APP_*PASSWORD/SEED/…}` →
  `${secret.*}` inputs).

## What it can't translate (flagged for review)

- Umbrel infra env vars it doesn't know (`${APP_BITCOIN_NODE_IP}`,
  `${TOR_PROXY_IP}`, app-specific domains/ports) — left verbatim and listed.
- shared `${UMBREL_ROOT}/data/storage` mounts (mapped to a private volume).
- non-web published ports (left **unpublished** by default — add by hand).
- elevated privileges (privileged, `cap_add`, host devices, host network mode,
  `docker.sock`) and shipped default passwords — preserved but surfaced in the
  security review so you decide before promoting.

## Promoting an app

After reviewing/fixing `converted/<id>/`, move it into `apps/<id>/` and verify it
loads (`go test`/build). Resolve every △ token and 🔒 flag first.
