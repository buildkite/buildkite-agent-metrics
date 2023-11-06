#!/bin/bash
set -eu

docker run --rm --volume "$PWD:/code" \
  --workdir "/code" \
  --rm \
  --env CGO_ENABLED=0 \
  golang:1.20 \
    go build -tags lambda.norpc -o lambda/bootstrap ./lambda

chmod +x lambda/bootstrap

mkdir -p dist/
zip -j handler.zip lambda/bootstrap

buildkite-agent artifact upload handler.zip
