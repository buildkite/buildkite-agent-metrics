#!/bin/bash
set -eu

docker run --rm --volume "$PWD:/code" \
  --workdir "/code" \
  --rm \
    golang:1.11 \
    sh -c "go get ./lambda && go build -o ./lambda/handler ./lambda"

chmod +x ./lambda/handler

mkdir -p dist/
zip -j handler.zip lambda/handler

buildkite-agent artifact upload handler.zip
