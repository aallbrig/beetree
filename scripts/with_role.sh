#!/usr/bin/env bash
# Run a command under one of the least-privilege roles from infra/iam.yaml.
#   scripts/with_role.sh DeployerRoleArn -- sam deploy ...
# Falls back to the current credentials, with a warning, if the IAM stack is not deployed yet.
set -euo pipefail

KEY="${1:?usage: with_role.sh <OutputKey> -- command...}"
shift
[ "${1:-}" = "--" ] && shift

IAM_STACK="${IAM_STACK:-beetreecraft-iam}"
REGION="${AWS_REGION:-us-east-1}"

ARN="$(aws cloudformation describe-stacks --stack-name "$IAM_STACK" --region "$REGION" \
  --query "Stacks[0].Outputs[?OutputKey=='$KEY'].OutputValue" --output text 2>/dev/null || true)"

if [ -z "$ARN" ] || [ "$ARN" = "None" ]; then
  echo "warning: $IAM_STACK not deployed, using your current credentials (run 'task deploy:iam')" >&2
  exec "$@"
fi

CREDS="$(aws sts assume-role --role-arn "$ARN" --role-session-name "beetreecraft-$KEY" \
  --query 'Credentials.[AccessKeyId,SecretAccessKey,SessionToken]' --output text)"
read -r AK SK ST <<<"$CREDS"
AWS_ACCESS_KEY_ID="$AK" AWS_SECRET_ACCESS_KEY="$SK" AWS_SESSION_TOKEN="$ST" exec "$@"
