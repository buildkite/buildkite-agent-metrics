#!/bin/bash
set -eu

export AWS_DEFAULT_REGION=us-east-1
export AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID:-$SANDBOX_AWS_ACCESS_KEY_ID}
export AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY:-$SANDBOX_AWS_SECRET_ACCESS_KEY}

EXTRA_REGIONS=(
	us-east-2
	us-west-1
	us-west-2
	eu-west-1
	eu-west-2
	eu-central-1
	ap-northeast-1
	ap-northeast-2
	ap-southeast-1
	ap-southeast-2
	ap-south-1
	sa-east-1
)

BASE_BUCKET=buildkite-metrics

echo "~~~ :buildkite: Downloading artifacts"
mkdir -p dist/
buildkite-agent artifact download "dist/*" dist/

ls -al dist/

# echo "+++ :s3: Uploading files to ${BASE_BUCKET} in ${AWS_DEFAULT_REGION}"
# aws s3 sync --acl public-read ./dist "s3://${BASE_BUCKET}/"
# for f in build/* ;
# 	do echo "https://s3.amazonaws.com/bucket/$f"
# done

# for region in "${EXTRA_REGIONS[@]}" ; do
# 	bucket="${BASE_BUCKET}-${region}"
# 	echo "+++ :s3: Copying files to ${bucket}"
# 	if ! aws s3api head-bucket --bucket "${bucket}" --region "${region}" &> /dev/null ; then
# 		echo "Creating s3://${bucket}/"
# 		aws s3 mb "s3://${bucket}/" --region "${region}"
# 	fi
# 	aws --region "${region}" s3 sync --exclude "*" --include "*.zip" --delete --acl public-read "s3://${BASE_BUCKET}/" "s3://${bucket}/"
# 	for f in build/* ; do
# 		echo "https://${bucket}.s3-${region}.amazonaws.com/$f"
# 	done
# done

# echo "+++ :s3: Uploading binary to s3://${BASE_BUCKET}"
# aws s3 cp --acl public-read build/buildkite-metrics-Linux-x86_64* "s3://${BASE_BUCKET}/buildkite-metrics-Linux-x86_64"
