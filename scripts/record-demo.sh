#!/usr/bin/env bash
set -euo pipefail

# ── BeeTree Demo Recorder ──
# Records individual VHS tapes and stitches them into a single demo video.
# Requirements: vhs (charmbracelet.sh/vhs), ffmpeg, beetree on PATH

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SEGMENTS_DIR="$REPO_ROOT/demos/segments"
DIST_DIR="$REPO_ROOT/dist"
CONCAT_LIST="$DIST_DIR/concat.txt"

RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
BOLD='\033[1m'
RESET='\033[0m'

info()  { echo -e "${CYAN}▸${RESET} $*"; }
ok()    { echo -e "${GREEN}✓${RESET} $*"; }
err()   { echo -e "${RED}✗${RESET} $*" >&2; }

# ── Preflight checks ──
for cmd in vhs ffmpeg beetree; do
  if ! command -v "$cmd" &>/dev/null; then
    err "Required command not found: $cmd"
    exit 1
  fi
done

# ── Prep directories ──
mkdir -p "$SEGMENTS_DIR" "$DIST_DIR"

# ── Record each tape ──
TAPES=($(ls "$REPO_ROOT"/demos/[0-9]*.tape 2>/dev/null | sort))
if [[ ${#TAPES[@]} -eq 0 ]]; then
  err "No tape files found in demos/"
  exit 1
fi

info "Recording ${#TAPES[@]} tape(s)..."
FAILED=0
for tape in "${TAPES[@]}"; do
  name="$(basename "$tape" .tape)"
  info "Recording $name..."
  if vhs "$tape" 2>&1 | tail -1; then
    ok "$name recorded"
  else
    err "$name FAILED"
    FAILED=$((FAILED + 1))
  fi
done

if [[ $FAILED -gt 0 ]]; then
  err "$FAILED tape(s) failed to record"
  exit 1
fi

# ── Stitch segments ──
info "Stitching segments into dist/demo.mp4..."

# Build concat list
> "$CONCAT_LIST"
for seg in "$SEGMENTS_DIR"/*.mp4; do
  echo "file '$seg'" >> "$CONCAT_LIST"
done

ffmpeg -y -f concat -safe 0 -i "$CONCAT_LIST" \
  -c:v libx264 -preset medium -crf 23 \
  -pix_fmt yuv420p \
  "$DIST_DIR/demo.mp4" 2>/dev/null

rm -f "$CONCAT_LIST"

ok "Demo video: dist/demo.mp4"
SIZE=$(du -h "$DIST_DIR/demo.mp4" | cut -f1)
info "Size: $SIZE"

echo ""
echo -e "${BOLD}Done!${RESET} Share dist/demo.mp4 or upload to GitHub Releases."
