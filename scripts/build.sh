#!/usr/bin/env bash
#
# Builds slashnoded for all target platforms and generates the SHA-256
# checksums expected by bootstrap.sh.
#
#   ./scripts/build.sh [version]
#
# Output in dist/:
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

# os/arch to produce.
TARGETS=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
)

mkdir -p "$OUT"
echo "Build slashnoded ${VERSION}"

# --- Go binaries (multi-arch) ---
for t in "${TARGETS[@]}"; do
  os="${t%/*}"; arch="${t#*/}"
  bin="${OUT}/slashnoded-${os}-${arch}"
  echo "→ ${os}/${arch}"
  CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" \
    go build -trimpath -ldflags "$LDFLAGS" -o "$bin" "$PKG"
  # Checksum (relative file name for `sha256sum -c`).
  ( cd "$OUT" && sha256sum "$(basename "$bin")" > "$(basename "$bin").sha256" )
done

# --- Next.js front end (standalone bundle) ---
echo "→ Next.js front end"
(
  cd web
  [ -d node_modules ] || npm ci
  npm run build
)
# Assemble the standalone build: server.js + minimal node_modules + assets.
rm -rf "${OUT}/web"
mkdir -p "${OUT}/web"
cp -r web/.next/standalone/. "${OUT}/web/"
mkdir -p "${OUT}/web/.next"
cp -r web/.next/static "${OUT}/web/.next/static"
[ -d web/public ] && cp -r web/public "${OUT}/web/public"
tar -czf "${OUT}/slashnode-web.tar.gz" -C "${OUT}/web" .
( cd "$OUT" && sha256sum slashnode-web.tar.gz > slashnode-web.tar.gz.sha256 )

# --- App catalog (manifests) ---
echo "→ app catalog"
rm -rf "${OUT}/apps"
mkdir -p "${OUT}/apps"
cp -r examples/. "${OUT}/apps/"
tar -czf "${OUT}/slashnode-apps.tar.gz" -C "${OUT}/apps" .
( cd "$OUT" && sha256sum slashnode-apps.tar.gz > slashnode-apps.tar.gz.sha256 )

echo
echo "Artifacts:"
ls -1 "$OUT"
