#!/bin/bash
set -e

GITHUB_RELEASE_IMAGE="buildkite/github-release@sha256:e5ae9753a8246ace67f3669baa63103429367c999bfda0a227c91a4ebf34c23f"

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

docker run -v "$PWD:$PWD" -w "$PWD" -e GITHUB_RELEASE_ACCESS_TOKEN --rm "${GITHUB_RELEASE_IMAGE}" "v${version}" dist/* \
  --commit "${BUILDKITE_COMMIT}" \
  --tag "v${version}" \
  --github-repository "buildkite/buildkite-agent-metrics"
