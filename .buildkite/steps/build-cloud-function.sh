#!/usr/bin/env sh
set -eu

# Create a zip file containing just the code required for the cloud function

mkdir -p dist
rm -f dist/buildkite-agent-metrics.zip
( cd cloud_function && zip ../dist/buildkite-agent-metrics.zip main.go go.mod go.sum )
