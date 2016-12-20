package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/buildkite/buildkite-metrics/backend"
	"github.com/buildkite/buildkite-metrics/collector"
	"github.com/eawsy/aws-lambda-go/service/lambda/runtime"
	"gopkg.in/buildkite/go-buildkite.v2/buildkite"
)

func handle(evt json.RawMessage, ctx *runtime.Context) (interface{}, error) {
	org := os.Getenv("BUILDKITE_ORG")
	token := os.Getenv("BUILDKITE_TOKEN")
	backendOpt := os.Getenv("BUILDKITE_BACKEND")
	queue := os.Getenv("BUILDKITE_QUEUE")

	config, err := buildkite.NewTokenConfig(token, false)
	if err != nil {
		return nil, err
	}

	client := buildkite.NewClient(config.Client())
	t := time.Now()

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

	res, err := col.Collect()
	if err != nil {
		return nil, err
	}

	res.Dump()

	err = bk.Collect(res)
	if err != nil {
		return nil, err
	}

	log.Printf("Finished in %s", time.Now().Sub(t))
	return "", nil
}

func init() {
	runtime.HandleFunc(handle)
}
