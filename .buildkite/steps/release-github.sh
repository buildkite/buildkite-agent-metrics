#!/bin/bash
set -e

echo '--- Getting credentials from SSM'
export GITHUB_RELEASE_ACCESS_TOKEN=$(aws ssm get-parameter --name /pipelines/buildkite-agent-metrics/GITHUB_RELEASE_ACCESS_TOKEN --with-decryption --output text --query Parameter.Value --region us-east-1)

if [[ "$GITHUB_RELEASE_ACCESS_TOKEN" == "" ]]; then
  echo "Error: Missing \$GITHUB_RELEASE_ACCESS_TOKEN"
  exit 1
fi

version=$(awk -F\" '/const Version/ {print $2}' version/version.go)

echo '--- Downloading binaries'

rm -rf dist
mkdir -p dist
buildkite-agent artifact download handler.zip ./dist
buildkite-agent artifact download "buildkite-agent-metrics-*" ./dist

github-release "v${version}" dist/* \
  --commit "${BUILDKITE_COMMIT}" \
  --tag "v${version}" \
  --github-repository "buildkite/buildkite-agent-metrics"
