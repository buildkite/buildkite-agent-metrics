#!/bin/bash
set -eu

go_pkg="github.com/buildkite/buildkite-metrics"
go_src_dir="/go/src/${go_pkg}"
version=$(awk -F\" '/const Version/ {print $2}' metrics/version.go)-${BUILDKITE_BUILD_NUMBER}

docker run --rm -v "${PWD}:${go_src_dir}" -w "${go_src_dir}" eawsy/aws-lambda-go
mkdir -p build/
mv handler.zip "build/buildkite-metrics-${version}-lambda.zip"

buildkite-agent artifact upload "${DISTFILE}"
