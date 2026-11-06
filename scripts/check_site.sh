#!/usr/bin/env bash
# Verify the built site: required pages exist and internal links in docs resolve.
set -euo pipefail

PUBLIC="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/website/public"

required=(index.html 404.html docs/index.html docs/manual/index.html docs/cli/index.html
  docs/cli-reference/index.html docs/trees/index.html docs/api/index.html docs/api/reference/index.html
  docs/mcp/index.html docs/tui/index.html references/index.html resources/index.html schemas/beetree.schema.json app/index.html vendor/swagger/swagger-ui-bundle.js js/swagger-init.js communitree.css js/communitree.js version.json sitemap.xml site-map/index.html)
grep -Eq 'class="?ver[" >]' "$PUBLIC/index.html" || { echo "footer version missing" >&2; exit 1; }
grep -Eq 'id="?menu-toggle"?[^>]*aria-label' "$PUBLIC/index.html" || { echo "menu toggle must be an icon button with aria-label" >&2; exit 1; }
grep -q 'prefers-color-scheme' "$PUBLIC/js/site.js" || { echo "site.js must follow prefers-color-scheme until the visitor chooses" >&2; exit 1; }
for p in index.html docs/index.html learn/building-blocks/index.html; do
  for t in 'property="?og:title' 'property="?og:image"? content="?https://[^" >]*/img/explore.jpg' 'name="?twitter:card'; do
    grep -Eq "$t" "$PUBLIC/$p" || { echo "$p missing social card tag: $t" >&2; exit 1; }
  done
done
[ -s "$PUBLIC/img/explore.jpg" ] || { echo "missing: img/explore.jpg" >&2; exit 1; }
grep -q 'public/trees' "$PUBLIC/docs/api/index.html" || { echo "docs/api must document /v1/public/trees" >&2; exit 1; }
for d in docs/api docs/mcp; do
  grep -q 'X-API-Key' "$PUBLIC/$d/index.html" || { echo "$d must document X-API-Key" >&2; exit 1; }
done
# every curated resource needs an https link, a summary and a passing score
data="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/website/data/resources.yaml"
n=$(grep -c '^  - title:' "$data")
[ "$n" -ge 10 ] || { echo "resources.yaml has only $n entries" >&2; exit 1; }
[ "$(grep -cE '^    url: "https?://' "$data")" = "$n" ] && [ "$(grep -c '^    summary:' "$data")" = "$n" ] || { echo "every resource needs url and summary" >&2; exit 1; }
awk '/^    score:/ && $2 < 9 {bad=1} END {exit bad}' "$data" || { echo "a resource scores below the rubric threshold of 9" >&2; exit 1; }
# the TUI page leads with the step-through debugger and embeds its recorded demos
[ "$(grep -oE '<img[^>]*/img/tui-[a-z-]+\.gif' "$PUBLIC/docs/tui/index.html" | wc -l)" -ge 2 ] || { echo "docs/tui must embed its GIFs" >&2; exit 1; }
for g in tui-step-through.gif tui-editor.gif; do [ -s "$PUBLIC/img/$g" ] || { echo "missing: img/$g" >&2; exit 1; }; done
for f in "${required[@]}"; do
  [ -f "$PUBLIC/$f" ] || { echo "missing: $f" >&2; exit 1; }
done

# the building-blocks page must render its tabbed code (raw HTML is dropped silently if a shortcode is misused)
[ "$(grep -oE 'class="?codetab[" >]' "$PUBLIC/learn/building-blocks/index.html" | wc -l)" -ge 28 ] || { echo "building-blocks code tabs missing" >&2; exit 1; }

# the human site map must list every top-level section and be linked from the footer
for l in /app/ /learn/behavior-trees/ /docs/ /docs/api/ /docs/mcp/ /references/ '/app/#/explore' '/app/#/forum' /sitemap.xml; do
  grep -Eq "href=\"?$l[\" >]" "$PUBLIC/site-map/index.html" || { echo "site map missing $l" >&2; exit 1; }
done
grep -Eq 'href="?/site-map/[" >]' "$PUBLIC/index.html" || { echo "footer must link /site-map/" >&2; exit 1; }

# the home page names the trope templates and links to the live preview
for t in 'Patrol loop' 'Guard' 'Coward' 'Boss with phases' 'Loyal companion'; do
  grep -q "<h3>$t</h3>" "$PUBLIC/index.html" || { echo "home must name template: $t" >&2; exit 1; }
done
tr -d '\n' < "$PUBLIC/index.html" | grep -oE '<section class="?wrap"? id="?templates.*' | sed 's#</section>.*##' | grep -Eq 'href="?/app/#/explore' || { echo "home templates section must link to the live preview" >&2; exit 1; }

# sign-up belongs below the fold: the hero section must not link to it
hero=$(tr -d '\n' < "$PUBLIC/index.html" | grep -oE '<section class="?hero[" >].*' | sed 's#</section>.*##')
[ -n "$hero" ] || { echo "home hero not found" >&2; exit 1; }
if printf '%s' "$hero" | grep -q '#/signup'; then echo "hero must not link to sign-up" >&2; exit 1; fi

# the builder bundle ships the copy and download controls
for c in copy-yaml download-yaml copy-code download-code; do
  grep -rqF "$c" "$PUBLIC/app/assets" || { echo "builder is missing the $c control" >&2; exit 1; }
done

# the dashboard ships tree search, sort and rename
for c in tree-search tree-sort 'Renamed.'; do
  grep -rqF "$c" "$PUBLIC/app/assets" || { echo "dashboard is missing the $c control" >&2; exit 1; }
done

# every root-relative href in the built HTML must resolve (the /app/ SPA owns its #/ routes)
bad=0
while IFS= read -r link; do
  path="${link%%#*}"; path="${path%%\?*}"
  [ -z "$path" ] && continue
  if [ -e "$PUBLIC$path" ] || [ -e "$PUBLIC${path%/}/index.html" ]; then continue; fi
  echo "broken link: $link" >&2; bad=1
done < <(grep -rhoE 'href="/[^"]*"' "$PUBLIC" --include=*.html | sed -E 's/^href="//; s/"$//' | sort -u)
exit $bad
