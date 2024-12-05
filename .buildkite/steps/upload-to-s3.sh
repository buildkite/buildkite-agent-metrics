#!/bin/bash
set -eu

export AWS_DEFAULT_REGION="us-east-1"

# Updated from https://docs.aws.amazon.com/general/latest/gr/rande.html#lambda_region
EXTRA_REGIONS=(
  us-east-2
  us-west-1
  us-west-2
  af-south-1
  ap-east-1
  ap-south-1
  ap-northeast-2
  ap-northeast-1
  ap-southeast-2
  ap-southeast-1
  ca-central-1
  eu-central-1
  eu-west-1
  eu-west-2
  eu-south-1
  eu-west-3
  eu-north-1
  me-south-1
  sa-east-1
)

VERSION="$(awk -F\" '/const Version/ {print $2}' version/version.go)"
BASE_BUCKET="buildkite-lambdas"
BUCKET_PATH="buildkite-agent-metrics"

if [[ "${1:-}" == "release" ]] ; then
  BUCKET_PATH="${BUCKET_PATH}/v${VERSION}"
else
  BUCKET_PATH="${BUCKET_PATH}/builds/${BUILDKITE_BUILD_NUMBER}"
fi

echo "~~~ :buildkite: Downloading artifacts"

mkdir -p dist
if [[ "${1:-}" == "release" ]] ; then
  buildkite-agent artifact download --build "${BUILDKITE_TRIGGERED_FROM_BUILD_ID}" dist/handler.zip ./dist
else
  buildkite-agent artifact download dist/handler.zip ./dist
fi

echo "--- :s3: Uploading lambda to ${BASE_BUCKET}/${BUCKET_PATH}/ in ${AWS_DEFAULT_REGION}"
aws s3 cp --acl public-read dist/handler.zip "s3://${BASE_BUCKET}/${BUCKET_PATH}/handler.zip"

for region in "${EXTRA_REGIONS[@]}" ; do
	bucket="${BASE_BUCKET}-${region}"
	echo "--- :s3: Copying files to ${bucket}"
	aws --region "${region}" s3 cp --acl public-read "s3://${BASE_BUCKET}/${BUCKET_PATH}/handler.zip" "s3://${bucket}/${BUCKET_PATH}/handler.zip"
done
