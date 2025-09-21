package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/buildkite/buildkite-agent-metrics/v5/backend"
	"github.com/buildkite/buildkite-agent-metrics/v5/collector"
	"github.com/buildkite/buildkite-agent-metrics/v5/version"
)

// Where we send metrics
var metricsBackend backend.Backend

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	var (
		interval     = flag.Duration("interval", 0, "Update metrics every interval, rather than once")
		showVersion  = flag.Bool("version", false, "Show the version")
		quiet        = flag.Bool("quiet", false, "Only print errors")
		debug        = flag.Bool("debug", false, "Show debug output")
		debugHttp    = flag.Bool("debug-http", false, "Show full http traces")
		dryRun       = flag.Bool("dry-run", false, "Whether to only print metrics")
		endpoint     = flag.String("endpoint", "https://agent.buildkite.com/v3", "A custom Buildkite Agent API endpoint")
		timeout      = flag.Int("timeout", 15, "Timeout, in seconds, TLS handshake and idle connections, for HTTP requests, to Buildkite API")
		maxIdleConns = flag.Int("max-idle-conns", 100, "Maximum number of idle (keep-alive) HTTP connections for Buildkite Agent API. Zero means no limit, -1 disables connection reuse.")

		// backend config
		backendOpt        = flag.String("backend", "cloudwatch", "Specify the backend to use: cloudwatch, newrelic, prometheus, stackdriver, statsd, opentelemetry")
		statsdMetricsPrefix = flag.String("statsd-metrics-prefix", "buildkite.", "Specify the StatsD metrics prefix")
		statsdHost        = flag.String("statsd-host", "127.0.0.1:8125", "Specify the StatsD server")
		statsdTags        = flag.Bool("statsd-tags", false, "Whether your StatsD server supports tagging like Datadog")
		prometheusAddr    = flag.String("prometheus-addr", ":8080", "Prometheus metrics transport bind address")
		prometheusPath    = flag.String("prometheus-path", "/metrics", "Prometheus metrics transport path")
		clwRegion         = flag.String("cloudwatch-region", "", "AWS Region to connect to, defaults to $AWS_REGION or us-east-1")
		clwDimensions     = flag.String("cloudwatch-dimensions", "", "Cloudwatch dimensions to index metrics under, in the form of Key=Value, Other=Value")
		clwHighResolution = flag.Bool("cloudwatch-high-resolution", false, "Send metrics at a high-resolution, which incurs extra costs")
		gcpProjectID      = flag.String("stackdriver-projectid", "", "Specify Stackdriver Project ID")
		nrAppName         = flag.String("newrelic-app-name", "", "New Relic application name for metric events")
		nrLicenseKey      = flag.String("newrelic-license-key", "", "New Relic license key for publishing events")
	)

	// custom config for multiple tokens and queues
	var tokens, queues stringSliceFlag
	flag.Var(&tokens, "token", "Buildkite Agent registration tokens. At least one is required. Multiple cluster tokens can be used to gather metrics for multiple clusters.")
	flag.Var(&queues, "queue", "Specific queues to process")

	flag.Parse()

	if *showVersion {
		fmt.Printf("buildkite-agent-metrics %s\n", version.Version)
		os.Exit(0)
	}

	if os.Getenv("BUILDKITE_AGENT_ENDPOINT") != "" {
		*endpoint = os.Getenv("BUILDKITE_AGENT_ENDPOINT")
	}

	if len(tokens) == 0 {
		envTokens := strings.Split(os.Getenv("BUILDKITE_AGENT_TOKEN"), ",")
		for _, t := range envTokens {
			t = strings.TrimSpace(t)
			if t == "" {
				continue
			}
			tokens = append(tokens, t)
		}
	}

	if len(tokens) == 0 {
		fmt.Println("Must provide at least one token with either --token or BUILDKITE_AGENT_TOKEN")
		os.Exit(1)
	}

	var err error
	switch strings.ToLower(*backendOpt) {
	case "cloudwatch":
		region := *clwRegion
		if region == "" {
			region = os.Getenv("AWS_REGION")
		}
		if region == "" {
			region = "us-east-1"
		}
		dimensions, err := backend.ParseCloudWatchDimensions(*clwDimensions)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		metricsBackend = backend.NewCloudWatchBackend(region, dimensions, int64(interval.Seconds()), *clwHighResolution)
	case "statsd":
		metricsBackend, err = backend.NewStatsDBackend(*statsdHost, *statsdTags, *statsdMetricsPrefix)
		if err != nil {
			fmt.Printf("Error starting StatsD, err: %v\n", err)
			os.Exit(1)
		}

	case "prometheus":
		prom := backend.NewPrometheusBackend()
		go prom.Serve(*prometheusPath, *prometheusAddr)
		metricsBackend = prom

	case "stackdriver":
		if *gcpProjectID == "" {
			*gcpProjectID = os.Getenv(`GCP_PROJECT_ID`)
		}
		metricsBackend, err = backend.NewStackDriverBackend(*gcpProjectID)
		if err != nil {
			fmt.Printf("Error starting Stackdriver backend, err: %v\n", err)
			os.Exit(1)
		}

	case "newrelic":
		metricsBackend, err = backend.NewNewRelicBackend(*nrAppName, *nrLicenseKey)
		if err != nil {
			fmt.Printf("Error starting New Relic client: %v\n", err)
			os.Exit(1)
		}

	case "opentelemetry":
		metricsBackend, err = backend.NewOpenTelemetryBackend()
		if err != nil {
			fmt.Printf("Error starting OpenTelemetry backend: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Println("Must provide a supported backend: cloudwatch, newrelic, prometheus, stackdriver, statsd, opentelemetry")
		os.Exit(1)
	}

	if closableMetrics, ok := metricsBackend.(backend.Closer); ok {
		defer func(closer backend.Closer) {
			err := closer.Close()
			log.Println("Closing metrics backend")
			if err != nil {
				fmt.Printf("Error closing metrics backend: %v\n", err)
				os.Exit(1)
			}
		}(closableMetrics)
	}

	if *quiet {
		log.SetOutput(io.Discard)
	}

	userAgent := fmt.Sprintf("buildkite-agent-metrics/%s buildkite-agent-metrics-cli", version.Version)
	if *interval > 0 {
		userAgent += fmt.Sprintf(" interval=%s", *interval)
	}

	// Queues passed as flags take precedence. But if no queues are passed in we
	// check env vars. If no env vars are defined we default to ingesting metrics
	// for all queues.
	// NOTE: `BUILDKITE_QUEUE` is a comma separated string of queues
	// i.e. "default,deploy,test"
	if len(queues) == 0 {
		if q, exists := os.LookupEnv(`BUILDKITE_QUEUE`); exists {
			queues = strings.Split(q, ",")
		}
	}

	httpClient := collector.NewHTTPClient(*timeout, *maxIdleConns)

	collectors := make([]*collector.Collector, 0, len(tokens))
	for _, token := range tokens {
		collectors = append(collectors, &collector.Collector{
			Client:    httpClient,
			UserAgent: userAgent,
			Endpoint:  *endpoint,
			Token:     token,
			Queues:    []string(queues),
			Quiet:     *quiet,
			Debug:     *debug,
			DebugHttp: *debugHttp,
		})
	}

	collectFunc := func() (time.Duration, error) {
		start := time.Now()

		// minimum result.PollDuration across collectors
		var pollDuration time.Duration

		for _, c := range collectors {
			result, err := c.Collect()
			if err != nil {
				fmt.Printf("Error collecting agent metrics, err: %s\n", err)
				if errors.Is(err, collector.ErrUnauthorized) {
					// Unique exit code to signal HTTP 401
					os.Exit(4)
				}
				return time.Duration(0), err
			}

			if *dryRun {
				continue
			}

			if err := metricsBackend.Collect(result); err != nil {
				return time.Duration(0), err
			}
			if result.PollDuration > pollDuration {
				pollDuration = result.PollDuration
			}
		}

		collectionDuration := time.Since(start)
		log.Printf("Finished in %s", collectionDuration)

		return pollDuration, nil
	}

	minPollDuration, err := collectFunc()
	if err != nil {
		fmt.Println(err)
	}

	if *interval > 0 {
		for {
			waitTime := *interval

			// Respect the min poll duration returned by the API
			if *interval < minPollDuration {
				log.Printf("Increasing poll duration based on rate-limit headers")
				waitTime = minPollDuration
			}

			log.Printf("Waiting for %v (minimum of %v)", waitTime, minPollDuration)
			time.Sleep(waitTime)

			minPollDuration, err = collectFunc()
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

type stringSliceFlag []string

func (i *stringSliceFlag) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *stringSliceFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}
