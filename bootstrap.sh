#!/usr/bin/env bash
#
# SlashNode bootstrap — installs (or updates) slashnoded on an
# existing Debian/Ubuntu.
#
#   curl -fsSL https://raw.githubusercontent.com/slashbinslashnoname/slashnode/main/bootstrap.sh | bash
#
# Recommended audit before running (crypto-conscious audience):
#
#   curl -fsSL https://raw.githubusercontent.com/slashbinslashnoname/slashnode/main/bootstrap.sh -o slashnode.sh
#   less slashnode.sh
#   bash slashnode.sh
#
# The script stays minimal: it gets the binary onto the machine and delegates all
# bootstrap logic to `slashnoded init` (config, secrets, systemd, Avahi),
# versioned and tested inside the Go binary.

set -euo pipefail

SLASHNODE_VERSION="${SLASHNODE_VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
REPO="slashbinslashnoname/slashnode"
BASE_URL="https://github.com/${REPO}/releases/download"

# Minimal prerequisites.
MIN_RAM_MB=1024
MIN_DISK_GB=10

# Extra arguments passed to `slashnoded init` (filled by configure_access).
INIT_ARGS=()
# Whether the operator opted into Tor hidden services (set by configure_access).
ENABLE_TOR=0

red()  { printf '\033[1;31m%s\033[0m\n' "$*"; }
dim()  { printf '\033[2m%s\033[0m\n' "$*"; }

banner() {
  printf '\033[1;31m      //\n     //\n    //\n   //\n  //\n //\033[0m\n'
  printf '\033[1;31m/\033[0mSlashNode\n'
  printf '\033[2myour node, your rules\033[0m\n\n'
}
info() { printf '→ %s\n' "$*"; }
die()  { printf '\033[1;31m✗ %s\033[0m\n' "$*" >&2; exit 1; }

require_root() {
  if [ "${EUID:-$(id -u)}" -ne 0 ]; then
    command -v sudo >/dev/null || die "root required and sudo missing."
    exec sudo -E bash "$0" "$@"
  fi
}

detect_arch() {
  case "$(uname -m)" in
    x86_64)  echo amd64 ;;
    aarch64|arm64) echo arm64 ;;
    *) die "unsupported architecture: $(uname -m) (amd64/arm64 only)" ;;
  esac
}

pre_checks() {
  info "Running pre-checks…"
  [ "$(uname -s)" = "Linux" ] || die "unsupported OS: $(uname -s) (Linux only)."
  if [ -r /etc/os-release ]; then
    . /etc/os-release
    case "${ID:-} ${ID_LIKE:-}" in
      *debian*|*ubuntu*) : ;;
      *) red "⚠ untested distribution (${ID:-unknown}) — continuing anyway." ;;
    esac
  fi

  local ram_mb
  ram_mb=$(awk '/MemTotal/ {print int($2/1024)}' /proc/meminfo 2>/dev/null || echo 0)
  [ "$ram_mb" -ge "$MIN_RAM_MB" ] || red "⚠ low RAM: ${ram_mb}MB (recommended ≥ ${MIN_RAM_MB}MB)."

  local disk_gb
  disk_gb=$(df -BG --output=avail / 2>/dev/null | tail -1 | tr -dc '0-9' || echo 0)
  [ "${disk_gb:-0}" -ge "$MIN_DISK_GB" ] || red "⚠ low disk: ${disk_gb}GB free (recommended ≥ ${MIN_DISK_GB}GB)."

  for c in curl install; do
    command -v "$c" >/dev/null || die "required command missing: $c"
  done
}

install_docker() {
  if command -v docker >/dev/null; then
    info "Docker already present."
    return
  fi
  info "Installing Docker (official script)…"
  curl -fsSL https://get.docker.com | sh
}

install_node() {
  # Node is required for the Next.js front end launched by the daemon.
  if command -v node >/dev/null; then
    info "Node already present ($(node -v))."
    return
  fi
  info "Installing Node.js (NodeSource 22.x)…"
  curl -fsSL https://deb.nodesource.com/setup_22.x | bash -
  apt-get install -y nodejs
}

install_caddy() {
  # Caddy reverse-proxies each app under its own subdomain. avahi-utils lets us
  # advertise *.slashnode.local on the LAN.
  if ! command -v caddy >/dev/null; then
    info "Installing Caddy…"
    apt-get install -y debian-keyring debian-archive-keyring apt-transport-https curl
    curl -fsSL "https://dl.cloudsmith.io/public/caddy/stable/gpg.key" \
      | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
    curl -fsSL "https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt" \
      > /etc/apt/sources.list.d/caddy-stable.list
    apt-get update -y
    apt-get install -y caddy
  else
    info "Caddy already present."
  fi
  command -v avahi-publish >/dev/null || apt-get install -y avahi-utils || true
}

install_tor() {
  # Tor exposes the UI and apps as .onion hidden services. Only installed when
  # the operator opts in (see configure_access).
  if command -v tor >/dev/null; then
    info "Tor already present."
  else
    info "Installing Tor…"
    apt-get install -y tor
  fi
  systemctl enable tor >/dev/null 2>&1 || true
}

# fetch_verify <name>: download /tmp/<name> and /tmp/<name>.sha256 and verify
# the checksum, retrying on any failure. The rolling "latest" release is
# refreshed by CI, so a download caught mid-publish is retried rather than fatal.
fetch_verify() {
  local name="$1"
  local url="${BASE_URL}/${SLASHNODE_VERSION}/${name}"
  local attempt
  for attempt in 1 2 3 4 5; do
    if curl -fsSL -o "/tmp/${name}" "$url" \
      && curl -fsSL -o "/tmp/${name}.sha256" "${url}.sha256" \
      && ( cd /tmp && sha256sum -c "${name}.sha256" >/dev/null 2>&1 ); then
      return 0
    fi
    info "fetch/verify of ${name} failed (release may be mid-publish), retry in 5s (${attempt}/5)…"
    sleep 5
  done
  die "could not fetch/verify ${name} after retries — installation aborted."
}

install_web() {
  info "Downloading the Next.js front end…"
  fetch_verify slashnode-web.tar.gz
  rm -rf /usr/share/slashnode/web
  mkdir -p /usr/share/slashnode/web
  tar -xzf /tmp/slashnode-web.tar.gz -C /usr/share/slashnode/web
  rm -f /tmp/slashnode-web.tar.gz /tmp/slashnode-web.tar.gz.sha256
  info "Front end installed: /usr/share/slashnode/web"
}

install_apps() {
  info "Downloading the app catalog…"
  fetch_verify slashnode-apps.tar.gz
  rm -rf /usr/share/slashnode/apps
  mkdir -p /usr/share/slashnode/apps
  tar -xzf /tmp/slashnode-apps.tar.gz -C /usr/share/slashnode/apps
  rm -f /tmp/slashnode-apps.tar.gz /tmp/slashnode-apps.tar.gz.sha256
  info "App catalog installed: /usr/share/slashnode/apps"
}

# Interactively choose the access mode and (optional) password protection.
# Reads from /dev/tty so it works even under `curl | bash`. Falls back to local
# mode with no password when no terminal is available.
configure_access() {
  if [ ! -r /dev/tty ]; then
    info "No terminal: defaulting to local mode (no password)."
    INIT_ARGS=(--access local)
    return
  fi

  printf '\nAccess mode:\n' >/dev/tty
  printf '  1) local  — reachable on your LAN as slashnode.local (default)\n' >/dev/tty
  printf '  2) server — public address, password protected\n' >/dev/tty
  printf '> ' >/dev/tty
  local mode; read -r mode </dev/tty

  if [ "$mode" = "2" ]; then
    local addr pass
    printf 'Public address (e.g. node.example.com): ' >/dev/tty
    read -r addr </dev/tty
    printf 'Set admin password: ' >/dev/tty
    read -rs pass </dev/tty; printf '\n' >/dev/tty
    INIT_ARGS=(--access server --address "$addr" --password "$pass" --password-protect)
  else
    local yn pass
    printf 'Protect the local UI with a password? [y/N] ' >/dev/tty
    read -r yn </dev/tty
    case "$yn" in
      y|Y)
        printf 'Set admin password: ' >/dev/tty
        read -rs pass </dev/tty; printf '\n' >/dev/tty
        INIT_ARGS=(--access local --password "$pass" --password-protect)
        ;;
      *)
        INIT_ARGS=(--access local)
        ;;
    esac
  fi

  local tor
  printf 'Expose the UI and apps as Tor hidden services (.onion)? [y/N] ' >/dev/tty
  read -r tor </dev/tty
  case "$tor" in
    y|Y) ENABLE_TOR=1; INIT_ARGS+=(--tor) ;;
  esac
}

# Compares the installed version to the target. Returns 0 if a (re)install is
# needed, 1 if already up to date.
needs_install() {
  local target="$1"
  command -v slashnoded >/dev/null || return 0
  [ "$target" = "latest" ] && return 0   # can't compare against "latest": (re)install.
  local current
  current="$(slashnoded version 2>/dev/null | awk '{print $2}')" || return 0
  if [ "$current" = "$target" ]; then
    info "slashnoded ${current} already up to date."
    return 1
  fi
  info "updating ${current} → ${target}."
  return 0
}

install_binary() {
  local arch tag
  arch="$(detect_arch)"
  tag="$SLASHNODE_VERSION"

  # Download under the original artifact name so the checksum file (which
  # references that name) verifies with `sha256sum -c`.
  local name="slashnoded-linux-${arch}"
  info "Downloading slashnoded (linux/${arch}, ${tag})…"
  fetch_verify "$name"
  install -m 0755 "/tmp/${name}" "${INSTALL_DIR}/slashnoded"
  rm -f "/tmp/${name}" "/tmp/${name}.sha256"
  info "Binary installed: ${INSTALL_DIR}/slashnoded"
}

main() {
  require_root "$@"
  banner
  pre_checks
  install_docker
  install_node
  install_caddy

  if needs_install "$SLASHNODE_VERSION"; then
    install_binary
  fi
  install_web
  install_apps

  configure_access
  if [ "$ENABLE_TOR" = "1" ]; then
    install_tor
  fi
  info "Initializing (config, secrets, systemd, Avahi, Caddy)…"
  slashnoded init --quiet "${INIT_ARGS[@]}"

  info "Enabling the service…"
  systemctl daemon-reload
  systemctl enable slashnoded
  # restart (not just enable --now) so a reinstall picks up the new binary AND
  # the new web build — otherwise the old node process serves stale asset names.
  systemctl restart slashnoded
  systemctl enable --now slashnoded-update.timer
  systemctl enable --now slashnoded-prune.timer 2>/dev/null || true
  systemctl enable --now caddy 2>/dev/null || true
  systemctl reload caddy 2>/dev/null || systemctl restart caddy 2>/dev/null || true

  echo
  slashnoded status --post-install
}

main "$@"
