#!/usr/bin/env bash
# Lay out CLI release assets for the site bucket.
# Usage: downloads_layout.sh vX.Y.Z <assets-dir> <out-dir>
#   <out-dir>/downloads/<version>/   archives with their release names + SHA256SUMS
#   <out-dir>/downloads/latest/      stable names (beetree_<os>_<arch>.<ext>) + SHA256SUMS + latest.json
set -euo pipefail

VERSION="${1:?usage: downloads_layout.sh vX.Y.Z assets-dir out-dir}"
SRC="${2:?assets-dir}"
OUT="${3:?out-dir}"
[[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo "version must look like v1.2.3" >&2; exit 2; }

VDIR="$OUT/downloads/$VERSION"
LDIR="$OUT/downloads/latest"
rm -rf "$OUT/downloads" && mkdir -p "$VDIR" "$LDIR"

files=()
for f in "$SRC"/beetree_"${VERSION}"_*.tar.gz "$SRC"/beetree_"${VERSION}"_*.zip; do
  [ -f "$f" ] && files+=("$f")
done
[ "${#files[@]}" -gt 0 ] || { echo "no CLI archives for $VERSION in $SRC" >&2; exit 1; }

entries=()
for f in "${files[@]}"; do
  base="$(basename "$f")"
  stable="${base/_${VERSION}_/_}"
  cp "$f" "$VDIR/$base"
  cp "$f" "$LDIR/$stable"
  sum="$(sha256sum "$f" | cut -d' ' -f1)"
  rest="${stable#beetree_}"; rest="${rest%.tar.gz}"; rest="${rest%.zip}"
  os="${rest%_*}"; arch="${rest#*_}"
  entries+=("{\"name\":\"$stable\",\"os\":\"$os\",\"arch\":\"$arch\",\"sha256\":\"$sum\",\"url\":\"/downloads/latest/$stable\",\"versioned_url\":\"/downloads/$VERSION/$base\"}")
done

(cd "$VDIR" && sha256sum ./*.tar.gz ./*.zip 2>/dev/null | sed 's# \./# #' > SHA256SUMS)
(cd "$LDIR" && sha256sum ./*.tar.gz ./*.zip 2>/dev/null | sed 's# \./# #' > SHA256SUMS)

printf '{"version":"%s","files":[%s]}\n' "$VERSION" "$(IFS=,; echo "${entries[*]}")" > "$LDIR/latest.json"
echo "laid out ${#files[@]} archives under $OUT/downloads"
