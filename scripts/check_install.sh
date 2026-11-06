#!/usr/bin/env bash
# Gate: install.sh passes shellcheck, dry-runs, installs from a fake site, and rejects a bad checksum.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
S="$ROOT/website/static/install.sh"
shellcheck "$S"
T="$(mktemp -d)"; trap 'rm -rf "$T"' EXIT
V=v9.9.9
mkdir -p "$T/in" "$T/stage/beetree_${V}_linux_amd64"
printf '#!/bin/sh\necho beetree fake\n' > "$T/stage/beetree_${V}_linux_amd64/beetree"
chmod 755 "$T/stage/beetree_${V}_linux_amd64/beetree"
tar -C "$T/stage" -czf "$T/in/beetree_${V}_linux_amd64.tar.gz" "beetree_${V}_linux_amd64"
for p in linux_arm64 darwin_amd64 darwin_arm64; do echo x > "$T/in/beetree_${V}_$p.tar.gz"; done
echo x > "$T/in/beetree_${V}_windows_amd64.zip"
"$ROOT/scripts/downloads_layout.sh" "$V" "$T/in" "$T/site" >/dev/null
export BEETREE_BASE="file://$T/site" BEETREE_FORCE_PLATFORM=linux_amd64

out="$(BEETREE_DRY_RUN=1 BEETREE_INSTALL_DIR="$T/bin" sh "$S")"
grep -q "beetree $V for linux/amd64" <<<"$out" || { echo "dry run must name version and platform" >&2; exit 1; }
[ ! -e "$T/bin" ] || { echo "dry run must not install" >&2; exit 1; }

BEETREE_INSTALL_DIR="$T/bin" sh "$S" >/dev/null
[ "$("$T/bin/beetree")" = "beetree fake" ] || { echo "installed binary does not run" >&2; exit 1; }

rm -rf "$T/bin"
printf 'tampered' >> "$T/site/downloads/latest/beetree_linux_amd64.tar.gz"
if BEETREE_INSTALL_DIR="$T/bin" sh "$S" >/dev/null 2>&1; then echo "tampered archive must be rejected" >&2; exit 1; fi
[ ! -e "$T/bin/beetree" ] || { echo "nothing may be installed after a checksum failure" >&2; exit 1; }
echo "install script ok"
