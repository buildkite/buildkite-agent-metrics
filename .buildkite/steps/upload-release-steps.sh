#!/bin/bash
set -euo pipefail

block_step() {
  local label="$1"
  cat <<YAML
  - wait
  - block: "${label}"
YAML
}

release_step() {
  local label="$1"
  local command="$2"
  cat <<YAML
  - block
  - label: "${label}"
    command: "${command}"
    branches: master
    agents:
      queue: "deploy"
    concurrency: 1
    concurrency_group: 'release_buildkite_metrics'
YAML
}

output_steps_yaml() {
  local version="$1"

  echo "steps:"

  block_step "release?"
  release_step ":s3: $version" ".buildkite/scripts/upload-to-s3.sh"
}

version=$(awk -F\" '/const Version/ {print $2}' version/version.go)

git fetch --tags

echo "Checking if $version is a tag..."

# If there is already a release (which means a tag), we want to avoid trying to create
# another one, as this will fail and cause partial broken releases
if git rev-parse -q --verify "refs/tags/v${version}" >/dev/null; then
  echo "Tag refs/tags/v${agent_version} already exists"
  exit 0
fi

output_steps_yaml "$version" | buildkite-agent pipeline upload
