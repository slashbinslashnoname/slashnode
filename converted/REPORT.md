# Umbrel → SlashNode conversion report

Converted 379 apps · ✓ 176 clean · △ 203 need review · ✗ 0 failed · 🔒 81 with security flags.

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
### `adventurelog`
- ⚠️ server: declares ports ["8016:80"] — left UNpublished, add by hand if required
### `anything-llm`
- ⚠️ app: extra capabilities ["SYS_ADMIN"]
### `arcane`
- ⚠️ ships a default password ("arcane-admin") — must be changed on first login
- ⚠️ docker: runs privileged
- ⚠️ docker: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
### `back-that-mac-up`
- ⚠️ timemachine: declares ports ["137:137/udp","138:138/udp","139:139","445:445"] — left UNpublished, add by hand if required
### `bassin`
- ⚠️ ckpool: declares ports ["3456:3333/tcp"] — left UNpublished, add by hand if required
### `bitcoin-regtest-dashboard`
- ⚠️ electrs: declares ports ["60401:50001"] — left UNpublished, add by hand if required
### `blockstream-blind-oracle`
- ⚠️ node: declares ports ["$APP_PINSERVER_PORT:8096"] — left UNpublished, add by hand if required
### `bookstack`
- ⚠️ ships a default password ("password") — must be changed on first login
### `calibre-web`
- ⚠️ ships a default password ("admin123") — must be changed on first login
### `copyparty`
- ⚠️ ships a default password ("umbrel") — must be changed on first login
### `core-lightning`
- ⚠️ lightningd: declares ports ["${APP_CORE_LIGHTNING_DAEMON_PORT}:9735","${APP_CORE_LIGHTNING_WEBSOCKET_PORT}:${APP_CORE_LIGHTNING_WEBSOCKET_PORT}","${CORE_LIGHTNING_REST_PORT}:${CORE_LIGHTNING_REST_PORT}","${APP_CORE_LIGHTNING_DAEMON_GRPC_PORT}:${APP_CORE_LIGHTNING_DAEMON_GRPC_PORT}"] — left UNpublished, add by hand if required
### `dockge`
- ⚠️ docker: runs privileged
- ⚠️ docker: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
### `ee-gateway`
- ⚠️ worker: extra capabilities ["NET_ADMIN","NET_RAW"]
- ⚠️ worker: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
### `electrs`
- ⚠️ electrs: declares ports ["${APP_ELECTRS_NODE_PORT}:${APP_ELECTRS_NODE_PORT}"] — left UNpublished, add by hand if required
### `electrumx`
- ⚠️ electrumx: declares ports ["${APP_ELECTRUMX_PUBLIC_CONNECTION_PORT}:${APP_ELECTRUMX_NODE_PORT}"] — left UNpublished, add by hand if required
### `elements`
- ⚠️ node: declares ports ["$APP_ELEMENTS_NODE_RPC_PORT:$APP_ELEMENTS_NODE_RPC_PORT","$APP_ELEMENTS_NODE_P2P_PORT:$APP_ELEMENTS_NODE_P2P_PORT"] — left UNpublished, add by hand if required
### `endurain`
- ⚠️ ships a default password ("admin") — must be changed on first login
### `esphome`
- ⚠️ server: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
### `formicaio`
- ⚠️ app: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
### `frigate`
- ⚠️ web: runs privileged
### `fulcrum`
- ⚠️ fulcrum: declares ports ["${APP_FULCRUM_NODE_PORT}:${APP_FULCRUM_NODE_PORT}"] — left UNpublished, add by hand if required
### `gobrrr-pool`
- ⚠️ ckpool: declares ports ["21420:3333","21422:4444"] — left UNpublished, add by hand if required
### `grafana`
- ⚠️ ships a default password ("admin") — must be changed on first login
### `grocy`
- ⚠️ ships a default password ("admin") — must be changed on first login
### `home-assistant`
- ⚠️ server: runs privileged
- ⚠️ server: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
### `homebridge`
- ⚠️ server: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
### `homey`
- ⚠️ web: runs privileged
- ⚠️ web: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
### `kimai`
- ⚠️ ships a default password ("changeme") — must be changed on first login
### `kollider`
- ⚠️ ws: declares ports ["4244:8080"] — left UNpublished, add by hand if required
### `komodo`
- ⚠️ docker: runs privileged
- ⚠️ docker: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
### `libre-relay`
- ⚠️ bitcoind: declares ports ["${APP_LIBRE_RELAY_P2P_PORT}:${APP_LIBRE_RELAY_P2P_PORT}","${APP_LIBRE_RELAY_RPC_PORT}:${APP_LIBRE_RELAY_RPC_PORT}"] — left UNpublished, add by hand if required
### `lightning`
- ⚠️ lnd: declares ports ["$APP_LIGHTNING_NODE_PORT:$APP_LIGHTNING_NODE_PORT","$APP_LIGHTNING_NODE_REST_PORT:$APP_LIGHTNING_NODE_REST_PORT","$APP_LIGHTNING_NODE_GRPC_PORT:$APP_LIGHTNING_NODE_GRPC_PORT"] — left UNpublished, add by hand if required
### `lobe-chat`
- ⚠️ rustfs: declares ports ["7458:9000"] — left UNpublished, add by hand if required
### `mailarchiver`
- ⚠️ ships a default password ("secure123!") — must be changed on first login
### `matter-server`
- ⚠️ server: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
### `monero`
- ⚠️ monerod: declares ports ["${APP_MONERO_P2P_PORT}:${APP_MONERO_P2P_PORT}","${APP_MONERO_RPC_PORT}:${APP_MONERO_RPC_PORT}"] — left UNpublished, add by hand if required
### `mosquitto`
- ⚠️ broker: declares ports ["1883:1883"] — left UNpublished, add by hand if required
### `music-assistant`
- ⚠️ web: extra capabilities ["SYS_ADMIN","DAC_READ_SEARCH"]
- ⚠️ web: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
### `netbird`
- ⚠️ app: extra capabilities ["NET_ADMIN","SYS_ADMIN","SYS_RESOURCE"]
- ⚠️ app: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
### `nginx-proxy-manager`
- ⚠️ docker-host: extra capabilities ["NET_ADMIN","NET_RAW"]
### `node-red`
- ⚠️ ships a default password ("moneyprintergobrrr") — must be changed on first login
### `node-red-standalone`
- ⚠️ web: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
### `nostr-vpn`
- ⚠️ daemon: extra capabilities ["NET_ADMIN"]
- ⚠️ daemon: maps host devices ["/dev/net/tun:/dev/net/tun"]
- ⚠️ daemon: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
### `octoprint`
- ⚠️ web: runs privileged
### `onlyoffice-nextcloud`
- ⚠️ documentserver: declares ports ["${DOCSERVER_PORT}:80"] — left UNpublished, add by hand if required
### `openhands`
- ⚠️ docker: runs privileged
- ⚠️ docker: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
### `openresty-manager`
- ⚠️ ships a default password ("#Passw0rd") — must be changed on first login
### `openthread-border-router`
- ⚠️ setup: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
- ⚠️ server: runs privileged
- ⚠️ server: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
### `outline`
- ⚠️ dex: declares ports ["8943:5556"] — left UNpublished, add by hand if required
### `pi-hole`
- ⚠️ server: extra capabilities ["NET_ADMIN"]
- ⚠️ server: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
### `plane`
- ⚠️ minio: declares ports ["8763:8763"] — left UNpublished, add by hand if required
### `plex`
- ⚠️ server: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
### `pocketbase`
- ⚠️ ships a default password ("umbrel-pocketbase") — must be changed on first login
### `pogolo`
- ⚠️ pogolo: declares ports ["5661:5661","5662:5662"] — left UNpublished, add by hand if required
### `portainer`
- ⚠️ ships a default password ("changeme") — must be changed on first login
- ⚠️ docker: runs privileged
- ⚠️ docker: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
### `poznote`
- ⚠️ ships a default password ("admin") — must be changed on first login
- ⚠️ mcp: declares ports ["8340:8045"] — left UNpublished, add by hand if required
### `public-pool`
- ⚠️ server: declares ports ["2018:2018/tcp"] — left UNpublished, add by hand if required
### `pyload-ng`
- ⚠️ ships a default password ("pyload") — must be changed on first login
### `qbittorrent`
- ⚠️ ships a default password ("adminadmin") — must be changed on first login
### `readur`
- ⚠️ ships a default password ("readur2024") — must be changed on first login
### `rustdesk-server`
- ⚠️ hbbs: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
- ⚠️ hbbs: declares ports ["21115:21115","21116:21116","21116:21116/udp","21118:21118"] — left UNpublished, add by hand if required
- ⚠️ hbbr: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
- ⚠️ hbbr: declares ports ["21117:21117","21119:21119"] — left UNpublished, add by hand if required
### `rusty-kaspad`
- ⚠️ kaspad: declares ports ["16110:16110","16111:16111","17110:17110","18110:18110"] — left UNpublished, add by hand if required
### `saifa`
- ⚠️ backend: declares ports ["9988:80"] — left UNpublished, add by hand if required
### `samba`
- ⚠️ server: declares ports ["446:445"] — left UNpublished, add by hand if required
### `samourai-server`
- ⚠️ nginx: declares ports ["$APP_SAMOURAI_SERVER_DOJO_PORT:80"] — left UNpublished, add by hand if required
### `scanservjs`
- ⚠️ server: runs privileged
### `seafile`
- ⚠️ seadoc: declares ports ["8921:80"] — left UNpublished, add by hand if required
### `suredbits-wallet`
- ⚠️ walletserver: declares ports ["$APP_SUREDBITS_WALLET_P2P_PORT:$APP_SUREDBITS_WALLET_P2P_PORT"] — left UNpublished, add by hand if required
### `sv2-ui`
- ⚠️ docker: runs privileged
- ⚠️ docker: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
### `syslog-ng`
- ⚠️ syslog: declares ports ["514:5514/udp","601:6601/tcp"] — left UNpublished, add by hand if required
### `tailscale`
- ⚠️ web: extra capabilities ["NET_ADMIN","NET_RAW"]
- ⚠️ web: maps host devices ["/dev/net/tun:/dev/net/tun"]
- ⚠️ web: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
### `tdex`
- ⚠️ tdexd: declares ports ["${APP_TDEX_PORT}:${APP_TDEX_PORT}"] — left UNpublished, add by hand if required
### `teamspeak`
- ⚠️ server: declares ports ["9987:9987/udp","10011:10011","30033:30033"] — left UNpublished, add by hand if required
### `technitium-dns`
- ⚠️ dns-server: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
### `umami`
- ⚠️ ships a default password ("umami") — must be changed on first login
### `watch-your-lan`
- ⚠️ web: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
### `wireguard`
- ⚠️ app: extra capabilities ["NET_ADMIN","SYS_MODULE"]
### `zabbix`
- ⚠️ ships a default password ("zabbix") — must be changed on first login
- ⚠️ zabbix-server: declares ports ["10050:10050"] — left UNpublished, add by hand if required
### `zerotier`
- ⚠️ zerotier: extra capabilities ["NET_ADMIN"]
- ⚠️ zerotier: maps host devices ["/dev/net/tun"]
- ⚠️ zerotier: host network mode — bypasses Caddy/Tor isolation, binds host interfaces directly
### `zigbee2mqtt`
- ⚠️ app: runs privileged

## ✓ Clean
- `affine`
- `airtrail`
- `akaunting`
- `appsmith`
- `archivebox`
- `autobrr`
- `baikal`
- `bark-wallet`
- `bentopdf`
- `bffless`
- `bitaxe-sentry`
- `bitbalance`
- `bitboard`
- `bitfeed`
- `bitmagnet`
- `bitwatch`
- `booklore`
- `btc-rpc-explorer`
- `budibase`
- `campfire`
- `canary`
- `cashu-me`
- `chainforensics`
- `chatbot-ui`
- `chatpad-ai`
- `chromium`
- `code-server`
- `convertx`
- `databag`
- `datum`
- `dcrdex`
- `deepsea`
- `docuseal`
- `domain-locker`
- `donetick`
- `dropgate-server`
- `dtan-server`
- `dumbpad`
- `element`
- `enclosed`
- `etherpad`
- `excalidraw`
- `file-drop`
- `firefly-iii`
- `firefox`
- `fizzy`
- `flaresolverr`
- `flatnotes`
- `fossflow`
- `freshrss`
- `gitlab`
- `gupt`
- `heimdall`
- `hermes-agent`
- `holesail-switchboard`
- `homarr`
- `homebox`
- `homehub`
- `hortusfox`
- `immich`
- `influxdb`
- `influxdb2`
- `invidious`
- `ipfs-podcasting`
- `itchysats`
- `ittools`
- `jellyseerr`
- `jotty`
- `just-download`
- `kitchenowl`
- `kokoro`
- `l-town`
- `langflow`
- `librechat`
- `libreddit`
- `libreoffice`
- `librephotos`
- `librespeed`
- `libretranslate`
- `linkstack`
- `llama-gpt`
- `lnbits-holesail-proxy`
- `localai`
- `lubelogger`
- `lunalytics`
- `mainsail`
- `mattermost`
- `maybe`
- `mazanoke`
- `mealie`
- `memos`
- `meshchatx`
- `minio`
- `monetr`
- `morphos`
- `mqttx-web`
- `myspeed`
- `n8n`
- `neko`
- `networkingtoolbox`
- `nitter`
- `nocodb`
- `nostr-relay`
- `nostrudel`
- `notediscovery`
- `nutstash-wallet`
- `obsidian`
- `ollama`
- `omnitools`
- `open-webui`
- `openclaw`
- `opencode`
- `originless`
- `overseerr`
- `palmr`
- `passky-client`
- `passky-server`
- `pearcircle-seeder`
- `perplexica`
- `picoclaw`
- `picsur`
- `pingvin-share`
- `privatebin`
- `readeck`
- `reitti`
- `remmina`
- `restreamer`
- `romm`
- `rotki`
- `satoshi-dashboard`
- `satwatch`
- `searxng`
- `shopstr`
- `sikka`
- `snapdrop`
- `snort`
- `snowflake`
- `spacebot`
- `specter-desktop`
- `sqlitebrowser`
- `stash`
- `stirling-pdf`
- `strix`
- `super-productivity`
- `sure`
- `syncthing`
- `tautulli`
- `telegrapho`
- `termix`
- `thelounge`
- `thinkdashboard`
- `threema`
- `torbrowser`
- `toshi-moto`
- `trilium-notes`
- `trip`
- `uptime-kuma`
- `urbit`
- `urbit-bitcoin-connector`
- `vaultwarden`
- `wallos`
- `wavelog`
- `wealthfolio`
- `webcheck`
- `whoogle-search`
- `wikijs`
- `wingfit`
- `wizarr`
- `woofbot`
- `wordpress`
- `yamtrack`
- `yucca`
- `yuvomi`
- `zen`
- `zeronote`
- `zoraxy`

## △ Needs review
- `activepieces` — unmapped tokens: DEVICE_DOMAIN_NAME
- `adguard-home` — 
- `adventurelog` — unmapped tokens: DEVICE_DOMAIN_NAME
- `agent-zero` — unmapped tokens: DEVICE_DOMAIN_NAME, APP_AGENTZERO_PORT, APP_AGENTZERO_LOCAL_URLS
- `agora` — unmapped tokens: ${APP_LIGHTNING_NODE_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR
- `alby-nostr-wallet-connect` — unmapped tokens: ${APP_LIGHTNING_NODE_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR
- `albyhub` — unmapped tokens: APP_ALBYHUB_LND_ADDRESS, APP_ALBYHUB_LND_CERT_FILE, APP_ALBYHUB_LND_MACAROON_FILE; shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `am-i-exposed` — unmapped tokens: APP_MEMPOOL_IP, APP_MEMPOOL_PORT, APP_MEMPOOL_HIDDEN_SERVICE, TOR_PROXY_IP, TOR_PROXY_PORT
- `anything-llm` — unmapped tokens: APP_ANYTHING_LLM_SIG_SALT
- `arcane` — unmapped tokens: DEVICE_DOMAIN_NAME
- `audiobookshelf` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `back-that-mac-up` — 
- `bassin` — unmapped tokens: APP_BITCOIN_NODE_IP, APP_BITCOIN_RPC_PORT, APP_BITCOIN_RPC_USER, APP_BITCOIN_RPC_PASS, APP_BITCOIN_ZMQ_HASHBLOCK_PORT
- `bazarr` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `bitcoin` — unmapped tokens: NETWORK_IP, APP_BITCOIN_P2P_PORT, APP_BITCOIN_P2P_WHITEBIND_PORT, APP_BITCOIN_RPC_PORT, APP_BITCOIN_TOR_PORT, APP_BITCOIN_ZMQ_RAWBLOCK_PORT, APP_BITCOIN_ZMQ_RAWTX_PORT, APP_BITCOIN_ZMQ_HASHBLOCK_PORT, APP_BITCOIN_ZMQ_SEQUENCE_PORT, APP_BITCOIN_ZMQ_HASHTX_PORT, APP_BITCOIN_TOR_PROXY_IP, APP_BITCOIN_I2P_DAEMON_IP, APP_BITCOIN_NODE_IP, APP_BITCOIN_RPC_USER, APP_BITCOIN_RPC_PASS, DEVICE_DOMAIN_NAME, APP_BITCOIN_P2P_HIDDEN_SERVICE, APP_BITCOIN_RPC_HIDDEN_SERVICE, ${TOR_DATA_DIR}, TOR_DATA_DIR
- `bitcoin-knots` — unmapped tokens: NETWORK_IP, APP_BITCOIN_KNOTS_P2P_PORT, APP_BITCOIN_KNOTS_P2P_WHITEBIND_PORT, APP_BITCOIN_KNOTS_RPC_PORT, APP_BITCOIN_KNOTS_TOR_PORT, APP_BITCOIN_KNOTS_ZMQ_RAWBLOCK_PORT, APP_BITCOIN_KNOTS_ZMQ_RAWTX_PORT, APP_BITCOIN_KNOTS_ZMQ_HASHBLOCK_PORT, APP_BITCOIN_KNOTS_ZMQ_SEQUENCE_PORT, APP_BITCOIN_KNOTS_ZMQ_HASHTX_PORT, APP_BITCOIN_KNOTS_TOR_PROXY_IP, APP_BITCOIN_KNOTS_I2P_DAEMON_IP, APP_BITCOIN_KNOTS_NODE_IP, APP_BITCOIN_KNOTS_RPC_USER, APP_BITCOIN_KNOTS_RPC_PASS, DEVICE_DOMAIN_NAME, APP_BITCOIN_KNOTS_P2P_HIDDEN_SERVICE, APP_BITCOIN_KNOTS_RPC_HIDDEN_SERVICE, ${TOR_DATA_DIR}, TOR_DATA_DIR
- `bitcoin-regtest-dashboard` — 
- `bleskomat-server` — unmapped tokens: $APP_DATA_DIR/data/db, $APP_DATA_DIR/data/web, $APP_LIGHTNING_NODE_DATA_DIR
- `blinko` — unmapped tokens: DEVICE_DOMAIN_NAME
- `blockstream-blind-oracle` — unmapped tokens: APP_PINSERVER_HIDDEN_SERVICE, APP_PINSERVER_PORT, APP_TAILSCALE_URL, ${TOR_DATA_DIR}, TOR_DATA_DIR
- `bluewallet` — unmapped tokens: APP_HIDDEN_SERVICE, APP_BITCOIN_NETWORK, ${APP_LIGHTNING_NODE_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR
- `bolt12-pay` — unmapped tokens: APP_BITCOIN_NETWORK, ${APP_LIGHTNING_NODE_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR
- `bookstack` — 
- `btcpay-server` — unmapped tokens: APP_BITCOIN_P2P_PORT, ${APP_LIGHTNING_NODE_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR
- `btctracker` — unmapped tokens: DEVICE_DOMAIN_NAME
- `calibre-web` — 
- `changedetection-io` — unmapped tokens: DEVICE_DOMAIN_NAME
- `chantools` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `circuitbreaker` — unmapped tokens: ${APP_LIGHTNING_NODE_DATA_DIR}, APP_BITCOIN_NETWORK, APP_LIGHTNING_NODE_IP, APP_LIGHTNING_NODE_GRPC_PORT, APP_LIGHTNING_NODE_DATA_DIR
- `cloudflared` — unmapped tokens: APP_CLOUDFLARED_METRICS_PORT
- `cobalt` — unmapped tokens: DEVICE_DOMAIN_NAME
- `copyparty` — 
- `core-lightning` — unmapped tokens: APP_CORE_LIGHTNING_IP, APP_CORE_LIGHTNING_PORT, APP_CORE_LIGHTNING_BITCOIN_NETWORK, APP_CORE_LIGHTNING_DAEMON_IP, APP_CORE_LIGHTNING_HIDDEN_SERVICE, APP_MODE, CORE_LIGHTNING_PATH, APP_CONFIG_DIR, COMMANDO_CONFIG, APP_CORE_LIGHTNING_WEBSOCKET_PORT, DEVICE_DOMAIN_NAME, CORE_LIGHTNING_REST_PORT, APP_CORE_LIGHTNING_DAEMON_GRPC_PORT, ${APP_CORE_LIGHTNING_DATA_DIR}, ${TOR_DATA_DIR}, APP_CORE_LIGHTNING_DATA_DIR, APP_BITCOIN_NODE_IP, APP_BITCOIN_RPC_USER, APP_BITCOIN_RPC_PASS, APP_BITCOIN_RPC_PORT, TOR_PROXY_IP, TOR_PROXY_PORT, TOR_PASSWORD, TOR_DATA_DIR
- `core-lightning-rtl` — unmapped tokens: APP_CORE_RTL_BLOCK_EXPLORER_URL, ${APP_CORE_LIGHTNING_DATA_DIR}, APP_CORE_LIGHTNING_DATA_DIR
- `dcrpulse` — unmapped tokens: TOR_PROXY_IP, TOR_PROXY_PORT, APP_SEED
- `dockge` — 
- `docmost` — unmapped tokens: DEVICE_DOMAIN_NAME
- `downtify` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `duplicati` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `ee-gateway` — 
- `electrs` — unmapped tokens: APP_ELECTRS_RPC_HIDDEN_SERVICE, DEVICE_DOMAIN_NAME, APP_ELECTRS_NODE_IP, APP_BITCOIN_NODE_IP, APP_BITCOIN_RPC_USER, APP_BITCOIN_RPC_PASS, APP_BITCOIN_RPC_PORT, APP_BITCOIN_NETWORK_ELECTRS, APP_BITCOIN_P2P_PORT, APP_ELECTRS_NODE_PORT, APP_VERSION, ${APP_BITCOIN_DATA_DIR}, ${TOR_DATA_DIR}, APP_BITCOIN_DATA_DIR, TOR_DATA_DIR
- `electrumx` — unmapped tokens: APP_ELECTRUMX_RPC_HIDDEN_SERVICE, DEVICE_DOMAIN_NAME, APP_ELECTRUMX_NODE_IP, APP_ELECTRUMX_PUBLIC_CONNECTION_PORT, APP_ELECTRUMX_RPC_PORT, APP_BITCOIN_NODE_IP, APP_BITCOIN_RPC_USER, APP_BITCOIN_RPC_PASS, APP_BITCOIN_RPC_PORT, ${TOR_DATA_DIR}, TOR_DATA_DIR
- `elements` — unmapped tokens: ${TOR_DATA_DIR}, TOR_DATA_DIR
- `emby` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `endurain` — 
- `ersatztv-legacy` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `esphome` — 
- `fedimint-gateway` — unmapped tokens: APP_BITCOIN_NETWORK_ELECTRS, APP_LIGHTNING_NODE_IP, APP_LIGHTNING_NODE_GRPC_PORT, APP_BITCOIN_NETWORK, APP_BITCOIN_RPC_USER, APP_BITCOIN_RPC_PASS, APP_BITCOIN_NODE_IP, APP_BITCOIN_RPC_PORT, ${APP_LIGHTNING_NODE_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR
- `fedimintd` — unmapped tokens: APP_BITCOIN_NETWORK_ELECTRS, APP_BITCOIN_RPC_USER, APP_BITCOIN_RPC_PASS, APP_BITCOIN_NODE_IP, APP_BITCOIN_RPC_PORT
- `file-browser` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `firefly-iii-importer` — unmapped tokens: DEVICE_DOMAIN_NAME
- `forgejo` — unmapped tokens: APP_DOMAIN, APP_FORGEJO_SSH_PORT
- `formicaio` — 
- `frigate` — 
- `fulcrum` — unmapped tokens: APP_FULCRUM_RPC_HIDDEN_SERVICE, DEVICE_DOMAIN_NAME, APP_FULCRUM_NODE_IP, APP_FULCRUM_NODE_PORT, APP_BITCOIN_NODE_IP, APP_BITCOIN_RPC_USER, APP_BITCOIN_RPC_PASS, APP_BITCOIN_RPC_PORT, APP_FULCRUM_ADMIN_PORT, ${TOR_DATA_DIR}, TOR_DATA_DIR
- `ghost` — unmapped tokens: DEVICE_DOMAIN_NAME
- `ghostfolio` — unmapped tokens: APP_GHOSTFOLIO_DB_USERNAME, APP_GHOSTFOLIO_DB_DATABASE_NAME, APP_GHOSTFOLIO_REDIS_PASSWORD
- `gitea` — unmapped tokens: APP_DOMAIN, APP_GITEA_SSH_PORT
- `gitea-mirror` — unmapped tokens: DEVICE_DOMAIN_NAME, APP_GITEA_MIRROR_PORT
- `gitingest` — unmapped tokens: DEVICE_DOMAIN_NAME, APP_GITINGEST_LOCAL_IPS
- `gobrrr-pool` — unmapped tokens: APP_BITCOIN_NODE_IP, APP_BITCOIN_RPC_PORT, APP_BITCOIN_RPC_USER, APP_BITCOIN_RPC_PASS, APP_BITCOIN_ZMQ_HASHBLOCK_PORT, APP_MEMPOOL_IP, APP_MEMPOOL_PORT
- `grafana` — 
- `grocy` — 
- `habitica` — unmapped tokens: DEVICE_DOMAIN_NAME
- `hashrate-autopilot` — unmapped tokens: APP_BITCOIN_NODE_IP, APP_BITCOIN_RPC_PORT, APP_BITCOIN_RPC_USER, APP_BITCOIN_RPC_PASS
- `helipad` — unmapped tokens: ${APP_LIGHTNING_NODE_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR
- `hermitstash` — unmapped tokens: DEVICE_DOMAIN_NAME
- `home-assistant` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `home-assistant-fusion-ui` — unmapped tokens: DEVICE_DOMAIN_NAME
- `homebridge` — 
- `homey` — 
- `invio` — unmapped tokens: DEVICE_DOMAIN_NAME
- `invoice-ninja` — unmapped tokens: APP_DOMAIN
- `jackett` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `jam` — unmapped tokens: APP_BITCOIN_RPC_PASS
- `jellyfin` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `jupyterlab` — unmapped tokens: APP_PASSWORD
- `kan` — unmapped tokens: DEVICE_DOMAIN_NAME
- `karakeep` — unmapped tokens: DEVICE_DOMAIN_NAME
- `kimai` — 
- `kiwix` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `kollider` — unmapped tokens: ${APP_LIGHTNING_NODE_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR
- `komodo` — 
- `krystal-bull` — unmapped tokens: TOR_PROXY_IP, TOR_PROXY_PORT
- `libre-relay` — unmapped tokens: APP_LIBRE_RELAY_NODE_IP, APP_LIBRE_RELAY_RPC_PORT, APP_LIBRE_RELAY_RPC_USER, APP_LIBRE_RELAY_RPC_PASS, APP_LIBRE_RELAY_RPC_HIDDEN_SERVICE, APP_LIBRE_RELAY_P2P_HIDDEN_SERVICE, APP_LIBRE_RELAY_P2P_PORT, DEVICE_DOMAIN_NAME, APP_LIBRE_RELAY_TOR_PROXY_IP, APP_LIBRE_RELAY_I2P_DAEMON_IP, ${APP_LIBRE_RELAY_DATA_DIR}, ${TOR_DATA_DIR}, APP_LIBRE_RELAY_DATA_DIR, APP_LIBRE_RELAY_COMMAND, TOR_DATA_DIR
- `lidarr` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `lightning` — unmapped tokens: TOR_PROXY_IP, TOR_PROXY_PORT, APP_BITCOIN_NODE_IP, APP_BITCOIN_RPC_PORT, APP_BITCOIN_RPC_USER, APP_BITCOIN_RPC_PASS, APP_BITCOIN_NETWORK, APP_LIGHTNING_NODE_IP, APP_LIGHTNING_REST_HIDDEN_SERVICE, APP_LIGHTNING_GRPC_HIDDEN_SERVICE, DEVICE_DOMAIN_NAME, APP_MEMPOOL_PORT, APP_MEMPOOL_HIDDEN_SERVICE, ${APP_LIGHTNING_NODE_DATA_DIR}, ${TOR_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR, APP_LIGHTNING_COMMAND, TOR_DATA_DIR
- `lightning-shell` — unmapped tokens: APP_BITCOIN_NODE_IP, APP_BITCOIN_RPC_PORT, APP_BITCOIN_RPC_USER, APP_BITCOIN_RPC_PASS, APP_LIGHTNING_NODE_IP, APP_LIGHTNING_NODE_GRPC_PORT, ${APP_LIGHTNING_NODE_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR
- `lightning-terminal` — unmapped tokens: ${APP_LIGHTNING_NODE_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR
- `lightningmate` — unmapped tokens: ${APP_LIGHTNING_NODE_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR
- `linkwarden` — unmapped tokens: DEVICE_DOMAIN_NAME
- `ln-visualizer` — unmapped tokens: APP_BITCOIN_NETWORK, APP_LIGHTNING_NODE_IP, APP_LIGHTNING_NODE_GRPC_PORT, ${APP_LIGHTNING_NODE_DATA_DIR}/tls.cert, ${APP_LIGHTNING_NODE_DATA_DIR}/data/chain/bitcoin/${APP_BITCOIN_NETWORK}/readonly.macaroon, APP_LIGHTNING_NODE_DATA_DIR
- `lnbits` — unmapped tokens: ${APP_LIGHTNING_NODE_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR
- `lndboss` — unmapped tokens: ${APP_LIGHTNING_NODE_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR
- `lndg` — unmapped tokens: ${APP_LIGHTNING_NODE_DATA_DIR}, APP_BITCOIN_NETWORK, APP_LIGHTNING_NODE_IP, APP_LIGHTNING_NODE_GRPC_PORT, APP_PASSWORD, APP_LNDG_PORT, APP_LIGHTNING_NODE_DATA_DIR
- `lnmarkets` — unmapped tokens: ${APP_LIGHTNING_NODE_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR
- `lnplus` — unmapped tokens: ${APP_LIGHTNING_NODE_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR
- `lnswitchboard` — unmapped tokens: ${APP_LIGHTNING_NODE_DATA_DIR}/data/chain/bitcoin/${APP_BITCOIN_NETWORK}/invoice.macaroon, ${APP_LIGHTNING_NODE_DATA_DIR}/data/chain/bitcoin/${APP_BITCOIN_NETWORK}/readonly.macaroon, ${APP_LIGHTNING_NODE_DATA_DIR}/tls.cert, APP_LIGHTNING_NODE_DATA_DIR, APP_BITCOIN_NETWORK
- `lobe-chat` — unmapped tokens: DEVICE_DOMAIN_NAME, ./bucket.config.json, APP_PASSWORD
- `mailarchiver` — 
- `mailflow` — unmapped tokens: DEVICE_DOMAIN_NAME
- `matter-server` — 
- `mempool` — unmapped tokens: max, file, size; shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `metube` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `miner-sentinel` — unmapped tokens: DEVICE_DOMAIN_NAME
- `monero` — unmapped tokens: APP_MONERO_NODE_IP, APP_MONERO_P2P_PORT, APP_MONERO_RPC_PORT, APP_MONERO_RESTRICTED_RPC_PORT, MONERO_DEFAULT_NETWORK, APP_MONERO_RPC_USER, APP_MONERO_RPC_PASS, APP_MONERO_RPC_HIDDEN_SERVICE, APP_MONERO_P2P_HIDDEN_SERVICE, DEVICE_DOMAIN_NAME, APP_MONERO_TOR_PROXY_IP, APP_MONERO_I2P_DAEMON_IP, ${APP_MONERO_DATA_DIR}, ${TOR_DATA_DIR}, APP_MONERO_DATA_DIR, APP_MONERO_COMMAND, TOR_DATA_DIR
- `mosquitto` — 
- `mstream` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `music-assistant` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `navidrome` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `netbird` — 
- `nextcloud` — unmapped tokens: NETWORK_IP, APP_DOMAIN, APP_NEXTCLOUD_PORT, APP_HIDDEN_SERVICE, APP_NEXTCLOUD_LOCAL_IPS
- `nginx-proxy-manager` — 
- `node-red` — unmapped tokens: ${APP_LIGHTNING_NODE_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR
- `node-red-standalone` — 
- `nolooking` — unmapped tokens: APP_BITCOIN_NETWORK, ${APP_LIGHTNING_NODE_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR
- `nostr-vpn` — 
- `ntfy` — unmapped tokens: DEVICE_DOMAIN_NAME, APP_PROXY_PORT
- `oak-node` — unmapped tokens: $APP_LIGHTNING_NODE_DATA_DIR
- `octoprint` — 
- `onlyoffice-nextcloud` — 
- `openhands` — 
- `openreader` — unmapped tokens: DEVICE_DOMAIN_NAME, APP_OPENREADER_LOCAL_URLS
- `openresty-manager` — 
- `openthread-border-router` — 
- `ordinals` — unmapped tokens: APP_BITCOIN_RPC_USER, APP_BITCOIN_RPC_PASS, APP_BITCOIN_NODE_IP, APP_BITCOIN_RPC_PORT, APP_BITCOIN_NETWORK, ${APP_BITCOIN_DATA_DIR}, APP_BITCOIN_DATA_DIR
- `outline` — unmapped tokens: DEVICE_DOMAIN_NAME, APP_OUTLINE_PORT
- `owncloud` — unmapped tokens: DEVICE_DOMAIN_NAME, APP_OWNCLOUD_LOCAL_IPS
- `paperclip` — unmapped tokens: DEVICE_DOMAIN_NAME, APP_DOMAIN, APP_HIDDEN_SERVICE, APP_PAPERCLIP_LOCAL_IPS
- `paperless` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `papra` — unmapped tokens: DEVICE_DOMAIN_NAME
- `pastefy` — unmapped tokens: DEVICE_DOMAIN_NAME
- `peerswap` — unmapped tokens: APP_LIGHTNING_NODE_IP, APP_LIGHTNING_NODE_GRPC_PORT, APP_BITCOIN_NETWORK, ${APP_LIGHTNING_NODE_DATA_DIR}, ${ELEMENTS_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR, ELEMENTS_DATA_DIR
- `penpot` — unmapped tokens: DEVICE_DOMAIN_NAME, APP_PENPOT_UI_PORT
- `photoprism` — unmapped tokens: APP_DOMAIN
- `pi-hole` — 
- `pinchflat` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `plane` — 
- `planka` — unmapped tokens: DEVICE_DOMAIN_NAME
- `plausible` — unmapped tokens: DEVICE_DOMAIN_NAME
- `plex` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `pocketbase` — 
- `pogolo` — 
- `portainer` — 
- `poznote` — 
- `prowlarr` — unmapped tokens: APP_PROWLARR_RADARR_CONFIG_XML, APP_PROWLARR_LIDARR_CONFIG_XML, APP_PROWLARR_SONARR_CONFIG_XML, APP_PROWLARR_READARR_CONFIG_XML; shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `public-pool` — unmapped tokens: DEVICE_DOMAIN_NAME, APP_BITCOIN_NODE_IP, APP_BITCOIN_RPC_USER, APP_BITCOIN_RPC_PASS, APP_BITCOIN_RPC_PORT
- `public-pool-web` — unmapped tokens: APP_PUBLIC_POOL_WEB_DATABASE_URL, APP_PUBLIC_POOL_WEB_REDIS_URL, APP_PUBLIC_POOL_WEB_POSTGRES_USERNAME, APP_PUBLIC_POOL_WEB_POSTGRES_DBNAME
- `pyload-ng` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `qbittorrent` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `radarr` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `readarr` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `readur` — 
- `ride-the-lightning` — unmapped tokens: APP_RTL_BLOCK_EXPLORER_URL, ${APP_LIGHTNING_NODE_DATA_DIR}, ${APP_BITCOIN_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR, APP_BITCOIN_DATA_DIR
- `robosats` — unmapped tokens: TOR_PROXY_IP, TOR_PROXY_PORT
- `route96` — unmapped tokens: APP_DOMAIN
- `rustdesk-server` — 
- `rusty-kaspad` — 
- `sabnzbd` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `saifa` — unmapped tokens: ${APP_LIGHTNING_NODE_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR
- `samba` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `samourai-server` — unmapped tokens: ${TOR_DATA_DIR}, APP_SAMOURAI_SERVER_DB_IP, TOR_DATA_DIR
- `satsale` — unmapped tokens: APP_BITCOIN_NETWORK, ${APP_LIGHTNING_NODE_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR
- `satsbook` — unmapped tokens: APP_LIGHTNING_NODE_IP, APP_LIGHTNING_NODE_GRPC_PORT, APP_BITCOIN_NETWORK, ${APP_LIGHTNING_NODE_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR
- `scanservjs` — 
- `seafile` — unmapped tokens: DEVICE_DOMAIN_NAME
- `simple-torrent` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `slink` — unmapped tokens: DEVICE_DOMAIN_NAME
- `solidtime` — unmapped tokens: DEVICE_DOMAIN_NAME
- `sonarr` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `sparkkiosk` — unmapped tokens: ${APP_LIGHTNING_NODE_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR
- `sphinx-relay` — unmapped tokens: APP_BITCOIN_NETWORK, ${APP_LIGHTNING_NODE_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR
- `squeaknode` — unmapped tokens: TOR_PROXY_IP, TOR_PROXY_PORT, ${APP_LIGHTNING_NODE_DATA_DIR}, ${TOR_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR, TOR_DATA_DIR
- `squeakroad` — unmapped tokens: $APP_LIGHTNING_NODE_DATA_DIR
- `stalwart` — unmapped tokens: DEVICE_DOMAIN_NAME
- `suredbits-wallet` — unmapped tokens: TOR_PROXY_IP, TOR_PROXY_PORT, ${TOR_DATA_DIR}, TOR_DATA_DIR
- `sv2-ui` — unmapped tokens: ${APP_BITCOIN_DATA_DIR}, DEVICE_DOMAIN_NAME, APP_BITCOIN_DATA_DIR
- `swingmusic` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `synapse` — unmapped tokens: APP_SYNAPSE_PORT, APP_HIDDEN_SERVICE
- `syslog-ng` — unmapped tokens: TZ
- `tailscale` — 
- `tallycoin-connect` — unmapped tokens: ${APP_LIGHTNING_NODE_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR
- `tandoor` — unmapped tokens: DEVICE_DOMAIN_NAME, APP_TANDOOR_PORT, APP_TANDOOR_LOCAL_URLS
- `tdex` — unmapped tokens: APP_TDEX_PORT, APP_TDEX_DAEMON_HIDDEN_SERVICE, ${TOR_DATA_DIR}, TOR_DATA_DIR
- `teamspeak` — 
- `technitium-dns` — unmapped tokens: DEVICE_DOMAIN_NAME
- `thunderhub` — unmapped tokens: ${APP_LIGHTNING_NODE_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR
- `torq` — unmapped tokens: ${APP_LIGHTNING_NODE_DATA_DIR}, APP_PASSWORD, APP_LIGHTNING_NODE_IP, APP_LIGHTNING_NODE_GRPC_PORT, APP_BITCOIN_NETWORK, APP_LIGHTNING_NODE_DATA_DIR
- `transmission` — shared ${UMBREL_ROOT} storage mapped to a private volume — review
- `tubearchivist` — unmapped tokens: DEVICE_DOMAIN_NAME
- `twenty` — unmapped tokens: DEVICE_DOMAIN_NAME
- `umami` — 
- `usocial` — unmapped tokens: APP_LIGHTNING_NODE_IP, APP_LIGHTNING_NODE_GRPC_PORT, ${APP_LIGHTNING_NODE_DATA_DIR}, APP_LIGHTNING_NODE_DATA_DIR
- `vert` — unmapped tokens: DEVICE_DOMAIN_NAME
- `vikunja` — unmapped tokens: DEVICE_DOMAIN_NAME
- `wanderer` — unmapped tokens: DEVICE_DOMAIN_NAME
- `watch-your-lan` — 
- `wger` — unmapped tokens: DEVICE_DOMAIN_NAME, APP_WGER_PORT, APP_WGER_LOCAL_URLS, APP_DATA_DIR
- `wireguard` — 
- `woofbot-lnd` — unmapped tokens: ${APP_LIGHTNING_NODE_DATA_DIR}/tls.cert, ${APP_LIGHTNING_NODE_DATA_DIR}/data/chain/bitcoin/${APP_BITCOIN_NETWORK}/readonly.macaroon, APP_LIGHTNING_NODE_DATA_DIR, APP_BITCOIN_NETWORK
- `zabbix` — 
- `zerotier` — 
- `zigbee2mqtt` — 

## ✗ Failed
