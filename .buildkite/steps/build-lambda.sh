#!/bin/bash
set -eu

go_pkg="github.com/buildkite/buildkite-agent-metrics"
go_src_dir="/go/src/${go_pkg}"
version=$(awk -F\" '/const Version/ {print $2}' version/version.go)
dist_file="dist/buildkite-agent-metrics-v${version}-lambda.zip"

docker run --rm --volume "$PWD:${go_src_dir}" \
  --workdir "${go_src_dir}" \
  --rm golang:1.11 \
  --env GO111MODULE=on \
  sh -c "go get ./lambda && go build -o ./lambda/handler ./lambda"

chmod +x ./lambda/handler

mkdir -p dist/
zip -j "$dist_file" lambda/handler

buildkite-agent artifact upload "$dist_file"
