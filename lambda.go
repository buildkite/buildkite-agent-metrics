package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/buildkite/buildkite-metrics/backend"
	"github.com/buildkite/buildkite-metrics/collector"
	"github.com/buildkite/go-buildkite/buildkite"
	"github.com/eawsy/aws-lambda-go/service/lambda/runtime"
)

func handle(evt json.RawMessage, ctx *runtime.Context) (interface{}, error) {
	org := os.Getenv("BUILDKITE_ORG")
	token := os.Getenv("BUILDKITE_TOKEN")
	backendOpt := os.Getenv("BUILDKITE_BACKEND")
	queue := os.Getenv("BUILDKITE_QUEUE")
	quiet := os.Getenv("BUILDKITE_QUIET")

	if quiet == "1" || quiet == "false" {
		log.SetOutput(ioutil.Discard)
	}

	config, err := buildkite.NewTokenConfig(token, false)
	if err != nil {
		return nil, err
	}

	client := buildkite.NewClient(config.Client())
	t := time.Now()

	client.UserAgent = client.UserAgent + " buildkite-metrics/" + Version + " buildkite-metrics-lambda"

	col := collector.New(client, collector.Opts{
		OrgSlug:    org,
		Historical: time.Hour * 24,
	})

	if queue != "" {
		col.Queue = queue
	}

	var bk backend.Backend
	if backendOpt == "statsd" {
		bk, err = backend.NewStatsDBackend(os.Getenv("STATSD_HOST"), strings.ToLower(os.Getenv("STATSD_TAGS")) == "true")
		if err != nil {
			return nil, err
		}
	} else {
		bk = &backend.CloudWatchBackend{}
	}

	return "", retry(time.Minute, func() error {
		res, err := col.Collect()
		if err != nil {
			return err
		}

		res.Dump()

		err = bk.Collect(res)
		if err != nil {
			return err
		}

		log.Printf("Finished in %s", time.Now().Sub(t))
		return nil
	})
}

func retry(timeout time.Duration, callback func() error) (err error) {
	t0 := time.Now()
	i := 0
	for {
		i++

		err = callback()
		if err == nil {
			return
		}

		delta := time.Now().Sub(t0)
		if delta > timeout {
			return fmt.Errorf("after %d attempts (during %s), last error: %s", i, delta, err)
		}

		time.Sleep(time.Second * 2)
		log.Println("retrying after error:", err)
	}
}

func init() {
	runtime.HandleFunc(handle)
}
