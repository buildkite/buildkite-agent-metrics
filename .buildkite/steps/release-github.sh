#!/usr/bin/env bash

set -eufo pipefail

if [[ "$GITHUB_RELEASE_ACCESS_TOKEN" == "" ]]; then
  echo "Error: Missing \$GITHUB_RELEASE_ACCESS_TOKEN" >&2
  exit 1
fi

echo --- Fetching tags
git fetch --prune --force origin "+refs/tags/*:refs/tags/*"

echo --- Downloading binaries
rm -rf dist
mkdir -p dist
buildkite-agent artifact download handler.zip ./dist
buildkite-agent artifact download "buildkite-agent-metrics-*" ./dist

echo --- Checking tags
version=$(awk -F\" '/const Version/ {print $2}' version/version.go)
tag="v${version#v}"

if [[ $tag != "$BUILDKITE_TAG" ]]; then
  echo "Error: version.go has not been updated to ${BUILDKITE_TAG#v}"
  exit 1
fi

last_tag=$(git describe --tags --abbrev=0 --exclude "$tag")

# escape . so we can use in regex
escaped_tag="${tag//\./\\.}"
escaped_last_tag="${last_tag//\./\\.}"

if ! grep "^## \[$escaped_tag\]" CHANGELOG.md; then
  echo "Error: CHANGELOG.md has not been updated for $tag" >&2
  exit 1
fi

popd dist
set +f
sha256sum ./* > sha256sums.txt
set -f
pushd

# Find lines between headers of the changelogs (inclusive)
# Delete the lines included from the headers
# Trim the newlines
notes=$(
  sed -n "/\.\.\.${escaped_tag})$/,/^## \[${escaped_last_tag}\]/p" CHANGELOG.md \
    | sed '1d;$d' \
    | tr -d '\n' \
)

set +f
GITHUB_TOKEN="$GITHUB_RELEASE_ACCESS_TOKEN" \
  gh release create \
    --draft \
    --notes "$notes" \
    --verify-tags \
    "v$version" \
    dist/*
set -f
