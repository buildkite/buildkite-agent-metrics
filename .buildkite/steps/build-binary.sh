#!/usr/bin/env sh

set -eu

GOOS=${1:-linux}
GOARCH=${2:-amd64}

export GOOS
export GOARCH
export CGO_ENABLED=0

go build -o "buildkite-agent-metrics-${GOOS}-${GOARCH}" .
buildkite-agent artifact upload "buildkite-agent-metrics-${GOOS}-${GOARCH}"
