#!/usr/bin/env bash
# Verify the built site: required pages exist and internal links in docs resolve.
set -euo pipefail

PUBLIC="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/website/public"

required=(index.html 404.html docs/index.html docs/manual/index.html docs/cli/index.html
  docs/cli-reference/index.html docs/trees/index.html docs/api/index.html docs/api/reference/index.html
  docs/mcp/index.html references/index.html schemas/beetree.schema.json app/index.html vendor/swagger/swagger-ui-bundle.js js/swagger-init.js sitemap.xml)
for f in "${required[@]}"; do
  [ -f "$PUBLIC/$f" ] || { echo "missing: $f" >&2; exit 1; }
done

# the building-blocks page must render its tabbed code (raw HTML is dropped silently if a shortcode is misused)
[ "$(grep -oE 'class="?codetab[" >]' "$PUBLIC/learn/building-blocks/index.html" | wc -l)" -ge 28 ] || { echo "building-blocks code tabs missing" >&2; exit 1; }

# every root-relative href in the built HTML must resolve (the /app/ SPA owns its #/ routes)
bad=0
while IFS= read -r link; do
  path="${link%%#*}"; path="${path%%\?*}"
  [ -z "$path" ] && continue
  if [ -e "$PUBLIC$path" ] || [ -e "$PUBLIC${path%/}/index.html" ]; then continue; fi
  echo "broken link: $link" >&2; bad=1
done < <(grep -rhoE 'href="/[^"]*"' "$PUBLIC" --include=*.html | sed -E 's/^href="//; s/"$//' | sort -u)
exit $bad
