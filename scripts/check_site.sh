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
# every curl block on the MCP page must send the API key, or it returns 401 verbatim
bad=$(sed 's/<[^>]*>//g' "$PUBLIC/docs/mcp/index.html" | awk '
  /curl -s https:\/\/api.beetreecraft.com\/mcp/ {n++; if (open && !keyed) bad++; open=1; keyed=0}
  open && /X-API-Key/ {keyed=1}
  open && /^[[:space:]]*-d / {if (!keyed) bad++; open=0}
  END {print (n < 1 ? 1 : bad + 0)}')
[ "$bad" = 0 ] || { echo "every curl on docs/mcp must send X-API-Key" >&2; exit 1; }
# every curated resource needs an https link, a summary and a passing score
data="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/website/data/resources.yaml"
n=$(grep -c '^  - title:' "$data")
[ "$n" -ge 10 ] || { echo "resources.yaml has only $n entries" >&2; exit 1; }
[ "$(grep -cE '^    url: "https?://' "$data")" = "$n" ] && [ "$(grep -c '^    summary:' "$data")" = "$n" ] || { echo "every resource needs url and summary" >&2; exit 1; }
awk '/^    score:/ && $2 < 9 {bad=1} END {exit bad}' "$data" || { echo "a resource scores below the rubric threshold of 9" >&2; exit 1; }
# the references page credits the inspiration for Demo scene 2
grep -q 'beach-landing style firefight' "$PUBLIC/references/index.html" || { echo "references must credit the scene 2 inspiration" >&2; exit 1; }
# the manual documents the Explore scenes and the scene 2 timer
for t in 'Demo scene 1' 'Demo scene 2' 'damage timer'; do
  grep -q "$t" "$PUBLIC/docs/manual/index.html" || { echo "manual must document: $t" >&2; exit 1; }
done
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

# footer trust links resolve, and the privacy page names what the code stores
for l in /privacy/ /about/ https://github.com/aallbrig/beetreecraft/releases https://github.com/aallbrig/beetreecraft/issues/new; do
  sed -n '/<footer/,/<\/footer>/p' "$PUBLIC/index.html" | grep -Eq "href=\"?$l[\" >]" || { echo "footer must link $l" >&2; exit 1; }
done
for t in Cognito DynamoDB birthdate SHA-256 'tracking'; do
  grep -q "$t" "$PUBLIC/privacy/index.html" || { echo "privacy page must mention: $t" >&2; exit 1; }
done
grep -q 'Andrew Allbright' "$PUBLIC/about/index.html" || { echo "about page must name the maintainer" >&2; exit 1; }

# the home page names the trope templates and links to the live preview
for pair in "shopkeeper:Shopkeeper's routine" 'guard:Guard' 'coward:Coward' 'boss:Boss with phases' 'companion:Loyal companion'; do
  slug="${pair%%:*}"; t="${pair#*:}"
  tr -d '\n' < "$PUBLIC/index.html" | grep -Fq "href=\"/app/#/builder?template=$slug\"><h3>$t</h3>" || { echo "home card for template $slug must be headed: $t" >&2; exit 1; }
  grep -q "id: 'sample:$slug'" web/src/samples.ts || { echo "home links to unknown template: $slug" >&2; exit 1; }
done
tr -d '\n' < "$PUBLIC/index.html" | grep -oE '<section class="?wrap"? id="?templates.*' | sed 's#</section>.*##' | grep -Eq 'href="?/app/#/explore' || { echo "home templates section must link to the live preview" >&2; exit 1; }

# sign-up belongs below the fold: the hero section must not link to it
hero=$(tr -d '\n' < "$PUBLIC/index.html" | grep -oE '<section class="?hero[" >].*' | sed 's#</section>.*##')
[ -n "$hero" ] || { echo "home hero not found" >&2; exit 1; }
printf '%s' "$hero" | grep -q '<h1>Build and simulate game AI behavior trees</h1>' || { echo "hero headline must state the job" >&2; exit 1; }
if printf '%s' "$hero" | grep -q '#/signup'; then echo "hero must not link to sign-up" >&2; exit 1; fi

# the builder bundle ships the copy and download controls
for c in copy-yaml download-yaml copy-code download-code; do
  grep -rqF "$c" "$PUBLIC/app/assets" || { echo "builder is missing the $c control" >&2; exit 1; }
done

# the builder offers a template picker for new trees
grep -rqF "template-picker" "$PUBLIC/app/assets" || { echo "builder is missing the template picker" >&2; exit 1; }
grep -rqF "template-lesson" "$PUBLIC/app/assets" || { echo "builder must show what a template teaches" >&2; exit 1; }

# the dashboard ships tree search, sort and rename
for c in tree-search tree-sort 'Renamed.'; do
  grep -rqF "$c" "$PUBLIC/app/assets" || { echo "dashboard is missing the $c control" >&2; exit 1; }
done

# visitors who ask for reduced motion get no animated transitions, on the site and in the builder
grep -qF 'prefers-reduced-motion' "$PUBLIC/site.css" || { echo "site.css must honour prefers-reduced-motion" >&2; exit 1; }
grep -rqF 'prefers-reduced-motion' "$PUBLIC/app/assets" --include=*.css || { echo "builder css must honour prefers-reduced-motion" >&2; exit 1; }

# one Docs menu holds the docs and the learning pages; the theme toggle lives in the footer, not the nav
nav="$(sed -n '/<nav /,/<\/nav>/p' "$PUBLIC/index.html")"
if printf '%s' "$nav" | grep -q '>Learn</button>'; then echo "nav must not have a separate Learn menu" >&2; exit 1; fi
if printf '%s' "$nav" | grep -qE 'id="?theme-toggle'; then echo "theme toggle must not sit in the nav" >&2; exit 1; fi
for l in /docs/ /learn/behavior-trees/ /learn/building-blocks/ /resources/ /references/; do
  printf '%s' "$nav" | grep -qE "href=\"?$l[\" >]" || { echo "Docs menu lost $l" >&2; exit 1; }
done
sed -n '/<footer/,/<\/footer>/p' "$PUBLIC/index.html" | grep -qE 'id="?theme-toggle' || { echo "footer must hold the theme toggle" >&2; exit 1; }

# every root-relative href in the built HTML must resolve (the /app/ SPA owns its #/ routes)
bad=0
while IFS= read -r link; do
  path="${link%%#*}"; path="${path%%\?*}"
  [ -z "$path" ] && continue
  if [ -e "$PUBLIC$path" ] || [ -e "$PUBLIC${path%/}/index.html" ]; then continue; fi
  echo "broken link: $link" >&2; bad=1
done < <(grep -rhoE 'href="/[^"]*"' "$PUBLIC" --include=*.html | sed -E 's/^href="//; s/"$//' | sort -u)
exit $bad
