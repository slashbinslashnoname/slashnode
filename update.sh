#!/usr/bin/env bash
#
# SlashNode updater — refreshes the binary, web bundle and app catalog to the
# latest release and restarts the service. Keeps your config, secrets and access
# settings. For a first install use bootstrap.sh instead.
#
#   curl -fsSL https://raw.githubusercontent.com/slashbinslashnoname/slashnode/master/update.sh | bash

set -euo pipefail

SLASHNODE_VERSION="${SLASHNODE_VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
REPO="slashbinslashnoname/slashnode"
BASE_URL="https://github.com/${REPO}/releases/download"

red()  { printf '\033[1;31m%s\033[0m\n' "$*"; }
info() { printf '→ %s\n' "$*"; }
die()  { printf '\033[1;31m✗ %s\033[0m\n' "$*" >&2; exit 1; }

banner() {
  printf '\033[1;31m      //\n     //\n    //\n   //\n  //\n //\033[0m\n'
  printf '\033[1;31m/\033[0mSlashNode \033[2mupdate\033[0m\n\n'
}

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

# fetch_verify <name>: download /tmp/<name> (+ .sha256) and verify, with retries
# (the rolling release may be mid-publish).
fetch_verify() {
  local name="$1" url="${BASE_URL}/${SLASHNODE_VERSION}/$1" attempt
  for attempt in 1 2 3 4 5; do
    if curl -fsSL -o "/tmp/${name}" "$url" \
      && curl -fsSL -o "/tmp/${name}.sha256" "${url}.sha256" \
      && ( cd /tmp && sha256sum -c "${name}.sha256" >/dev/null 2>&1 ); then
      return 0
    fi
    info "fetch/verify of ${name} failed, retry in 5s (${attempt}/5)…"
    sleep 5
  done
  die "could not fetch/verify ${name} after retries — aborted."
}

extract_bundle() {
  local name="$1" dest="$2"
  fetch_verify "$name"
  rm -rf "$dest"
  mkdir -p "$dest"
  tar -xzf "/tmp/${name}" -C "$dest"
  rm -f "/tmp/${name}" "/tmp/${name}.sha256"
}

main() {
  require_root "$@"
  banner
  command -v slashnoded >/dev/null || die "slashnoded is not installed — run bootstrap.sh first."

  local arch; arch="$(detect_arch)"
  local bin="slashnoded-linux-${arch}"

  info "Updating slashnoded ($arch)…"
  fetch_verify "$bin"
  install -m 0755 "/tmp/${bin}" "${INSTALL_DIR}/slashnoded"
  rm -f "/tmp/${bin}" "/tmp/${bin}.sha256"

  info "Updating the front end…"
  extract_bundle slashnode-web.tar.gz /usr/share/slashnode/web

  info "Updating the app catalog…"
  extract_bundle slashnode-apps.tar.gz /usr/share/slashnode/apps

  info "Re-running init (keeps config + secrets)…"
  slashnoded init --quiet

  info "Restarting…"
  systemctl daemon-reload
  systemctl restart slashnoded
  systemctl reload caddy 2>/dev/null || systemctl restart caddy 2>/dev/null || true

  echo
  red "✓ Updated to $(slashnoded version | awk '{print $2}')"
  slashnoded status --post-install
}

main "$@"
