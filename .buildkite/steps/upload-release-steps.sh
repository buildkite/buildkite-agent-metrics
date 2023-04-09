#!/bin/bash
set -euo pipefail

if [[ "${RELEASE_DRY_RUN:-false}" != "true" && "$BUILDKITE_BRANCH" != "${BUILDKITE_TAG:-}" ]]; then
  echo "Skipping release for a non-tag build on $BUILDKITE_BRANCH" >&2
  exit 0
fi

buildkite-agent pipeline upload .buildkite/pipeline.release.yml
