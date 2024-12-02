#!/usr/bin/env bash

set -Eeufo pipefail

source .buildkite/lib/release_dry_run.sh

if [[ "${RELEASE_DRY_RUN:-false}" != "true" && "${BUILDKITE_BRANCH}" != "${BUILDKITE_TAG:-}" ]]; then
  echo "Skipping release for a non-tag build on ${BUILDKITE_BRANCH}" >&2
  exit 0
fi

registry="public.ecr.aws/buildkite/agent-metrics"
image_tag="${registry}:build-${BUILDKITE_BUILD_NUMBER}"

echo --- Fetching tags
git fetch --prune --force origin "+refs/tags/*:refs/tags/*"

echo --- Checking tags
version="$(awk -F\" '/const Version/ {print $2}' version/version.go)"
tag="v${version#v}"

if [[ "${RELEASE_DRY_RUN:-false}" != true && "${tag}" != "${BUILDKITE_TAG}" ]]; then
  echo "Error: version.go has not been updated to ${BUILDKITE_TAG#v}"
  exit 1
fi

echo --- Downloading binaries

artifacts_build="$(buildkite-agent meta-data get "metrics-artifacts-build")"

rm -rf dist
mkdir -p dist
buildkite-agent artifact download --build "${artifacts_build}" "dist/buildkite-agent-metrics-linux-*" ./dist

echo "--- Building :docker: image"

builder_name="$(docker buildx create --use)"
# shellcheck disable=SC2064 # we want the current $builder_name to be trapped, not the runtime one
trap "docker buildx rm ${builder_name} || true" EXIT

echo "--- Building :docker: ${image_tag} for all architectures"
docker buildx build \
  --progress plain \
  --builder "${builder_name}" \
  --platform linux/amd64,linux/arm64 \
  --file .buildkite/Dockerfile.public \
  .

# Tag images for just the native architecture. There is a limitation in docker that prevents this
# from being done in one command. Luckliy the second build will be quick because of docker layer caching
# As this is just a native build, we don't need the lock.
docker buildx build \
  --progress plain \
  --builder "${builder_name}" \
  --tag "${image_tag}" \
  --file .buildkite/Dockerfile.public \
  --load \
  .

echo --- :ecr: Pushing to ECR

# Convert 2.3.2 into [ 2.3.2 2.3 2 ] or 3.0-beta.42 in [ 3.0-beta.42 3.0 3 ]
parse_version() {
  local v="$1"
  IFS='.' read -r -a parts <<< "${v%-*}"

  for idx in $(seq 1 ${#parts[*]}) ; do
    sed -e 's/ /./g' <<< "${parts[@]:0:$idx}"
  done

  [[ "${v%-*}" == "${v}" ]] || echo "${v}"
}

version_tags=($(parse_version "${version#v}"))

# Do another build with all architectures and tags. The layers should be cached
# from the previous build with all architectures.
# Pushing to the docker registry in this way greatly simplifies creating the
# manifest list on the docker registry so that either architecture can be pulled
# with the same tag.
release_dry_run docker buildx build \
    --progress plain \
    --builder "${builder_name}" \
    --tag "${registry}:latest" \
    "${version_tags[@]/#/--tag ${registry}:v}" \
    --platform linux/amd64,linux/arm64 \
    --file .buildkite/Dockerfile.public \
    --push \
    .
