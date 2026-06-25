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

red()  { printf '\033[1;31m%s\033[0m\n' "$*"; }
dim()  { printf '\033[2m%s\033[0m\n' "$*"; }
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

install_web() {
  local tag="$SLASHNODE_VERSION"
  local url="${BASE_URL}/${tag}/slashnode-web.tar.gz"
  info "Downloading the Next.js front end…"
  curl -fsSL -o /tmp/slashnode-web.tar.gz "$url" \
    || die "front end download failed: $url"
  curl -fsSL -o /tmp/slashnode-web.tar.gz.sha256 "${url}.sha256" \
    || die "front end checksum not found: ${url}.sha256"
  ( cd /tmp && sha256sum -c slashnode-web.tar.gz.sha256 >/dev/null 2>&1 ) \
    || die "invalid front end checksum — installation aborted."
  rm -rf /usr/share/slashnode/web
  mkdir -p /usr/share/slashnode/web
  tar -xzf /tmp/slashnode-web.tar.gz -C /usr/share/slashnode/web
  rm -f /tmp/slashnode-web.tar.gz /tmp/slashnode-web.tar.gz.sha256
  info "Front end installed: /usr/share/slashnode/web"
}

install_apps() {
  local tag="$SLASHNODE_VERSION"
  local url="${BASE_URL}/${tag}/slashnode-apps.tar.gz"
  info "Downloading the app catalog…"
  curl -fsSL -o /tmp/slashnode-apps.tar.gz "$url" \
    || die "app catalog download failed: $url"
  curl -fsSL -o /tmp/slashnode-apps.tar.gz.sha256 "${url}.sha256" \
    || die "app catalog checksum not found: ${url}.sha256"
  ( cd /tmp && sha256sum -c slashnode-apps.tar.gz.sha256 >/dev/null 2>&1 ) \
    || die "invalid app catalog checksum — installation aborted."
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
  [ "$tag" = "latest" ] && tag="latest"

  local url="${BASE_URL}/${tag}/slashnoded-linux-${arch}"
  info "Downloading slashnoded (linux/${arch}, ${tag})…"
  curl -fsSL -o /tmp/slashnoded "$url" \
    || die "download failed: $url"
  curl -fsSL -o /tmp/slashnoded.sha256 "${url}.sha256" \
    || die "checksum not found: ${url}.sha256"

  info "Verifying checksum…"
  ( cd /tmp && sha256sum -c slashnoded.sha256 >/dev/null 2>&1 ) \
    || die "invalid checksum — installation aborted."

  install -m 0755 /tmp/slashnoded "${INSTALL_DIR}/slashnoded"
  rm -f /tmp/slashnoded /tmp/slashnoded.sha256
  info "Binary installed: ${INSTALL_DIR}/slashnoded"
}

main() {
  require_root "$@"
  red "SlashNode — installation"
  pre_checks
  install_docker
  install_node

  if needs_install "$SLASHNODE_VERSION"; then
    install_binary
  fi
  install_web
  install_apps

  configure_access
  info "Initializing (config, secrets, systemd, Avahi)…"
  slashnoded init --quiet "${INIT_ARGS[@]}"

  info "Enabling the service…"
  systemctl daemon-reload
  systemctl enable --now slashnoded
  systemctl enable --now slashnoded-update.timer

  echo
  slashnoded status --post-install
}

main "$@"
