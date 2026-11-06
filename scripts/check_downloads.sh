#!/usr/bin/env bash
# Gate: the downloads layout lists every platform with stable names and verifiable checksums.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
T="$(mktemp -d)"; trap 'rm -rf "$T"' EXIT
V=v9.9.9
mkdir -p "$T/in"
for p in linux_amd64 linux_arm64 darwin_amd64 darwin_arm64; do echo "bin-$p" > "$T/in/beetree_${V}_$p.tar.gz"; done
echo "bin-win" > "$T/in/beetree_${V}_windows_amd64.zip"
echo site > "$T/in/beetreecraft-site_${V}.tar.gz"

"$ROOT/scripts/downloads_layout.sh" "$V" "$T/in" "$T/out" >/dev/null
D="$T/out/downloads"
for n in beetree_linux_amd64.tar.gz beetree_linux_arm64.tar.gz beetree_darwin_amd64.tar.gz beetree_darwin_arm64.tar.gz beetree_windows_amd64.zip SHA256SUMS latest.json; do
  [ -f "$D/latest/$n" ] || { echo "missing downloads/latest/$n" >&2; exit 1; }
done
for n in beetree_${V}_linux_amd64.tar.gz beetree_${V}_windows_amd64.zip SHA256SUMS; do
  [ -f "$D/$V/$n" ] || { echo "missing downloads/$V/$n" >&2; exit 1; }
done
[ ! -e "$D/latest/beetreecraft-site_${V}.tar.gz" ] || { echo "site bundle must not be published as a download" >&2; exit 1; }
(cd "$D/latest" && sha256sum -c SHA256SUMS >/dev/null)
(cd "$D/$V" && sha256sum -c SHA256SUMS >/dev/null)
python3 -I - "$D/latest" <<'PY'
import hashlib, json, sys
d = sys.argv[1]
j = json.load(open(d + "/latest.json"))
assert j["version"] == "v9.9.9", j
assert len(j["files"]) == 5, j
for f in j["files"]:
    assert hashlib.sha256(open(d + "/" + f["name"], "rb").read()).hexdigest() == f["sha256"], f
PY
echo "downloads layout ok"
