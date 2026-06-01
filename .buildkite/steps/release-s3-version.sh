#!/bin/bash
set -euo pipefail

source .buildkite/lib/release_dry_run.sh

if [[ "${RELEASE_DRY_RUN:-false}" != "true" && -z "${RELEASE_VERSION_TAG:-}" ]]; then
  echo "Skipping release: RELEASE_VERSION_TAG not set" >&2
  exit 0
fi

release_dry_run .buildkite/steps/upload-to-s3.sh release
