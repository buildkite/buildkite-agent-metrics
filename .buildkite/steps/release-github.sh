#!/usr/bin/env bash

set -eufo pipefail

source .buildkite/lib/release_dry_run.sh

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

if [[ "$RELEASE_DRY_RUN" != true && $tag != "$BUILDKITE_TAG" ]]; then
  echo "Error: version.go has not been updated to ${BUILDKITE_TAG#v}"
  exit 1
fi

last_tag=$(git describe --tags --abbrev=0 --exclude "$tag")

# escape . so we can use in regex
escaped_tag="${tag//\./\\.}"
escaped_last_tag="${last_tag//\./\\.}"

if [[ "$RELEASE_DRY_RUN" != true ]] && ! grep "^## \[$escaped_tag\]" CHANGELOG.md; then
  echo "Error: CHANGELOG.md has not been updated for $tag" >&2
  exit 1
fi

pushd dist
set +f
sha256sum -- * > sha256sums.txt
set -f
popd

# The three commands below:
#   Find lines between headers of the changelogs (inclusive)
#   Delete the lines included from the headers
#   Trim empty lines from start
# The commad substituion will then delete the empty lines from the end
notes=$(
  sed -n "/\.\.\.${escaped_tag})\$/,/^## \[${escaped_last_tag}\]/p" CHANGELOG.md \
    | sed '1d;$d' \
    | sed '/./,$!d' \
)

echo --- The following notes will accompany the release:
echo "$notes"

echo --- :github: Publishing draft release
set +f
GITHUB_TOKEN="$GITHUB_RELEASE_ACCESS_TOKEN" \
  release_dry_run gh release create \
    --draft \
    --notes "'$notes'" \
    --verify-tag \
    "v$version" \
    dist/*
set -f
