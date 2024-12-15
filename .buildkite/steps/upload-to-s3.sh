#!/bin/bash
set -eu

export AWS_DEFAULT_REGION="us-east-1"

VERSION="$(awk -F\" '/const Version/ {print $2}' version/version.go)"
BASE_BUCKET="buildkite-lambdas"
BUCKET_PATH="buildkite-agent-metrics"

if [[ "${1:-}" == "release" ]] ; then
  BUCKET_PATH="${BUCKET_PATH}/v${VERSION}"
else
  BUCKET_PATH="${BUCKET_PATH}/builds/${BUILDKITE_BUILD_NUMBER}"
fi

echo "~~~ :buildkite: Downloading artifacts"

mkdir -p dist
if [[ "${1:-}" == "release" ]] ; then
  buildkite-agent artifact download --build "${BUILDKITE_TRIGGERED_FROM_BUILD_ID}" dist/handler.zip ./dist
else
  buildkite-agent artifact download dist/handler.zip ./dist
fi

echo "--- :s3: Uploading lambda to ${BASE_BUCKET}/${BUCKET_PATH}/ in ${AWS_DEFAULT_REGION}"
aws s3 cp dist/handler.zip "s3://${BASE_BUCKET}/${BUCKET_PATH}/handler.zip"
