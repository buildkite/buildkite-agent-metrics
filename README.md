# Buildkite Agent Metrics

A command-line tool for collecting [Buildkite](https://buildkite.com/) agent metrics, focusing on enabling auto-scaling. Currently [AWS Cloudwatch](http://aws.amazon.com/cloudwatch/), [StatsD](https://github.com/etsy/statsd) and [Prometheus](https://prometheus.io) are supported.

[![Build status](https://badge.buildkite.com/80d04fcde3a306bef44e77aadb1f1ffdc20ebb3c8f1f585a60.svg)](https://buildkite.com/buildkite/buildkite-agent-metrics)

**Note: Formerly known as `buildkite-metrics`, but now `buildkite-agent-metrics` to reflect the focus of the tool.**

## Installing

Either download the latest binary from [Github Releases](https://github.com/buildkite/buildkite-agent-metrics/releases) or install with:

```bash
go get github.com/buildkite/buildkite-agent-metrics
```

## Running

Several running modes are supported. All of them require an Agent Registration Token, found on the [Buildkite Agents page](https://buildkite.com/organizations/-/agents).

### Running as a Daemon

The simplest deployment is to run as a long-running daemon that collects metrics across all queues in an organization.

```
buildkite-agent-metrics -token abc123 -interval 30s
```

Restrict it to a single queue with `-queue` if you're scaling a single cluster of agents:

```
buildkite-agent-metrics -token abc123 -interval 30s -queue my-queue
```

### Running as an AWS Lambda

An AWS Lambda bundle is created and published as part of the build process. The lambda will require the [`cloudwatch:PutMetricData`](https://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/publishingMetrics.html) IAM permission.

It's entrypoint is `handler`, it requires a `go1.x` environment and respects the following env vars:

 - BUILDKITE_AGENT_TOKEN
 - BUILDKITE_BACKEND
 - BUILDKITE_QUEUE
 - BUILDKITE_QUIET
 - BUILDKITE_CLOUDWATCH_DIMENSIONS

```bash
aws lambda create-function \
  --function-name buildkite-agent-metrics \
  --memory 128 \
  --role arn:aws:iam::account-id:role/execution_role \
  --runtime go1.x \
  --zip-file fileb://handler.zip \
  --handler handler
```

### Running as a Container

You can build a docker image for the `buildkite-agent-metrics` following:

```
docker build -t buildkite-agent-metrics .
```

This will create a local docker image named as `buildkite-agent-metrics` that you can tag and push to your own registry.

You can use the command-line arguments in a docker execution in the same way as described before:

```
docker run --rm buildkite-agent-metrics -token abc123 -interval 30s -queue my-queue
```

### Backends

By default metrics will be submitted to CloudWatch but the backend can be switched to StatsD or Prometheus using the command-line argument `-backend statsd` or `-backend prometheus` respectively.

The Cloudwatch backend supports the following arguments:

* `-cloudwatch-dimensions`: A optional custom dimension in the form of `Key=Value, Key=Value`

The StatsD backend supports the following arguments:

* `-statsd-host HOST`: The StatsD host and port (defaults to `127.0.0.1:8125`).
* `-statsd-tags`: Some StatsD servers like the agent provided by Datadog support tags. If specified, metrics will be tagged by `queue` otherwise metrics will include the queue name in the metric. Only enable this option if you know your StatsD server supports tags.

The Prometheus backend supports the following arguments

* `-prometheus-addr`: The local address to listen on (defaults to `:8080`).
* `-prometheus-path`: The path under `prometheus-addr` to expose metrics on (defaults to `/metrics`).

### Upgrading from v2 to v3

1. The `-org` argument is no longer needed
2. The `-token` argument is now an _Agent Registration Token_ — the same used in the Buildkite Agent configuration file, and found on the [Buildkite Agents page](https://buildkite.com/organizations/-/agents).
3. Build and pipeline metrics have been removed, focusing on agents and jobs by queue for auto–scaling.
   If you have a compelling reason to gather build or pipeline metrics please continue to use the [previous version](https://github.com/buildkite/buildkite-agent-metrics/releases/tag/v2.1.0) or [open an issue](https://github.com/buildkite/buildkite-agent-metrics/issues) with details.

## Development

This tool is built with Go 1.11+ and [Go Modules](https://github.com/golang/go/wiki/Modules).

You can build and run the binary tool locally with golang installed:

```
export GO111MODULE=on
go run *.go -token [buildkite agent registration token]
```

Currently this will publish metrics to Cloudwatch under the custom metric prefix of `Buildkite`, using AWS credentials from your environment. The machine will require the [`cloudwatch:PutMetricData`](https://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/publishingMetrics.html) IAM permission.

## Metrics

The following metrics are gathered when no specific queue is supplied:

```
Buildkite > (Org) > RunningJobsCount
Buildkite > (Org) > ScheduledJobsCount
Buildkite > (Org) > UnfinishedJobsCount
Buildkite > (Org) > IdleAgentsCount
Buildkite > (Org) > BusyAgentsCount
Buildkite > (Org) > TotalAgentsCount

Buildkite > (Org, Queue) > RunningJobsCount
Buildkite > (Org, Queue) > ScheduledJobsCount
Buildkite > (Org, Queue) > UnfinishedJobsCount
Buildkite > (Org, Queue) > IdleAgentsCount
Buildkite > (Org, Queue) > BusyAgentsCount
Buildkite > (Org, Queue) > TotalAgentsCount
```

When a queue is specified, only that queue's metrics are published.

## License

See [LICENSE.md](LICENSE.md) (MIT)
