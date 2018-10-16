package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/buildkite/buildkite-agent-metrics/backend"
	"github.com/buildkite/buildkite-agent-metrics/collector"
	"github.com/buildkite/buildkite-agent-metrics/version"
	"github.com/eawsy/aws-lambda-go/service/lambda/runtime"
)

func handle(evt json.RawMessage, ctx *runtime.Context) (interface{}, error) {
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

	userAgent := fmt.Sprintf("buildkite-metrics/%s buildkite-metrics-lambda", version.Version)

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
			return nil, err
		}
	} else {
		dimensions, err := backend.ParseCloudWatchDimensions(clwDimensions)
		if err != nil {
			return nil, err
		}
		b = backend.NewCloudWatchBackend(dimensions)
	}

	res, err := c.Collect()
	if err != nil {
		return nil, err
	}

	res.Dump()

	err = b.Collect(res)
	if err != nil {
		return nil, err
	}

	log.Printf("Finished in %s", time.Now().Sub(t))
	return "", nil
}

func init() {
	runtime.HandleFunc(handle)
}
