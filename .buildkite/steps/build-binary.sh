#!/bin/bash
set -eu

export GOOS=$1
export GOARCH=$2

version=$(awk -F\" '/const Version/ {print $2}' version/version.go)
dist_file="dist/buildkite-metrics-v${version}-${GOOS}-${GOARCH}"

go_version="1.9.2"
go_pkg="github.com/buildkite/buildkite-metrics"

rm -rf dist
mkdir -p dist

go_build_in_docker() {
  docker run \
    -v "${PWD}:/go/src/${go_pkg}" \
    -w "/go/src/${go_pkg}" \
    -e "GOOS=${GOOS}" -e "GOARCH=${GOARCH}" -e "CGO_ENABLED=0" \
    --rm "golang:${go_version}" \
    go build "$@"
}

echo "+++ Building ${go_pkg} for $GOOS/$GOARCH with golang:${go_version} :golang:"

go_build_in_docker -a -tags netgo -ldflags="-X version.BuildNumber=${BUILDKITE_BUILD_NUMBER} -w" -o "${dist_file}" main.go
file "${dist_file}"

chmod +x "${dist_file}"
echo "👍 ${dist_file}"

buildkite-agent artifact upload "${dist_file}"
