#!/usr/bin/env bash

set -Eeufo pipefail

source .buildkite/lib/release_dry_run.sh

if [[ "${RELEASE_DRY_RUN:-false}" != "true" && -z "${RELEASE_VERSION_TAG:-}" ]]; then
  echo "Skipping release: RELEASE_VERSION_TAG not set" >&2
  exit 0
fi

registry="public.ecr.aws/buildkite/agent-metrics"
image_tag="${registry}:build-${BUILDKITE_BUILD_NUMBER}"

tag="${RELEASE_VERSION_TAG}"
version="${tag#v}"

echo --- Downloading binaries

rm -rf dist
mkdir -p dist
buildkite-agent artifact download --build "${BUILDKITE_TRIGGERED_FROM_BUILD_ID}" "dist/buildkite-agent-metrics-linux-*" ./dist

echo --- :ecr: Building and Pushing to ECR

# Convert 2.3.2 into [ 2.3.2 2.3 2 ] or 3.0-beta.42 in [ 3.0-beta.42 3.0 3 ]
parse_version() {
  local v="$1"
  IFS='.' read -r -a parts <<< "${v%-*}"

  for idx in $(seq 1 ${#parts[*]}) ; do
    sed -e 's/ /./g' <<< "${parts[@]:0:$idx}"
  done

  [[ "${v%-*}" == "${v}" ]] || echo "${v}"
}

version_tags=($(parse_version "${version}"))

# Pushing to the docker registry in this way greatly simplifies creating the
# manifest list on the docker registry so that either architecture can be pulled
# with the same tag.
release_dry_run docker buildx build \
    --progress plain \
    --tag "${registry}:latest" \
    "${version_tags[@]/#/--tag=${registry}:v}" \
    --platform linux/amd64,linux/arm64 \
    --file .buildkite/Dockerfile.public \
    --push \
    .
