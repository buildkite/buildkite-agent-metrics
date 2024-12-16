#!/usr/bin/env sh

set -eu

GOOS=${1:-linux}
GOARCH=${2:-amd64}

export GOOS
export GOARCH
export CGO_ENABLED=0

mkdir -p dist
go build -o "dist/buildkite-agent-metrics-${GOOS}-${GOARCH}" .
