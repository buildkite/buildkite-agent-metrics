# Buildkite Agent Metrics

A command-line tool for collecting [Buildkite](https://buildkite.com/) agent metrics, focusing on enabling auto-scaling. Currently [AWS Cloudwatch](http://aws.amazon.com/cloudwatch/), [StatsD](https://github.com/etsy/statsd) and [Prometheus](https://prometheus.io) are supported.

[![Build status](https://badge.buildkite.com/80d04fcde3a306bef44e77aadb1f1ffdc20ebb3c8f1f585a60.svg)](https://buildkite.com/buildkite/buildkite-metrics)

## Installing

Either download the latest binary from [Github Releases](https://github.com/buildkite/buildkite-metrics/releases) or install with:

```bash
go get github.com/buildkite/buildkite-metrics
```

## Running

Several running modes are supported. All of them require an Agent Registration Token, found on the [Buildkite Agents page](https://buildkite.com/organizations/-/agents).

### Running as a Daemon

The simplest deployment is to run as a long-running daemon that collects metrics across all queues in an organization.

```
buildkite-metrics -token abc123 -interval 30s
```

Restrict it to a single queue with `-queue` if you're scaling a single cluster of agents:

```
buildkite-metrics -token abc123 -interval 30s -queue my-queue
```

### Running as an AWS Lambda

An AWS Lambda bundle is created and published as part of the build process.

It's entrypoint is `handler.handle`, it requires a `python2.7` environment and respects the following env vars:

 - BUILDKITE_TOKEN
 - BUILDKITE_BACKEND
 - BUILDKITE_QUEUE
 - BUILDKITE_QUIET
 - BUILDKITE_CLOUDWATCH_DIMENSIONS

Take a look at https://github.com/buildkite/elastic-ci-stack-for-aws/blob/master/templates/metrics.yml for examples of usage.

### Backends

By default metrics will be submitted to CloudWatch but the backend can be switched to StatsD or Prometheus using the command-line argument `-backend statsd` or `-backend prometheus` respectively.

The Cloudwatch backend supports the following arguments:

* `-cloudwatch-dimensions`: A optional custom dimension in the form of `Key=Value, Key=Value`

The StatsD backend supports the following arguments:

* `-statsd-host HOST`: The StatsD host and port (defaults to `127.0.0.1:8125`).
* `-statsd-tags`: Some StatsD servers like the agent provided by DataDog support tags. If specified, metrics will be tagged by `queue` otherwise metrics will include the queue name in the metric. Only enable this option if you know your StatsD server supports tags.

The Prometheus backend supports the following arguments

* `-prometheus-addr`: The local address to listen on (defaults to `:8080`).
* `-prometheus-path`: The path under `prometheus-addr` to expose metrics on (defaults to `/metrics`).

### Upgrading from v2 to v3

1. The `-org` argument is no longer needed
2. The `-token` argument is now an _Agent Registration Token_ — the same used in the Buildkite Agent configuration file, and found on the [Buildkite Agents page](https://buildkite.com/organizations/-/agents).
3. Build and pipeline metrics have been removed, focusing on agents and jobs by queue for auto–scaling.
   If you have a compelling reason to gather build or pipeline metrics please continue to use the [previous version](https://github.com/buildkite/buildkite-metrics/releases/tag/v2.1.0) or [open an issue](https://github.com/buildkite/buildkite-metrics/issues) with details.

## Development

You can build and run the binary tool locally with golang installed:

```
go run *.go -token [buildkite agent registration token]
```

Currently this will publish metrics to Cloudwatch under the custom metric prefix of `Buildkite`, using AWS credentials from your environment. The machine will require the [`cloudwatch:PutMetricData`](https://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/publishingMetrics.html) IAM permission.

## Metrics

The following metrics are gathered when no specific queue is supplied:

```
Buildkite > RunningJobsCount
Buildkite > ScheduledJobsCount
Buildkite > UnfinishedJobsCount
Buildkite > IdleAgentsCount
Buildkite > BusyAgentsCount
Buildkite > TotalAgentsCount

Buildkite > (Queue) > RunningJobsCount
Buildkite > (Queue) > ScheduledJobsCount
Buildkite > (Queue) > UnfinishedJobsCount
Buildkite > (Queue) > IdleAgentsCount
Buildkite > (Queue) > BusyAgentsCount
Buildkite > (Queue) > TotalAgentsCount
```

When a queue is specified, only that queue's metrics are published.

## License

See [LICENSE.md](LICENSE.md) (MIT)
