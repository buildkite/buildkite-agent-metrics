#!/usr/bin/env bash

set -eufo pipefail

source .buildkite/lib/release_dry_run.sh

if [[ "${RELEASE_DRY_RUN:-false}" != "true" && "${BUILDKITE_BRANCH}" != "${RELEASE_VERSION_TAG:-}" ]]; then
  echo "Skipping release for a non-tag build on ${BUILDKITE_BRANCH}" >&2
  exit 0
fi

if [[ "${GITHUB_RELEASE_ACCESS_TOKEN}" == "" ]]; then
  echo "Error: Missing \$GITHUB_RELEASE_ACCESS_TOKEN" >&2
  exit 1
fi

echo --- Fetching tags
git fetch --prune --force origin "+refs/tags/*:refs/tags/*"

echo --- Downloading binaries

rm -rf dist
mkdir -p dist
buildkite-agent artifact download --build "${BUILDKITE_TRIGGERED_FROM_BUILD_ID}" "dist/*" ./dist

echo --- Checking tags
version="$(awk -F\" '/const Version/ {print $2}' version/version.go)"
tag="v${version#v}"

if [[ "${RELEASE_DRY_RUN:-false}" != true && "${tag}" != "${RELEASE_VERSION_TAG}" ]]; then
  echo "Error: version.go has not been updated to ${RELEASE_VERSION_TAG#v}"
  exit 1
fi

last_tag=$(git describe --tags --abbrev=0 --exclude "${tag}")

# escape . so we can use in regex
escaped_tag="${tag//\./\\.}"
escaped_last_tag="${last_tag//\./\\.}"

if [[ "${RELEASE_DRY_RUN:-false}" != true ]] && ! grep "^## \[${escaped_tag}\]" CHANGELOG.md; then
  echo "Error: CHANGELOG.md has not been updated for ${tag}" >&2
  exit 1
fi

echo --- Creating checksum file
pushd dist &>/dev/null
set +f
sha256sum -- * > sha256sums.txt
set -f
popd &>/dev/null

echo --- The following notes will accompany the release:
# The sed commands below:
#   Find lines between headers of the changelogs (inclusive)
#   Delete the lines included from the headers
# The command substituion will then delete the empty lines from the end
notes=$(sed -n "/^## \[${escaped_tag}\]/,/^## \[${escaped_last_tag}\]/p" CHANGELOG.md | sed '$d')

echo --- The following notes will accompany the release:
echo "${notes}"

echo --- :github: Publishing draft release
# TODO: add the following flag once github-cli in alpine repo hits v2.27+
# --verify-tag \
set +f
GITHUB_TOKEN="${GITHUB_RELEASE_ACCESS_TOKEN}" \
  release_dry_run gh release create \
    --draft \
    --notes "${notes}" \
    "${tag}" \
    dist/*
set -f
