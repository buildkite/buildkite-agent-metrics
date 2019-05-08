package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/buildkite/buildkite-agent-metrics/backend"
	"github.com/buildkite/buildkite-agent-metrics/collector"
	"github.com/buildkite/buildkite-agent-metrics/version"
)

var bk backend.Backend

func main() {
	var (
		token       = flag.String("token", "", "A Buildkite Agent Registration Token")
		interval    = flag.Duration("interval", 0, "Update metrics every interval, rather than once")
		showVersion = flag.Bool("version", false, "Show the version")
		quiet       = flag.Bool("quiet", false, "Only print errors")
		debug       = flag.Bool("debug", false, "Show debug output")
		debugHttp   = flag.Bool("debug-http", false, "Show full http traces")
		dryRun      = flag.Bool("dry-run", false, "Whether to only print metrics")
		endpoint    = flag.String("endpoint", "https://agent.buildkite.com/v3", "A custom Buildkite Agent API endpoint")

		// backend config
		backendOpt     = flag.String("backend", "cloudwatch", "Specify the backend to use: cloudwatch, statsd, prometheus, stackdriver")
		statsdHost     = flag.String("statsd-host", "127.0.0.1:8125", "Specify the StatsD server")
		statsdTags     = flag.Bool("statsd-tags", false, "Whether your StatsD server supports tagging like Datadog")
		prometheusAddr = flag.String("prometheus-addr", ":8080", "Prometheus metrics transport bind address")
		prometheusPath = flag.String("prometheus-path", "/metrics", "Prometheus metrics transport path")
		clwRegion      = flag.String("cloudwatch-region", "", "AWS Region to connect to, defaults to $AWS_REGION or us-east-1")
		clwDimensions  = flag.String("cloudwatch-dimensions", "", "Cloudwatch dimensions to index metrics under, in the form of Key=Value, Other=Value")
		gcpProjectID   = flag.String("stackdriver-projectid", "", "Specify Stackdriver Project ID")

		// filters
		queue = flag.String("queue", "", "Only include a specific queue")
	)

	flag.Parse()

	if *showVersion {
		fmt.Printf("buildkite-agent-metrics %s\n", version.Version)
		os.Exit(0)
	}

	if *token == "" {
		fmt.Println("Must provide a value for -token")
		os.Exit(1)
	}

	switch strings.ToLower(*backendOpt) {
	case "cloudwatch":
		region := *clwRegion
		if envRegion := os.Getenv(`AWS_REGION`); region == "" && envRegion != "" {
			region = envRegion
		} else {
			region = `us-east-1`
		}
		dimensions, err := backend.ParseCloudWatchDimensions(*clwDimensions)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		bk = backend.NewCloudWatchBackend(region, dimensions)
	case "statsd":
		var err error
		bk, err = backend.NewStatsDBackend(*statsdHost, *statsdTags)
		if err != nil {
			fmt.Printf("Error starting StatsD, err: %v\n", err)
			os.Exit(1)
		}
	case "prometheus":
		bk = backend.NewPrometheusBackend(*prometheusPath, *prometheusAddr)
	case "stackdriver":
		var err error
		bk, err = backend.NewStackDriverBackend(*gcpProjectID)
		if err != nil {
			fmt.Printf("Error starting Stackdriver backend, err: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Println("Must provide a supported backend: cloudwatch, statsd, prometheus")
		os.Exit(1)
	}

	if *quiet {
		log.SetOutput(ioutil.Discard)
	}

	userAgent := fmt.Sprintf("buildkite-agent-metrics/%s buildkite-agent-metrics-cli", version.Version)
	if *interval > 0 {
		userAgent += fmt.Sprintf(" interval=%s", *interval)
	}

	c := collector.Collector{
		UserAgent: userAgent,
		Endpoint:  *endpoint,
		Token:     *token,
		Queue:     *queue,
		Quiet:     *quiet,
		Debug:     *debug,
		DebugHttp: *debugHttp,
	}

	f := func() (time.Duration, error) {
		t := time.Now()

		result, err := c.Collect()
		if err != nil {
			return time.Duration(0), err
		}

		if !*dryRun {
			err = bk.Collect(result)
			if err != nil {
				return time.Duration(0), err
			}
		}

		log.Printf("Finished in %s", time.Now().Sub(t))
		return result.PollDuration, nil
	}

	minPollDuration, err := f()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
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

			minPollDuration, err = f()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}
}
