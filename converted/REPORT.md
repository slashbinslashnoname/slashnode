# Umbrel → SlashNode conversion report

Converted 55 apps · ✓ 16 clean · △ 39 need review · ✗ 0 failed · 🔒 14 with security flags.

These manifests are **auto-generated for review** and are NOT in the shipped
catalog (they live here, not under `apps/`). Promote one by moving its folder
into `apps/` after verifying it.

Every converted app publishes its web UI on **127.0.0.1 only** (reached via
Caddy + the admin login, and Tor) — never on a public interface. Non-web
ports are left unpublished by default.

## 🔒 Security review (per app)

Apps below use elevated-privilege constructs or ship a default credential —
review each before promoting. Apps not listed declared nothing notable.

### `adguard-home`
- ⚠️ server: extra capabilities ["NET_ADMIN"]
- ⚠️ server: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
### `calibre-web`
- ⚠️ ships a default password ("admin123") — must be changed on first login
### `frigate`
- ⚠️ web: runs privileged
### `grafana`
- ⚠️ ships a default password ("admin") — must be changed on first login
### `grocy`
- ⚠️ ships a default password ("admin") — must be changed on first login
### `home-assistant`
- ⚠️ server: runs privileged
- ⚠️ server: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
### `mosquitto`
- ⚠️ broker: declares ports ["1883:1883"] — left UNpublished, add by hand if required
### `node-red`
- ⚠️ ships a default password ("moneyprintergobrrr") — must be changed on first login
### `portainer`
- ⚠️ ships a default password ("changeme") — must be changed on first login
- ⚠️ docker: runs privileged
- ⚠️ docker: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
### `qbittorrent`
- ⚠️ ships a default password ("adminadmin") — must be changed on first login
### `tailscale`
- ⚠️ web: extra capabilities ["NET_ADMIN","NET_RAW"]
- ⚠️ web: maps host devices ["/dev/net/tun:/dev/net/tun"]
- ⚠️ web: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
### `umami`
- ⚠️ ships a default password ("umami") — must be changed on first login
### `wireguard`
- ⚠️ app: extra capabilities ["NET_ADMIN","SYS_MODULE"]
### `zigbee2mqtt`
- ⚠️ app: runs privileged

## ✓ Clean
- `code-server`
- `excalidraw`
- `freshrss`
- `jellyseerr`
- `librespeed`
- `mealie`
- `memos`
- `minio`
- `n8n`
- `overseerr`
- `stirling-pdf`
- `syncthing`
- `tautulli`
- `uptime-kuma`
- `vaultwarden`
- `wordpress`

## △ Needs review
- `adguard-home` — 
- `audiobookshelf` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `bazarr` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `calibre-web` — 
- `duplicati` — unmapped tokens: APP_PASSWORD; shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `forgejo` — unmapped tokens: APP_DOMAIN, APP_FORGEJO_SSH_PORT
- `frigate` — unmapped tokens: APP_PASSWORD
- `ghost` — unmapped tokens: DEVICE_DOMAIN_NAME
- `gitea` — unmapped tokens: APP_DOMAIN, APP_GITEA_SSH_PORT
- `grafana` — 
- `grocy` — 
- `homarr` — unmapped tokens: APP_SEED
- `home-assistant` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `immich` — unmapped tokens: APP_SEED
- `invidious` — unmapped tokens: APP_INV_SECRET_KEY, APP_SEED
- `jackett` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `jellyfin` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `lidarr` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `mosquitto` — 
- `navidrome` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `nextcloud` — unmapped tokens: NETWORK_IP, APP_DOMAIN, APP_NEXTCLOUD_PORT, APP_HIDDEN_SERVICE, APP_NEXTCLOUD_LOCAL_IPS, APP_PASSWORD
- `node-red` — unmapped tokens: ${APP_LIGHTNING_NODE_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR
- `owncloud` — unmapped tokens: DEVICE_DOMAIN_NAME, APP_OWNCLOUD_LOCAL_IPS, APP_PASSWORD
- `photoprism` — unmapped tokens: APP_DOMAIN, APP_PASSWORD
- `plausible` — unmapped tokens: DEVICE_DOMAIN_NAME, APP_SEED, APP_PLAUSIBLE_VAULT_KEY
- `portainer` — 
- `prowlarr` — unmapped tokens: APP_PROWLARR_RADARR_CONFIG_XML, APP_PROWLARR_LIDARR_CONFIG_XML, APP_PROWLARR_SONARR_CONFIG_XML, APP_PROWLARR_READARR_CONFIG_XML; shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `qbittorrent` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `radarr` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `readarr` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `sabnzbd` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `searxng` — unmapped tokens: APP_SEED
- `sonarr` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `tailscale` — 
- `transmission` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `umami` — unmapped tokens: APP_SEED
- `vikunja` — unmapped tokens: DEVICE_DOMAIN_NAME, APP_SEED
- `wireguard` — 
- `zigbee2mqtt` — 

## ✗ Failed
