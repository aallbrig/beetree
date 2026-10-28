#!/usr/bin/env bash
# Run the Store contract test against a throwaway DynamoDB table (never production).
# The schema mirrors infra/template.yaml; keep the two in step.
set -euo pipefail

REGION="${AWS_REGION:-us-east-1}"
TABLE="beetreecraft-contract-$$-$RANDOM"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cleanup() { aws dynamodb delete-table --table-name "$TABLE" --region "$REGION" >/dev/null 2>&1 || true; }
trap cleanup EXIT

aws dynamodb create-table --table-name "$TABLE" --region "$REGION" --billing-mode PAY_PER_REQUEST \
  --attribute-definitions AttributeName=pk,AttributeType=S AttributeName=sk,AttributeType=S \
    AttributeName=gsi1pk,AttributeType=S AttributeName=gsi1sk,AttributeType=S \
    AttributeName=gsi2pk,AttributeType=S AttributeName=gsi2sk,AttributeType=S \
  --key-schema AttributeName=pk,KeyType=HASH AttributeName=sk,KeyType=RANGE \
  --global-secondary-indexes \
    'IndexName=gsi1,KeySchema=[{AttributeName=gsi1pk,KeyType=HASH},{AttributeName=gsi1sk,KeyType=RANGE}],Projection={ProjectionType=ALL}' \
    'IndexName=gsi2,KeySchema=[{AttributeName=gsi2pk,KeyType=HASH},{AttributeName=gsi2sk,KeyType=RANGE}],Projection={ProjectionType=INCLUDE,NonKeyAttributes=[data,likeCount,commentCount,postCount,lastActivity]}' \
  >/dev/null
aws dynamodb wait table-exists --table-name "$TABLE" --region "$REGION"

cd "$ROOT/cli"
BEETREE_TEST_TABLE="$TABLE" AWS_REGION="$REGION" go test -count=1 -run TestDynamoStoreContract -v ./internal/api/
