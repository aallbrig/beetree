#!/bin/sh
# Install the beetree CLI from beetreecraft.com.
#   curl -fsSL https://beetreecraft.com/install.sh | sh
# Detects OS and architecture, downloads the archive named in latest.json,
# verifies its SHA-256 checksum and installs to a user-writable directory.
# Environment:
#   BEETREE_INSTALL_DIR  where to put the binary (default: $HOME/.local/bin)
#   BEETREE_BASE         site to download from (default: https://beetreecraft.com)
#   BEETREE_DRY_RUN=1    print what would happen and change nothing
set -eu

BASE="${BEETREE_BASE:-https://beetreecraft.com}"
DIR="${BEETREE_INSTALL_DIR:-$HOME/.local/bin}"
DRY="${BEETREE_DRY_RUN:-0}"

fail() { echo "install: $*" >&2; exit 1; }

case "$(uname -s)" in
  Linux) os=linux ;;
  Darwin) os=darwin ;;
  *) fail "unsupported OS $(uname -s); on Windows download the .zip from $BASE/get/" ;;
esac
case "$(uname -m)" in
  x86_64 | amd64) arch=amd64 ;;
  aarch64 | arm64) arch=arm64 ;;
  *) fail "unsupported architecture $(uname -m)" ;;
esac
if [ -n "${BEETREE_FORCE_PLATFORM:-}" ]; then os="${BEETREE_FORCE_PLATFORM%_*}"; arch="${BEETREE_FORCE_PLATFORM#*_}"; fi

command -v curl >/dev/null 2>&1 || fail "curl is required"
if command -v sha256sum >/dev/null 2>&1; then
  sha() { sha256sum "$1" | cut -d' ' -f1; }
elif command -v shasum >/dev/null 2>&1; then
  sha() { shasum -a 256 "$1" | cut -d' ' -f1; }
else
  fail "sha256sum or shasum is required to verify the download"
fi

manifest="$(curl -fsSL "$BASE/downloads/latest/latest.json")" || fail "could not fetch $BASE/downloads/latest/latest.json"
version="$(printf '%s' "$manifest" | sed -n 's/.*"version":"\([^"]*\)".*/\1/p')"
entry="$(printf '%s' "$manifest" | grep -o '{[^{}]*"os":"'"$os"'","arch":"'"$arch"'"[^{}]*}')" || fail "no build for ${os}/${arch} in latest.json"
name="$(printf '%s' "$entry" | sed -n 's/.*"name":"\([^"]*\)".*/\1/p')"
want="$(printf '%s' "$entry" | sed -n 's/.*"sha256":"\([0-9a-f]\{64\}\)".*/\1/p')"
if [ -z "$version" ] || [ -z "$name" ] || [ -z "$want" ]; then fail "latest.json is malformed"; fi

echo "beetree $version for ${os}/${arch}"
echo "  download: $BASE/downloads/latest/$name"
echo "  install:  $DIR/beetree"
if [ "$DRY" = 1 ]; then echo "dry run: nothing downloaded or installed"; exit 0; fi

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
curl -fsSL -o "$tmp/$name" "$BASE/downloads/latest/$name" || fail "download failed"
got="$(sha "$tmp/$name")"
[ "$got" = "$want" ] || fail "checksum mismatch for $name (expected $want, got $got); nothing installed"
echo "  checksum: ok"

tar -xzf "$tmp/$name" -C "$tmp" || fail "could not unpack $name"
bin="$(find "$tmp" -type f -name beetree | head -n 1)"
[ -n "$bin" ] || fail "archive has no beetree binary"
mkdir -p "$DIR"
cp "$bin" "$DIR/beetree.new" && chmod 755 "$DIR/beetree.new" && mv "$DIR/beetree.new" "$DIR/beetree"
echo "installed $DIR/beetree"
case ":$PATH:" in *":$DIR:"*) ;; *) echo "add $DIR to your PATH to run it: export PATH=\"$DIR:\$PATH\"" ;; esac
