#!/usr/bin/env bash
#
# Compile slashnoded pour toutes les plateformes cibles et génère les sommes
# SHA-256 attendues par bootstrap.sh.
#
#   ./scripts/build.sh [version]
#
# Sortie dans dist/ :
#   slashnoded-linux-amd64(.sha256)
#   slashnoded-linux-arm64(.sha256)
#   slashnoded-darwin-amd64(.sha256)   # Apple Intel
#   slashnoded-darwin-arm64(.sha256)   # Apple Silicon

set -euo pipefail

cd "$(dirname "$0")/.."

VERSION="${1:-${SLASHNODE_VERSION:-dev}}"
OUT="dist"
PKG="github.com/slashbinslashnoname/slashnode/cmd/slashnoded"
LDFLAGS="-s -w -X main.Version=${VERSION}"

# os/arch à produire.
TARGETS=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
)

mkdir -p "$OUT"
echo "Build slashnoded ${VERSION}"

# --- Binaires Go (multi-arch) ---
for t in "${TARGETS[@]}"; do
  os="${t%/*}"; arch="${t#*/}"
  bin="${OUT}/slashnoded-${os}-${arch}"
  echo "→ ${os}/${arch}"
  CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" \
    go build -trimpath -ldflags "$LDFLAGS" -o "$bin" "$PKG"
  # Somme de contrôle (nom de fichier relatif pour `sha256sum -c`).
  ( cd "$OUT" && sha256sum "$(basename "$bin")" > "$(basename "$bin").sha256" )
done

# --- Front Next.js (bundle autonome) ---
echo "→ front Next.js"
(
  cd web
  [ -d node_modules ] || npm ci
  npm run build
)
# Assemble le build standalone : server.js + node_modules minimal + assets.
rm -rf "${OUT}/web"
mkdir -p "${OUT}/web"
cp -r web/.next/standalone/. "${OUT}/web/"
mkdir -p "${OUT}/web/.next"
cp -r web/.next/static "${OUT}/web/.next/static"
[ -d web/public ] && cp -r web/public "${OUT}/web/public"
tar -czf "${OUT}/slashnode-web.tar.gz" -C "${OUT}/web" .
( cd "$OUT" && sha256sum slashnode-web.tar.gz > slashnode-web.tar.gz.sha256 )

echo
echo "Artefacts :"
ls -1 "$OUT"
