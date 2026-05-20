#!/usr/bin/env bash

set -eufo pipefail

source .buildkite/lib/release_dry_run.sh

if [[ "${RELEASE_DRY_RUN:-false}" != "true" && -z "${RELEASE_VERSION_TAG:-}" ]]; then
  echo "Skipping release: RELEASE_VERSION_TAG not set" >&2
  exit 0
fi

if [[ "${GITHUB_RELEASE_ACCESS_TOKEN}" == "" ]]; then
  echo "Error: Missing \$GITHUB_RELEASE_ACCESS_TOKEN" >&2
  exit 1
fi

export GH_TOKEN="${GITHUB_RELEASE_ACCESS_TOKEN}"

tag="${RELEASE_VERSION_TAG}"

echo --- Downloading binaries

rm -rf dist
mkdir -p dist
buildkite-agent artifact download --build "${BUILDKITE_TRIGGERED_FROM_BUILD_ID}" "dist/*" ./dist

echo --- Creating checksum file
pushd dist &>/dev/null
set +f
sha256sum -- * > sha256sums.txt
set -f
popd &>/dev/null

echo --- :github: Publishing release
set +f
release_dry_run gh release create \
  --repo buildkite/buildkite-agent-metrics \
  --target "$(git rev-parse HEAD)" \
  --generate-notes \
  --fail-on-no-commits \
  "${tag}" \
  dist/*
set -f
