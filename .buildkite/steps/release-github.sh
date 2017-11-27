#!/bin/bash
set -e

if [[ "$GITHUB_RELEASE_ACCESS_TOKEN" == "" ]]; then
  echo "Error: Missing \$GITHUB_RELEASE_ACCESS_TOKEN"
  exit 1
fi

version=$(awk -F\" '/const Version/ {print $2}' version/version.go)

echo '--- Downloading binaries'

rm -rf pkg
mkdir -p pkg
buildkite-agent artifact download "dist/*" .

docker run -e GITHUB_RELEASE_ACCESS_TOKEN --rm "buildkite/github-release" "${version}" dist/* \
  --commit "${BUILDKITE_COMMIT}" \
  --tag "${version}" \
  --github-repository "buildkite/buildkite-metrics"
