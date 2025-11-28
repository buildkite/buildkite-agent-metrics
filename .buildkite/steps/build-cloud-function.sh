#!/usr/bin/env bash
set -eu

# Create a zip file containing just the code required for the cloud function

mkdir -p dist
rm -f dist/buildkite-agent-metrics.zip
tar -a -cf dist/buildkite-agent-metrics.zip cloud_function/{main.go,go.mod,go.sum}
