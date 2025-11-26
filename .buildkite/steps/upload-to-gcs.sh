#!/bin/bash
set -eu

export GCP_DEFAULT_REGION="us-central1"

VERSION="$(awk -F\" '/const Version/ {print $2}' version/version.go)"
BASE_BUCKET="buildkite-cloud-functions"
BUCKET_PATH="buildkite-agent-metrics"
ARTIFACT="buildkite-agent-metrics.zip"

if [[ "${1:-}" == "release" ]] ; then
  BUCKET_PATH="${BUCKET_PATH}/v${VERSION}"
else
  BUCKET_PATH="${BUCKET_PATH}/builds/${BUILDKITE_BUILD_NUMBER}"
fi

echo "~~~ :buildkite: Downloading artifacts"

mkdir -p dist
if [[ "${1:-}" == "release" ]] ; then
  buildkite-agent artifact download --build "${BUILDKITE_TRIGGERED_FROM_BUILD_ID}" dist/"${ARTIFACT}" ./dist
else
  buildkite-agent artifact download dist/"${ARTIFACT}" ./dist
fi

echo "--- :GCS: Uploading cloud function to ${BASE_BUCKET}/${BUCKET_PATH}/ in ${GCP_DEFAULT_REGION}"
gcloud storage cp dist/"${ARTIFACT}" "gs://${BASE_BUCKET}/${BUCKET_PATH}/${ARTIFACT}"

if [[ "${1:-}" == "release" ]] ; then
  gcloud storage cp "gs://${BASE_BUCKET}/${BUCKET_PATH}/${ARTIFACT}" "gs://${BASE_BUCKET}/buildkite-agent-metrics/cloud-function-latest.zip"
fi
