#!/bin/bash
set -eu

export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=$SANDBOX_AWS_ACCESS_KEY_ID
export AWS_SECRET_ACCESS_KEY=$SANDBOX_AWS_SECRET_ACCESS_KEY

mkdir -p build/
buildkite-agent artifact download "build/*" build/

make upload branch=${BUILDKITE_BRANCH}