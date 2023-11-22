package backend

import (
	"log"
	"time"

	"github.com/buildkite/buildkite-agent-metrics/collector"
	newrelic "github.com/newrelic/go-agent"
)

const newRelicConnectionTimeout = time.Second * 30

// NewRelicBackend sends metrics to New Relic Insights
type NewRelicBackend struct {
	client newrelic.Application
}

// NewNewRelicBackend returns a backend for New Relic
// Where appName is your desired application name in New Relic
//
//	and licenseKey is your New Relic license key
func NewNewRelicBackend(appName string, licenseKey string) (*NewRelicBackend, error) {
	config := newrelic.NewConfig(appName, licenseKey)
	app, err := newrelic.NewApplication(config)
	if err != nil {
		return nil, err
	}

	// Waiting for connection is essential or no data will make it during short-lived execution (e.g. Lambda)
	err = app.WaitForConnection(newRelicConnectionTimeout)
	if err != nil {
		return nil, err
	}

	return &NewRelicBackend{
		client: app,
	}, nil
}

// Collect metrics
func (nr *NewRelicBackend) Collect(r *collector.Result) error {
	// Publish event for each queue
	for queue, metrics := range r.Queues {
		data := toCustomEvent(r.Cluster, queue, metrics)
		err := nr.client.RecordCustomEvent("BuildkiteQueueMetrics", data)
		if err != nil {
			return err
		}

		nr.client.RecordCustomEvent("queue_agent_metrics", data)
	}

	return nil
}

// toCustomEvent converts a map of metrics to a valid New Relic event body
func toCustomEvent(clusterName, queueName string, queueMetrics map[string]int) map[string]any {
	eventData := map[string]any{
		"Cluster": clusterName,
		"Queue":   queueName,
	}

	for k, v := range queueMetrics {
		eventData[k] = v
	}

	return eventData
}

// Close by shutting down NR client
func (nr *NewRelicBackend) Close() error {
	nr.client.Shutdown(newRelicConnectionTimeout)
	log.Printf("Disposed New Relic client")

	return nil
}
