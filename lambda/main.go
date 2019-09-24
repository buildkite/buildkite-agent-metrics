package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/buildkite/buildkite-agent-metrics/backend"
	"github.com/buildkite/buildkite-agent-metrics/collector"
	"github.com/buildkite/buildkite-agent-metrics/version"
)

var (
	nextPollTime time.Time
)

func main() {
	if os.Getenv(`DEBUG`) != "" {
		_, err := Handler(context.Background(), json.RawMessage([]byte{}))
		if err != nil {
			log.Fatal(err)
		}
	} else {
		lambda.Start(Handler)
	}
}

func Handler(ctx context.Context, evt json.RawMessage) (string, error) {
	token := os.Getenv("BUILDKITE_AGENT_TOKEN")
	awsRegion := os.Getenv("AWS_REGION")
	ssmTokenKey := os.Getenv("BUILDKITE_AGENT_TOKEN_SSM_KEY")
	backendOpt := os.Getenv("BUILDKITE_BACKEND")
	queue := os.Getenv("BUILDKITE_QUEUE")
	clwDimensions := os.Getenv("BUILDKITE_CLOUDWATCH_DIMENSIONS")
	quietString := os.Getenv("BUILDKITE_QUIET")
	quiet := quietString == "1" || quietString == "true"

	if quiet {
		log.SetOutput(ioutil.Discard)
	}

	t := time.Now()

	if !nextPollTime.IsZero() && nextPollTime.After(t) {
		log.Printf("Skipping polling, next poll time is in %v",
			nextPollTime.Sub(t))
		return "", nil
	}

	if ssmTokenKey != "" {
		token = backend.RetrieveFromParameterStore(ssmTokenKey)
	}

	var queues = []string{}
	if queue != "" {
		queues = append(queues, queue)
	}

	userAgent := fmt.Sprintf("buildkite-agent-metrics/%s buildkite-agent-metrics-lambda", version.Version)

	c := collector.Collector{
		UserAgent: userAgent,
		Endpoint:  "https://agent.buildkite.com/v3",
		Token:     token,
		Queues:    queues,
		Quiet:     quiet,
		Debug:     false,
		DebugHttp: false,
	}

	var b backend.Backend
	var err error
	switch strings.ToLower(backendOpt) {
	case "statsd":
		statsdHost := os.Getenv("STATSD_HOST")
		statsdTags := strings.EqualFold(os.Getenv("STATSD_TAGS"), "true")
		b, err = backend.NewStatsDBackend(statsdHost, statsdTags)
		if err != nil {
			return "", err
		}
	case "newrelic":
		nrAppName := os.Getenv("NEWRELIC_APP_NAME")
		nrLicenseKey := os.Getenv("NEWRELIC_LICENSE_KEY")
		b, err = backend.NewNewRelicBackend(nrAppName, nrLicenseKey)
		if err != nil {
			fmt.Printf("Error starting New Relic client: %v\n", err)
			os.Exit(1)
		}
	default:
		dimensions, err := backend.ParseCloudWatchDimensions(clwDimensions)
		if err != nil {
			return "", err
		}
		b = backend.NewCloudWatchBackend(awsRegion, dimensions)
	}

	res, err := c.Collect()
	if err != nil {
		return "", err
	}

	res.Dump()

	err = b.Collect(res)
	if err != nil {
		return "", err
	}

	original, ok := b.(backend.Closer)
	if ok {
		err := original.Close()
		if err != nil {
			return "", err
		}
	}

	log.Printf("Finished in %s", time.Now().Sub(t))

	// Store the next acceptable poll time in global state
	nextPollTime = time.Now().Add(res.PollDuration)

	return "", nil
}
