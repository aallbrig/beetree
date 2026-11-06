#!/usr/bin/env bash
# Publish website/public to the site bucket and invalidate CloudFront.
# Expects `task site:build` to have run. Reads bucket/distribution/config from the stack outputs.
set -euo pipefail

STACK="${STACK:-beetreecraft}"
REGION="${AWS_REGION:-us-east-1}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PUBLIC="$ROOT/website/public"

out() {
  aws cloudformation describe-stacks --stack-name "$STACK" --region "$REGION" \
    --query "Stacks[0].Outputs[?OutputKey=='$1'].OutputValue" --output text
}

[ -f "$PUBLIC/index.html" ] || { echo "run 'task site:build' first" >&2; exit 1; }

BUCKET="$(out BucketName)"
DIST="$(out DistributionId)"

# Runtime config for the SPA, generated from the stack so the bundle stays environment-neutral.
cat > "$PUBLIC/app/config.json" <<JSON
{"apiUrl":"$(out ApiUrl)","region":"$REGION","clientId":"$(out UserPoolClientId)","userPoolId":"$(out UserPoolId)"}
JSON

# Everything is short-lived by default; only Vite's content-hashed bundles are immutable.
# /downloads is published separately below, so the sync must not delete it.
aws s3 sync "$PUBLIC" "s3://$BUCKET" --delete --exclude "downloads/*" \
  --cache-control "public, max-age=300" --region "$REGION"

# CLI archives: DOWNLOADS_DIR holds the laid-out downloads/ tree (scripts/downloads_layout.sh).
if [ -n "${DOWNLOADS_DIR:-}" ]; then
  for d in "$DOWNLOADS_DIR"/downloads/v*; do
    aws s3 sync "$d" "s3://$BUCKET/downloads/$(basename "$d")" \
      --cache-control "public, max-age=31536000, immutable" --region "$REGION"
  done
  aws s3 sync "$DOWNLOADS_DIR/downloads/latest" "s3://$BUCKET/downloads/latest" --delete \
    --cache-control "public, max-age=300" --region "$REGION"
  aws s3 cp "$DOWNLOADS_DIR/downloads/latest/latest.json" "s3://$BUCKET/downloads/latest/latest.json" \
    --content-type application/json --cache-control "public, max-age=60" --region "$REGION"
fi
aws s3 cp "$PUBLIC/app/assets" "s3://$BUCKET/app/assets" --recursive --exclude "*" --include "index-*" \
  --cache-control "public, max-age=31536000, immutable" --region "$REGION"

aws cloudfront create-invalidation --distribution-id "$DIST" --paths "/*" --query 'Invalidation.Id' --output text
echo "published to $(out SiteUrl)"
