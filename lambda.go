package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/buildkite/buildkite-metrics/collector"
	"github.com/eawsy/aws-lambda-go/service/lambda/runtime"
	"gopkg.in/buildkite/go-buildkite.v2/buildkite"
)

func handle(evt json.RawMessage, ctx *runtime.Context) (interface{}, error) {
	output := &bytes.Buffer{}
	log.SetOutput(output)

	org := os.Getenv("BUILDKITE_ORG")
	token := os.Getenv("BUILDKITE_TOKEN")
	backendOpt := os.Getenv("BUILDKITE_BACKEND")

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

	var backend Backend
	if backendOpt == "statsd" {
		backend, err = NewStatsDClient(os.Getenv("STATSD_HOST"))
		if err != nil {
			return nil, err
		}
	} else {
		backend = &CloudWatchBackend{}
	}

	res, err := col.Collect()
	if err != nil {
		return nil, err
	}

	res.Dump()

	err = backend.Collect(res)
	if err != nil {
		return nil, err
	}

	log.Printf("Finished in %s", time.Now().Sub(t))
	return output.String(), nil
}

func init() {
	runtime.HandleFunc(handle)
}
