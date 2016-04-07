#!/bin/bash
buildkite-agent artifact download "build/*" build/
make upload branch=${BUILDKITE_BRANCH}