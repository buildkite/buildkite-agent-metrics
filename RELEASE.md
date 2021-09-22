# Release Instructions

1. Determine the SemVer version for the new release
1. Generate a changelog using [ghch](https://github.com/buildkite/ghch) `~/go/bin/ghch --format=markdown --next-version=$VERSION`
1. Create a new branch `git fetch origin && git checkout -b keithdunan/release/$VERSION`
1. Update [CHANGELOG.md](CHANEGLOG.md) with the generated changelog
1. Update [`version/version.go`](version/version.go) with the new version
1. Push your branch and wait for CI to pass
1. Merge your branch, the Buildkite Pipeline will upload the release steps to create a GitHub Release and Git tag