# Buildkite Metrics

A command-line tool for collecting [Buildkite](https://buildkite.com/) build/job statistics for external metrics systems. Currently [AWS Cloudwatch](http://aws.amazon.com/cloudwatch/) is supported.

[![Build status](https://badge.buildkite.com/80d04fcde3a306bef44e77aadb1f1ffdc20ebb3c8f1f585a60.svg)](https://buildkite.com/buildkite/buildkite-metrics)

## Installing

Either download the latest binary from [buildkite-metrics/buildkite-metrics-Linux-x86_64](https://s3.amazonaws.com/buildkite-metrics/buildkite-metrics-Linux-x86_64) or install with:

```bash
go get github.com/buildkite/buildkite-metrics
```

## Development

You can build and run the binary tool locally with golang installed:

```
go run *.go -org [myorg] -token [buildkite api access token]
```

Currently this will publish metrics to Cloudwatch under the custom metric prefic of `Buildkite`, using AWS credentials from your environment. The machine will require the [`cloudwatch:PutMetricData`](https://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/publishingMetrics.html) IAM permission, and the Buildkite API Access token requires the scopes `read_pipelines`, `read_builds` and `read_agents`.

## Metrics

The following metrics are gathered:

```
Buildkite > RunningBuildsCount
Buildkite > RunningJobsCount
Buildkite > ScheduledBuildsCount
Buildkite > ScheduledJobsCount
Buildkite > UnfinishedJobsCount
Buildkite > IdleAgentsCount
Buildkite > BusyAgentsCount
Buildkite > TotalAgentsCount

Buildkite > (Queue) > RunningBuildsCount
Buildkite > (Queue) > RunningJobsCount
Buildkite > (Queue) > ScheduledBuildsCount
Buildkite > (Queue) > ScheduledJobsCount
Buildkite > (Queue) > UnfinishedJobsCount
Buildkite > (Queue) > IdleAgentsCount
Buildkite > (Queue) > BusyAgentsCount
Buildkite > (Queue) > TotalAgentsCount

Buildkite > (Pipeline) > RunningBuildsCount
Buildkite > (Pipeline) > RunningJobsCount
Buildkite > (Pipeline) > ScheduledBuildsCount
Buildkite > (Pipeline) > ScheduledJobsCount
Buildkite > (Pipeline) > UnfinishedJobsCount
```

## License

See [LICENSE.md](LICENSE.md) (MIT)
