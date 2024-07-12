# Release Instructions

1. Determine the SemVer version for the new release: `$VERSION`. Note: [semver does not include a `v` prefix](https://github.com/semver/semver/blob/master/semver.md#is-v123-a-semantic-version).
1. Generate a changelog using [ghch](https://github.com/buildkite/ghch) `ghch --format=markdown --next-version=v$VERSION`.
1. Create a new branch pointing to the trunk `git fetch origin && git switch -c release/$VERSION origin/master`.
1. Update [CHANGELOG.md](CHANEGLOG.md) with the generated changelog.
1. Update [`version/version.go`](version/version.go) with the new version.
1. Push your branch and wait for CI to pass.
1. Merge your branch.
1. Switch back to the trunk branch: `git switch master`.
1. Pull the trunk branch with tags, and ensure it is pointing to the commit you want to release.
1. Tag the release with the new version: `git tag -sm v$VERSION v$VERSION`.
1. Push the tag to GitHub: `git push --tags`.
1. A tag build will commence on the [pipeline](https://buildkite.com/buildkite/buildkite-agent-metrics). Wait for and unblock the release steps.
1. Check that the [draft release](https://github.com/buildkite/buildkite-agent-metrics/releases) that the build creates is acceptable and release it. Note: the job log for release step on the tag build will contain a link to the draft release.
