package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/buildkite/buildkite-metrics/backend"
	"github.com/buildkite/buildkite-metrics/collector"
	"github.com/buildkite/buildkite-metrics/version"
	"github.com/buildkite/go-buildkite/buildkite"
)

var bk backend.Backend

func main() {
	var (
		accessToken = flag.String("token", "", "A Buildkite API Access Token")
		orgSlug     = flag.String("org", "", "A Buildkite Organization Slug")
		interval    = flag.Duration("interval", 0, "Update metrics every interval, rather than once")
		history     = flag.Duration("history", time.Hour*8, "Fetch historical data to use for finished builds")
		debug       = flag.Bool("debug", false, "Show debug output")
		debugHTTP   = flag.Bool("debug-http", false, "Show full http traces")
		showVersion = flag.Bool("version", false, "Show the version")
		quiet       = flag.Bool("quiet", false, "Only print errors")
		dryRun      = flag.Bool("dry-run", false, "Whether to only print metrics")
		apiEndpoint = flag.String("api-endpoint", "", "A custom buildkite api endpoint")

		// backend config
		backendOpt = flag.String("backend", "cloudwatch", "Specify the backend to send metrics to: cloudwatch, statsd")
		statsdHost = flag.String("statsd-host", "127.0.0.1:8125", "Specify the StatsD server")
		statsdTags = flag.Bool("statsd-tags", false, "Whether your StatsD server supports tagging like Datadog")

		// filters
		queue = flag.String("queue", "", "Only include a specific queue")
	)

	flag.Parse()

	if *showVersion {
		fmt.Printf("buildkite-metrics %s\n", version.VersionString())
		os.Exit(0)
	}

	if *accessToken == "" {
		fmt.Println("Must provide a value for -token")
		os.Exit(1)
	}

	if *orgSlug == "" {
		fmt.Println("Must provide a value for -org")
		os.Exit(1)
	}

	lowerBackendOpt := strings.ToLower(*backendOpt)
	if lowerBackendOpt == "cloudwatch" {
		bk = backend.NewCloudWatchBackend()
	} else if lowerBackendOpt == "statsd" {
		var err error
		bk, err = backend.NewStatsDBackend(*statsdHost, *statsdTags)
		if err != nil {
			fmt.Printf("Error starting StatsD, err: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("Must provide a supported backend: cloudwatch, statsd")
		os.Exit(1)
	}

	if *quiet {
		log.SetOutput(ioutil.Discard)
	}

	config, err := buildkite.NewTokenConfig(*accessToken, false)
	if err != nil {
		fmt.Printf("client config failed: %s\n", err)
		os.Exit(1)
	}

	client := buildkite.NewClient(config.Client())
	if *debugHTTP {
		buildkite.SetHttpDebug(*debug)
	}

	client.UserAgent = fmt.Sprintf(
		"%s buildkite-metrics/%s buildkite-metrics-cli queue=%q,interval=%v",
		client.UserAgent, version.Version, *queue, *interval,
	)

	if *apiEndpoint != "" {
		apiEndpointURL, err := url.Parse(*apiEndpoint)
		if err != nil {
			fmt.Printf("%v", err)
			os.Exit(1)
		}
		client.BaseURL = apiEndpointURL
		config.APIHost = apiEndpointURL.Host
	}

	col := collector.New(client, collector.Opts{
		OrgSlug: *orgSlug,
		History: *history,
		Queue:   *queue,
		Debug:   *debug,
	})

	f := func() error {
		t := time.Now()

		res, err := col.Collect()
		if err != nil {
			return err
		}

		if !*quiet {
			res.Dump()
		}

		if !*dryRun {
			err = bk.Collect(res)
			if err != nil {
				return err
			}
		}

		log.Printf("Finished in %s", time.Now().Sub(t))
		return nil
	}

	if err := f(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if *interval > 0 {
		for _ = range time.NewTicker(*interval).C {
			if err := f(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}
}
