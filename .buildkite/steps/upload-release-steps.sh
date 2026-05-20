#!/usr/bin/env bash

# Dynamically uploads the release trigger step when version/version.go has been
# bumped to a version that hasn't been released yet. Mirrors the agent repo's
# upload-release-steps.sh: the release pipeline calls `gh release create`,
# which creates the git tag itself — no manual tagging required.

set -euo pipefail

source .buildkite/lib/release_dry_run.sh

version="$(awk -F\" '/const Version/ {print $2}' version/version.go)"
tag="v${version#v}"

# If there is already a release (which means a tag), we want to avoid trying to
# create another one, as this will fail and cause partial broken releases.
if [[ "${RELEASE_DRY_RUN:-false}" == "false" ]] && git ls-remote --tags origin "refs/tags/${tag}" | grep -q "refs/tags/${tag}$"; then
  echo "Tag refs/tags/${tag} already exists"
  exit 0
fi

echo "--- :rocket: Uploading release trigger for ${tag}"

cat <<YAML | buildkite-agent pipeline upload
steps:
  - name: ":rocket: Release ${tag}"
    key: trigger-release
    trigger: "buildkite-agent-metrics-release"
    async: false
    build:
      message: "Release ${tag}, build \${BUILDKITE_BUILD_NUMBER}"
      commit: "\${BUILDKITE_COMMIT}"
      branch: "\${BUILDKITE_BRANCH}"
      env:
        RELEASE_VERSION_TAG: "${tag}"
        RELEASE_DRY_RUN: "${RELEASE_DRY_RUN:-false}"
YAML
