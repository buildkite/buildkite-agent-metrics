#!/bin/bash
set -euo pipefail

source .buildkite/lib/release_dry_run.sh

release_dry_run .buildkite/steps/upload-to-s3.sh release
