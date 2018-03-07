# Buildkite Metrics

A command-line tool for collecting [Buildkite](https://buildkite.com/) build/job statistics for external metrics systems. Currently [AWS Cloudwatch](http://aws.amazon.com/cloudwatch/) and [StatsD](https://github.com/etsy/statsd) are supported.

[![Build status](https://badge.buildkite.com/80d04fcde3a306bef44e77aadb1f1ffdc20ebb3c8f1f585a60.svg)](https://buildkite.com/buildkite/buildkite-metrics)

## Installing

Either download the latest binary from [Github Releases](https://github.com/buildkite/buildkite-metrics/releases) or install with:

```bash
go get github.com/buildkite/buildkite-metrics
```

### Backends

By default metrics will be submitted to CloudWatch but the backend can be switched to StatsD using the command-line argument `-backend statsd`. The StatsD backend supports the following arguments

* `-statsd-host HOST`: The StatsD host and port (defaults to `127.0.0.1:8125`).
* `-statsd-tags`: Some StatsD servers like the agent provided by DataDog support tags. If specified, metrics will be tagged by `queue` and `pipeline` otherwise metrics will include the queue/pipeline name in the metric. Only enable this option if you know your StatsD server supports tags.

## Development

You can build and run the binary tool locally with golang installed:

```
go run *.go -org [myorg] -token [buildkite api access token]
```

Currently this will publish metrics to Cloudwatch under the custom metric prefix of `Buildkite`, using AWS credentials from your environment. The machine will require the [`cloudwatch:PutMetricData`](https://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/publishingMetrics.html) IAM permission, and the Buildkite API Access token requires the scopes `read_pipelines`, `read_builds` and `read_agents`.

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

## AWS Lambda

An AWS Lambda bundle is created and published as part of the build process.

It's entrypoint is `handler.handle`, it requires a `python2.7` environment and makes use of the following env vars:

 - BUILDKITE_ORG
 - BUILDKITE_TOKEN
 - BUILDKITE_BACKEND
 - BUILDKITE_QUEUE
 - BUILDKITE_QUIET

Checkout https://github.com/buildkite/elastic-ci-stack-for-aws/blob/master/templates/metrics.yml for examples of usage.

## License

See [LICENSE.md](LICENSE.md) (MIT)
