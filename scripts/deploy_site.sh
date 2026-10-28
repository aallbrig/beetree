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

# Everything short-lived first, then hashed assets as immutable.
aws s3 sync "$PUBLIC" "s3://$BUCKET" --delete --exclude "app/assets/*" \
  --cache-control "public, max-age=300" --region "$REGION"
aws s3 sync "$PUBLIC/app/assets" "s3://$BUCKET/app/assets" --delete \
  --cache-control "public, max-age=31536000, immutable" --region "$REGION"

aws cloudfront create-invalidation --distribution-id "$DIST" --paths "/*" --query 'Invalidation.Id' --output text
echo "published to $(out SiteUrl)"
