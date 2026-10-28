#!/usr/bin/env bash
# Cut a release: build CLI archives for every platform and publish a GitHub release.
# Usage: scripts/release.sh vX.Y.Z   (run from a clean, up-to-date main after `task ci`)
set -euo pipefail

VERSION="${1:?usage: release.sh vX.Y.Z}"
[[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo "version must look like v1.2.3" >&2; exit 2; }

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST="$ROOT/dist/release"

[ -z "$(git -C "$ROOT" status --porcelain)" ] || { echo "working tree is not clean" >&2; exit 1; }
[ "$(git -C "$ROOT" rev-parse --abbrev-ref HEAD)" = main ] || { echo "release from main" >&2; exit 1; }
git -C "$ROOT" fetch -q origin main
[ "$(git -C "$ROOT" rev-parse HEAD)" = "$(git -C "$ROOT" rev-parse origin/main)" ] || { echo "main is not up to date with origin" >&2; exit 1; }

rm -rf "$DIST" && mkdir -p "$DIST"
for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64; do
  os="${target%/*}"; arch="${target#*/}"
  name="beetree_${VERSION}_${os}_${arch}"
  bin="beetree"; [ "$os" = windows ] && bin="beetree.exe"
  stage="$DIST/$name"; mkdir -p "$stage"
  (cd "$ROOT/cli" && CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -trimpath \
    -ldflags "-s -w -X github.com/aallbrig/beetree-cli/cmd.Version=$VERSION" -o "$stage/$bin" .)
  cp "$ROOT/README.md" "$stage/"
  if [ "$os" = windows ]; then (cd "$DIST" && zip -qr "$name.zip" "$name"); else tar -C "$DIST" -czf "$DIST/$name.tar.gz" "$name"; fi
  rm -rf "$stage"
done
(cd "$DIST" && sha256sum ./*.tar.gz ./*.zip > SHA256SUMS)

git -C "$ROOT" tag -a "$VERSION" -m "$VERSION"
git -C "$ROOT" push -q origin "$VERSION"
gh release create "$VERSION" --title "$VERSION" --generate-notes "$DIST"/*.tar.gz "$DIST"/*.zip "$DIST/SHA256SUMS"
