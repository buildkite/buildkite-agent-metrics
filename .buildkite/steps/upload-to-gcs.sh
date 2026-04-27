#!/bin/bash
set -eu

export GCP_DEFAULT_REGION="us-central1"

VERSION="$(awk -F\" '/const Version/ {print $2}' version/version.go)"
BASE_BUCKET="buildkite-cloud-functions"
BUCKET_PATH="buildkite-agent-metrics/builds/${BUILDKITE_BUILD_NUMBER}"
ARTIFACT="buildkite-agent-metrics.zip"

echo "~~~ :hammer: Creating dummy artifact for testing"
mkdir -p dist
echo "test" > dist/"${ARTIFACT}"

echo "--- :GCS: Uploading cloud function to ${BASE_BUCKET}/${BUCKET_PATH}/ in ${GCP_DEFAULT_REGION}"
gcloud storage cp dist/"${ARTIFACT}" "gs://${BASE_BUCKET}/${BUCKET_PATH}/${ARTIFACT}"
