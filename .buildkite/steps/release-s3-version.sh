#!/bin/bash
set -euo pipefail

source .buildkite/lib/release_dry_run.sh

if [[ "${RELEASE_DRY_RUN:-false}" != "true" && "${BUILDKITE_BRANCH}" != "${RELEASE_VERSION_TAG:-}" ]]; then
  echo "Skipping release for a non-tag build on ${BUILDKITE_BRANCH}" >&2
  exit 0
fi

release_dry_run .buildkite/steps/upload-to-s3.sh release
