#!/bin/bash
set -eu

# Install gcloud CLI if not already available (e.g. on Buildkite hosted agents)
if ! command -v gcloud &> /dev/null; then
  echo "~~~ :googlecloud: Installing gcloud CLI"
  curl -sSL https://sdk.cloud.google.com/google-cloud-cli.tar.gz | tar -xz
  ./google-cloud-sdk/install.sh --quiet --path-update true
  export PATH="$PWD/google-cloud-sdk/bin:$PATH"
fi

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
