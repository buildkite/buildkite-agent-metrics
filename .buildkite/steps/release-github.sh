#!/bin/bash
set -e

if [[ "$GITHUB_RELEASE_ACCESS_TOKEN" == "" ]]; then
  echo "Error: Missing \$GITHUB_RELEASE_ACCESS_TOKEN"
  exit 1
fi

version=$(awk -F\" '/const Version/ {print $2}' version/version.go)

echo '--- Downloading binaries'

rm -rf dist
mkdir -p dist
buildkite-agent artifact download "dist/*" .

docker run -v "$PWD:$PWD" -w "$PWD" -e GITHUB_RELEASE_ACCESS_TOKEN --rm "buildkite/github-release" "v${version}" dist/* \
  --commit "${BUILDKITE_COMMIT}" \
  --tag "v${version}" \
  --github-repository "buildkite/buildkite-agent-metrics"
