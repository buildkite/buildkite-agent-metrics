#!/usr/bin/env sh
set -eu

CGO_ENABLED=0 go build -tags lambda.norpc -o lambda/bootstrap ./lambda

chmod +x lambda/bootstrap

mkdir -p dist
zip -j dist/handler.zip lambda/bootstrap
