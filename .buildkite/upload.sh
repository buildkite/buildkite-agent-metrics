#!/bin/bash
set -eu

export AWS_DEFAULT_REGION=us-east-1
export AWS_ACCESS_KEY_ID=$SANDBOX_AWS_ACCESS_KEY_ID
export AWS_SECRET_ACCESS_KEY=$SANDBOX_AWS_SECRET_ACCESS_KEY

echo "~~~ :buildkite: Downloading artifacts"
mkdir -p build/
buildkite-agent artifact download "build/*" build/

echo "~~~ :s3: Uploading files"
make upload branch=${BUILDKITE_BRANCH}

echo "+++ :s3: Upload complete"
for f in $(ls build/) ; do
  printf "https://buildkite-metrics.s3.amazonaws.com/%s\n" $f
done