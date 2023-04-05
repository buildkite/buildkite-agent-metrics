# Release Instructions

1. Determine the SemVer version for the new release: `$VERSION`.
1. Generate a changelog using [ghch](https://github.com/buildkite/ghch) `ghch --format=markdown --next-version=$VERSION`.
1. Create a new branch `git fetch origin && git switch -c release/$VERSION`.
1. Update [CHANGELOG.md](CHANEGLOG.md) with the generated changelog.
1. Update [`version/version.go`](version/version.go) with the new version.
1. Push your branch and wait for CI to pass.
1. Merge your branch.
1. Pull the master branch with tags, and ensure it pointing to the commit you want to release.
1. Tag the release with the new version: `git tag -sm v$VERSION v$VERSION`.
1. Push the tag to github: `git push --tags`.
1. A tag build will commence on the [pipeline](https://buildkite.com/buildkite/buildkite-agent-metrics).
1. Check that the [draft release](https://github.com/buildkite/buildkite-agent-metrics/releases) that the build creates is acceptable and release it.
