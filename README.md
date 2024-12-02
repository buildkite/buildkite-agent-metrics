# Buildkite Agent Metrics

A command-line tool for collecting [Buildkite](https://buildkite.com/) agent metrics, focusing on enabling auto-scaling. Currently [AWS Cloudwatch](http://aws.amazon.com/cloudwatch/), [StatsD](https://github.com/etsy/statsd), [Prometheus](https://prometheus.io), [Stackdriver](https://cloud.google.com/stackdriver/) and [New Relic](https://newrelic.com/products/insights) are supported.

[![Build status](https://badge.buildkite.com/3642e233e25707a91db3a9e7d61a6fd46e1352c161605b56cf.svg)](https://buildkite.com/buildkite/buildkite-agent-metrics)

## Installing

Either download the latest binary from
[Github Releases](https://github.com/buildkite/buildkite-agent-metrics/releases) or install with:

```bash
go install github.com/buildkite/buildkite-agent-metrics/v5@latest
```

## Running

Several running modes are supported. All of them require an Agent Registration
Token, found on the
[Buildkite Agents page](https://buildkite.com/organizations/-/agents).

### Running as a Daemon

The simplest deployment is to run as a long-running daemon that collects metrics
across all queues in an organization.

```shell
buildkite-agent-metrics -token abc123 -interval 30s
```

Restrict it to a single queue with `-queue`:

```shell
buildkite-agent-metrics -token abc123 -interval 30s -queue my-queue
```

Restrict it to multiple queues by repeating `-queue`:

```shell
buildkite-agent-metrics -token abc123 -interval 30s -queue my-queue1 -queue my-queue2
```

When using clusters, you can pass a cluster registration token to gather metrics
only for that cluster:

```shell
buildkite-agent-metrics -token clustertoken ...
```

You can repeat `-token` to gather metrics for multiple clusters:

```shell
buildkite-agent-metrics -token clusterAtoken -token clusterBtoken ...
```

### Running as an AWS Lambda

An AWS Lambda bundle is created and published as part of the build process. The
lambda will require the
[`cloudwatch:PutMetricData`](https://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/publishingMetrics.html)
IAM permission.

It requires a `provided.al2` environment and respects the following env vars:

- `BUILDKITE_BACKEND` : The name of the backend to use (e.g. `cloudwatch`,
   `statsd`, `newrelic`. For the lambda, `prometheus` and `stackdriver` are not
   supported).
- `BUILDKITE_QUEUE` : A comma separated list of Buildkite queues to process
  (e.g. `backend-deploy,ui-deploy`).
- `BUILDKITE_QUIET` : A boolean specifying that only `ERROR` log lines must be
   printed. This accepts either `1` or `true` to enable.
- `BUILDKITE_CLOUDWATCH_DIMENSIONS` : A comma separated list in the form of
   `Key=Value,Other=Value` containing the Cloudwatch dimensions to index metrics
   under.
 - `BUILDKITE_CLOUDWATCH_HIGH_RESOLUTION` : Whether to enable [High-Resolution Metrics](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/publishingMetrics.html#high-resolution-metrics) which incurs additional charges. This accepts either `1` or `true` to enable.

To adjust timeouts, and connection pooling in the HTTP client use the following env vars:

- `BUILDKITE_AGENT_METRICS_TIMEOUT` : Timeout, in seconds, TLS handshake and idle connections, for HTTP requests, to Buildkite API (default 15).
- `BUILDKITE_AGENT_METRICS_MAX_IDLE_CONNS` : Maximum number of idle (keep-alive) HTTP connections 
   for Buildkite Agent API. Zero means no limit, -1 disables pooling (default 100).

To assist with debugging the following env vars are provided:

- `BUILDKITE_AGENT_METRICS_DEBUG` : A boolean which enables debug logging. This accepts either `1` or `true` to enable.
- `BUILDKITE_AGENT_METRICS_DEBUG_HTTP` : A boolean which enables printing of the HTTP responses. This accepts either `1` or `true` to enable.

Additionally, one of the following groups of environment variables must be set
in order to define how the Lambda function should obtain the required Buildkite
Agent API token:

#### Option 1 - Provide the token(s) as plain-text

- `BUILDKITE_AGENT_TOKEN` : The Buildkite Agent API token to use. You can supply
  multiple tokens comma-separated.

#### Option 2 - Retrieve token from AWS Systems Manager

- `BUILDKITE_AGENT_TOKEN_SSM_KEY` : The parameter name which contains the token
  value in AWS Systems Manager. You can supply multiple names comma-separated.

**Note**: Parameters stored as `String` and `SecureString` are currently
supported.

#### Option 3 - Retrieve token from AWS Secrets Manager

- `BUILDKITE_AGENT_SECRETS_MANAGER_SECRET_ID`: The id of the secret which
  contains the token value in AWS Secrets Manager. You can supply
  multiple ids comma-separated.
- (Optional) `BUILDKITE_AGENT_SECRETS_MANAGER_JSON_KEY`: The JSON key containing
  the token value in the secret JSON blob.

**Note 1**: Both `SecretBinary` and `SecretString` are supported. In the case of
`SecretBinary`, the secret payload will be automatically decoded and returned as
a plain-text string.

**Note 2**: `BUILDKITE_AGENT_SECRETS_MANAGER_JSON_KEY` can be used on secrets of
type `SecretBinary` only if their binary payload corresponds to a valid JSON
object containing the provided key.

```bash
aws lambda create-function \
  --function-name buildkite-agent-metrics \
  --memory 128 \
  --role arn:aws:iam::account-id:role/execution_role \
  --runtime provided.al2 \
  --zip-file fileb://handler.zip \
  --handler handler
```

### Running as a Container

You can build a docker image for the `buildkite-agent-metrics` following:

```shell
docker build -t buildkite-agent-metrics .
```

This will create a local docker image named as `buildkite-agent-metrics` that
you can tag and push to your own registry.

You can use the command-line arguments in a docker execution in the same way as
described before:

```shell
docker run --rm buildkite-agent-metrics -token abc123 -interval 30s -queue my-queue
```

### Supported command line flags

```shell
$ buildkite-agent-metrics --help
Usage of buildkite-agent-metrics:
  -backend string
        Specify the backend to use: cloudwatch, newrelic, prometheus, stackdriver, statsd (default "cloudwatch")
  -cloudwatch-dimensions string
        Cloudwatch dimensions to index metrics under, in the form of Key=Value, Other=Value
  -cloudwatch-region string
        AWS Region to connect to, defaults to $AWS_REGION or us-east-1
  -debug
        Show debug output
  -debug-http
        Show full http traces
  -dry-run
        Whether to only print metrics
  -endpoint string
        A custom Buildkite Agent API endpoint (default "https://agent.buildkite.com/v3")
  -interval duration
    	  Update metrics every interval, rather than once
  -cloudwatch-high-resolution
        If `-interval` is less than 60 seconds send metrics to CloudWatch as [High-Resolution Metrics](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/publishingMetrics.html#high-resolution-metrics) which incurs additional charges.
  -max-idle-conns int
        Maximum number of idle (keep-alive) HTTP connections for Buildkite Agent API. Zero means no limit, -1 disables connection reuse. (default 100)
  -newrelic-app-name string
        New Relic application name for metric events
  -newrelic-license-key string
        New Relic license key for publishing events
  -prometheus-addr string
        Prometheus metrics transport bind address (default ":8080")
  -prometheus-path string
        Prometheus metrics transport path (default "/metrics")
  -queue value
        Specific queues to process
  -quiet
        Only print errors
  -stackdriver-projectid string
        Specify Stackdriver Project ID
  -statsd-host string
        Specify the StatsD server (default "127.0.0.1:8125")
  -statsd-tags
        Whether your StatsD server supports tagging like Datadog
  -timeout int
        Timeout, in seconds, for HTTP requests to Buildkite API (default 15)
  -token value
        Buildkite Agent registration tokens. At least one is required. Multiple cluster tokens can be used to gather metrics for multiple clusters.
  -version
        Show the version
```

### Backends

By default metrics will be submitted to CloudWatch but the backend can be switched to StatsD or Prometheus using the command-line argument `-backend statsd` or `-backend prometheus` respectively.

#### Cloudwatch

The Cloudwatch backend supports the following arguments:

- `-cloudwatch-dimensions`: A optional custom dimension in the form of `Key=Value, Key=Value`

#### StatsD (Datadog)

The StatsD backend supports the following arguments:

- `-statsd-host HOST`: The StatsD host and port (defaults to `127.0.0.1:8125`).
- `-statsd-tags`: Some StatsD servers like the agent provided by Datadog support
   tags. If specified, metrics will be tagged by `queue` otherwise metrics will
   include the queue name in the metric. Only enable this option if you know
   your StatsD server supports tags.

#### Prometheus

The Prometheus backend supports the following arguments:

- `-prometheus-addr`: The local address to listen on (defaults to `:8080`).
- `-prometheus-path`: The path under `prometheus-addr` to expose metrics on
   (defaults to `/metrics`).

#### Stackdriver

The Stackdriver backend supports the following arguments:

- `-stackdriver-projectid`: The Google Cloud Platform project to report metrics
   for.

The New Relic backend supports the following arguments:

- `-newrelic-app-name`: String for the New Relic app name
- `-newrelic-license-key`: The New Relic license key. Must be of type `INGEST`

### Upgrading from v2 to v3

1. The `-org` argument is no longer needed
2. The `-token` argument is now an _Agent Registration Token_ — the same used in
   the Buildkite Agent configuration file, and found on the
   [Buildkite Agents page](https://buildkite.com/organizations/-/agents).
3. Build and pipeline metrics have been removed, focusing on agents and jobs by
   queue for auto–scaling.
   If you have a compelling reason to gather build or pipeline metrics please
   continue to use the
   [previous version](https://github.com/buildkite/buildkite-agent-metrics/releases/tag/v2.1.0)
   or [open an issue](https://github.com/buildkite/buildkite-agent-metrics/issues)
   with details.

## Development

This tool is built with Go 1.20+ and assumes
[Go Modules](https://github.com/golang/go/wiki/Modules) by default.

You can build and run the binary tool locally with Go installed:

```shell
go run *.go -token [buildkite agent registration token]
```

Currently this will publish metrics to Cloudwatch under the custom metric prefix
of `Buildkite`, using AWS credentials from your environment. The machine will
require the
[`cloudwatch:PutMetricData`](https://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/publishingMetrics.html)
IAM permission.

### The `token` package

It is an abstraction layer enabling the retrieval of a Buildkite Agent API token
from different sources.

The current supported sources are:

- AWS Systems Manager (a.k.a parameter store).
- AWS Secrets Manager.
- OS environment variable.

#### Tests

All the tests for AWS dependant resources require their corresponding auto-generated mocks. Thus,
before running them, you need to generate such mocks by executing:

```bash
go generate token/secretsmanager_test.go
go generate token/ssm_test.go
```

## Metrics

The following metrics are gathered when no specific queue is supplied:

```plain
Buildkite > (Org) > RunningJobsCount
Buildkite > (Org) > ScheduledJobsCount
Buildkite > (Org) > UnfinishedJobsCount
Buildkite > (Org) > WaitingJobsCount
Buildkite > (Org) > IdleAgentsCount
Buildkite > (Org) > BusyAgentsCount
Buildkite > (Org) > BusyAgentPercentage
Buildkite > (Org) > TotalAgentsCount

Buildkite > (Org, Queue) > RunningJobsCount
Buildkite > (Org, Queue) > ScheduledJobsCount
Buildkite > (Org, Queue) > UnfinishedJobsCount
Buildkite > (Org, Queue) > WaitingJobsCount
Buildkite > (Org, Queue) > IdleAgentsCount
Buildkite > (Org, Queue) > BusyAgentsCount
Buildkite > (Org, Queue) > BusyAgentPercentage
Buildkite > (Org, Queue) > TotalAgentsCount
```

When a queue is specified, only that queue's metrics are published.

We send metrics for Jobs in the following states:

- **Scheduled**: the job hasn't been assigned to an agent yet. If you have agent
  capacity, this value should be close to 0.
- **Waiting**: the job is known to exist but isn't schedulable yet due to
  dependencies, `wait` statements, etc. This information is mostly useful to an
  autoscaler, since it represents work that will start soon.
- **Running**: an agent is actively executing this job.

## License

See [LICENSE.md](LICENSE.md) (MIT)
