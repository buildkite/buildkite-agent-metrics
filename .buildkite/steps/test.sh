#!/bin/bash
set -eu

go_version="1.9.2"
go_pkg="github.com/buildkite/buildkite-metrics"

  docker run \
    -v "${PWD}:/go/src/${go_pkg}" \
    -w "/go/src/${go_pkg}" \
    --rm "golang:${go_version}" \
    go test -v ./... 
