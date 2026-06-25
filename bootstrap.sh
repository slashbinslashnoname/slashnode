#!/usr/bin/env bash
#
# SlashNode bootstrap — installe (ou met à jour) slashnoded sur un
# Debian/Ubuntu existant.
#
#   curl -fsSL https://get.slashnode.sh | bash
#
# Audit recommandé avant exécution (public crypto-conscient) :
#
#   curl -fsSL https://get.slashnode.sh -o slashnode.sh
#   less slashnode.sh
#   bash slashnode.sh
#
# Le script reste minimal : il amène le binaire sur la machine et délègue toute
# la logique d'amorçage à `slashnoded init` (config, secrets, systemd, Avahi),
# versionné et testé dans le binaire Go.

set -euo pipefail

SLASHNODE_VERSION="${SLASHNODE_VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
REPO="slashbinslashnoname/slashnode"
BASE_URL="https://github.com/${REPO}/releases/download"

# Pré-requis minimaux.
MIN_RAM_MB=1024
MIN_DISK_GB=10

red()  { printf '\033[1;31m%s\033[0m\n' "$*"; }
dim()  { printf '\033[2m%s\033[0m\n' "$*"; }
info() { printf '→ %s\n' "$*"; }
die()  { printf '\033[1;31m✗ %s\033[0m\n' "$*" >&2; exit 1; }

require_root() {
  if [ "${EUID:-$(id -u)}" -ne 0 ]; then
    command -v sudo >/dev/null || die "root requis et sudo absent."
    exec sudo -E bash "$0" "$@"
  fi
}

detect_arch() {
  case "$(uname -m)" in
    x86_64)  echo amd64 ;;
    aarch64|arm64) echo arm64 ;;
    *) die "architecture non supportée : $(uname -m) (amd64/arm64 uniquement)" ;;
  esac
}

pre_checks() {
  info "Vérifications préalables…"
  [ "$(uname -s)" = "Linux" ] || die "OS non supporté : $(uname -s) (Linux uniquement)."
  if [ -r /etc/os-release ]; then
    . /etc/os-release
    case "${ID:-} ${ID_LIKE:-}" in
      *debian*|*ubuntu*) : ;;
      *) red "⚠ distribution non testée (${ID:-inconnue}) — on continue quand même." ;;
    esac
  fi

  local ram_mb
  ram_mb=$(awk '/MemTotal/ {print int($2/1024)}' /proc/meminfo 2>/dev/null || echo 0)
  [ "$ram_mb" -ge "$MIN_RAM_MB" ] || red "⚠ RAM faible : ${ram_mb}MB (recommandé ≥ ${MIN_RAM_MB}MB)."

  local disk_gb
  disk_gb=$(df -BG --output=avail / 2>/dev/null | tail -1 | tr -dc '0-9' || echo 0)
  [ "${disk_gb:-0}" -ge "$MIN_DISK_GB" ] || red "⚠ disque faible : ${disk_gb}GB libre (recommandé ≥ ${MIN_DISK_GB}GB)."

  for c in curl install; do
    command -v "$c" >/dev/null || die "commande requise absente : $c"
  done
}

install_docker() {
  if command -v docker >/dev/null; then
    info "Docker déjà présent."
    return
  fi
  info "Installation de Docker (script officiel)…"
  curl -fsSL https://get.docker.com | sh
}

install_node() {
  # Node est requis pour le front Next.js lancé par le démon.
  if command -v node >/dev/null; then
    info "Node déjà présent ($(node -v))."
    return
  fi
  info "Installation de Node.js (NodeSource 22.x)…"
  curl -fsSL https://deb.nodesource.com/setup_22.x | bash -
  apt-get install -y nodejs
}

install_web() {
  local tag="$SLASHNODE_VERSION"
  local url="${BASE_URL}/${tag}/slashnode-web.tar.gz"
  info "Téléchargement du front Next.js…"
  curl -fsSL -o /tmp/slashnode-web.tar.gz "$url" \
    || die "téléchargement du front échoué : $url"
  curl -fsSL -o /tmp/slashnode-web.tar.gz.sha256 "${url}.sha256" \
    || die "checksum front introuvable : ${url}.sha256"
  ( cd /tmp && sha256sum -c slashnode-web.tar.gz.sha256 >/dev/null 2>&1 ) \
    || die "checksum du front invalide — installation interrompue."
  rm -rf /usr/share/slashnode/web
  mkdir -p /usr/share/slashnode/web
  tar -xzf /tmp/slashnode-web.tar.gz -C /usr/share/slashnode/web
  rm -f /tmp/slashnode-web.tar.gz /tmp/slashnode-web.tar.gz.sha256
  info "Front installé : /usr/share/slashnode/web"
}

# Compare la version installée à la cible. Renvoie 0 si une (ré)installation est
# nécessaire, 1 si déjà à jour.
needs_install() {
  local target="$1"
  command -v slashnoded >/dev/null || return 0
  [ "$target" = "latest" ] && return 0   # on ne sait pas comparer "latest" : on (re)pose.
  local current
  current="$(slashnoded version 2>/dev/null | awk '{print $2}')" || return 0
  if [ "$current" = "$target" ]; then
    info "slashnoded ${current} déjà à jour."
    return 1
  fi
  info "mise à jour ${current} → ${target}."
  return 0
}

install_binary() {
  local arch tag
  arch="$(detect_arch)"
  tag="$SLASHNODE_VERSION"
  [ "$tag" = "latest" ] && tag="latest"

  local url="${BASE_URL}/${tag}/slashnoded-linux-${arch}"
  info "Téléchargement de slashnoded (linux/${arch}, ${tag})…"
  curl -fsSL -o /tmp/slashnoded "$url" \
    || die "téléchargement échoué : $url"
  curl -fsSL -o /tmp/slashnoded.sha256 "${url}.sha256" \
    || die "checksum introuvable : ${url}.sha256"

  info "Vérification du checksum…"
  ( cd /tmp && sha256sum -c slashnoded.sha256 >/dev/null 2>&1 ) \
    || die "checksum invalide — installation interrompue."

  install -m 0755 /tmp/slashnoded "${INSTALL_DIR}/slashnoded"
  rm -f /tmp/slashnoded /tmp/slashnoded.sha256
  info "Binaire installé : ${INSTALL_DIR}/slashnoded"
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

  info "Initialisation (config, secrets, systemd, Avahi)…"
  slashnoded init --quiet

  info "Activation du service…"
  systemctl daemon-reload
  systemctl enable --now slashnoded
  systemctl enable --now slashnoded-update.timer

  echo
  slashnoded status --post-install
}

main "$@"
