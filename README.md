# Buildkite Metrics

A command-line tool for collecting [Buildkite](https://buildkite.com/) build/job statistics for external metrics systems. Currently [AWS Cloudwatch](http://aws.amazon.com/cloudwatch/) is supported.

## Developing

[![Build status](https://badge.buildkite.com/80d04fcde3a306bef44e77aadb1f1ffdc20ebb3c8f1f585a60.svg)](https://buildkite.com/buildkite-aws-stack/buildkite-metrics)

You can build and run the binary tool locally with golang installed:

```
go run *.go -org [myorg] -token [mytoken]
```

## Metrics

The following metrics are gathered:

```
Buildkite > RunningBuildsCount
Buildkite > RunningJobsCount
Buildkite > ScheduledBuildsCount
Buildkite > ScheduledJobsCount
Buildkite > IdleAgentsCount
Buildkite > BusyAgentsCount
Buildkite > TotalAgentsCount

Buildkite > (Queue) > RunningBuildsCount
Buildkite > (Queue) > RunningJobsCount
Buildkite > (Queue) > ScheduledBuildsCount
Buildkite > (Queue) > ScheduledJobsCount
Buildkite > (Queue) > IdleAgentsCount
Buildkite > (Queue) > BusyAgentsCount
Buildkite > (Queue) > TotalAgentsCount

Buildkite > (Pipeline) > RunningBuildsCount
Buildkite > (Pipeline) > RunningJobsCount
Buildkite > (Pipeline) > ScheduledBuildsCount
Buildkite > (Pipeline) > ScheduledJobsCount
```

## License

See [LICENSE.md](LICENSE.md) (MIT)
