package main

import (
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

func handle() (error) {
	token := os.Getenv("BUILDKITE_AGENT_TOKEN")
	backendOpt := os.Getenv("BUILDKITE_BACKEND")
	queue := os.Getenv("BUILDKITE_QUEUE")
	clwDimensions := os.Getenv("BUILDKITE_CLOUDWATCH_DIMENSIONS")
	quietString := os.Getenv("BUILDKITE_QUIET")
	quiet := quietString == "1" || quietString == "true"

	if quiet {
		log.SetOutput(ioutil.Discard)
	}

	t := time.Now()

	userAgent := fmt.Sprintf("buildkite-agent-metrics/%s buildkite-metrics-lambda", version.Version)

	c := collector.Collector{
		UserAgent: userAgent,
		Endpoint:  "https://agent.buildkite.com/v3",
		Token:     token,
		Queue:     queue,
		Quiet:     quiet,
		Debug:     false,
		DebugHttp: false,
	}

	var b backend.Backend
	var err error
	if backendOpt == "statsd" {
		statsdHost := os.Getenv("STATSD_HOST")
		statsdTags := strings.ToLower(os.Getenv("STATSD_TAGS")) == "true"
		b, err = backend.NewStatsDBackend(statsdHost, statsdTags)
		if err != nil {
			return err
		}
	} else {
		dimensions, err := backend.ParseCloudWatchDimensions(clwDimensions)
		if err != nil {
			return err
		}
		b = backend.NewCloudWatchBackend(dimensions)
	}

	res, err := c.Collect()
	if err != nil {
		return err
	}

	res.Dump()

	err = b.Collect(res)
	if err != nil {
		return err
	}

	log.Printf("Finished in %s", time.Now().Sub(t))
	return nil
}

