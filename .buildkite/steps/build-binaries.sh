#!/bin/bash
set -eu

export GOARCH=amd64

version=$(awk -F\" '/const Version/ {print $2}' version/version.go)
name="buildkite-agent-metrics"
go_version="1.9.2"
go_pkg="github.com/buildkite/buildkite-agent-metrics"

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

for os in "darwin" "linux"
do
    export GOOS=${os}
    echo "+++ Building ${go_pkg} for $GOOS/$GOARCH with golang:${go_version} :golang:"

    dist_file="dist/${name}-v${version}-${GOOS}-${GOARCH}"
    go_build_in_docker -a -tags netgo -ldflags="-X version.BuildNumber=${BUILDKITE_BUILD_NUMBER} -w" -o "${dist_file}" main.go lambda.go
    file "${dist_file}"

    chmod +x "${dist_file}"
    echo "👍 ${dist_file}"

    buildkite-agent artifact upload "${dist_file}"
done

ln  "dist/${name}-v${version}-linux-amd64" dist/handler.handle

dist_file="dist/buildkite-agent-metrics-v${version}-lambda.zip"
zip -j "$dist_file" dist/handler.handle
rm dist/handle.handle

buildkite-agent artifact upload "$dist_file"
