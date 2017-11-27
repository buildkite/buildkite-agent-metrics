#!/bin/bash
set -euo pipefail

export VERSION=$(awk -F\" '/const Version/ {print $2}' version/version.go)

git fetch --prune origin "+refs/tags/*:refs/tags/*"

echo "Checking if $VERSION is a tag..."

# If there is already a release (which means a tag), we want to avoid trying to create
# another one, as this will fail and cause partial broken releases
if git rev-parse -q --verify "refs/tags/v${VERSION}" ; then
  echo "Tag refs/tags/v${VERSION} already exists"
  exit 0
fi

buildkite-agent pipeline upload .buildkite/pipeline.release.yml
