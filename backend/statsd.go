package backend

import (
	"fmt"
	"log"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/buildkite/buildkite-agent-metrics/v5/collector"
)

// StatsD sends metrics to StatsD (Datadog spec)
type StatsD struct {
	client        *statsd.Client
	tagsSupported bool
}

func NewStatsDBackend(host string, tagsSupported bool, metricsPrefix string) (*StatsD, error) {
	client, err := statsd.NewBuffered(host, 100)
	if err != nil {
		return nil, err
	}
	// prefix every metric with the app name
	client.Namespace = metricsPrefix
	log.Printf("using metrics prefix: %s", metricsPrefix)
	return &StatsD{
		client:        client,
		tagsSupported: tagsSupported,
	}, nil
}

func (cb *StatsD) Collect(r *collector.Result) error {
	collectFunc := cb.collectWithoutTags
	if cb.tagsSupported {
		collectFunc = cb.collectWithTags
	}

	if err := collectFunc(r); err != nil {
		return err
	}

	return cb.client.Flush()
}

// collectWithTags tags clusters and queues.
func (cb *StatsD) collectWithTags(r *collector.Result) error {
	commonTags := make([]string, 0, 2)
	prefix := ""
	if r.Cluster != "" {
		commonTags = append(commonTags, "cluster:"+r.Cluster)
		prefix = "clusters."
	}

	for name, value := range r.Totals {
		if err := cb.client.Gauge(prefix+name, float64(value), commonTags, 1.0); err != nil {
			return err
		}
	}

	for queue, counts := range r.Queues {
		tags := append(commonTags, "queue:"+queue)

		for name, value := range counts {
			if err := cb.client.Gauge(prefix+"queues."+name, float64(value), tags, 1.0); err != nil {
				return err
			}
		}
	}

	return cb.client.Flush()
}

// collectWithoutTags embeds clusters and queues into metric names.
func (cb *StatsD) collectWithoutTags(r *collector.Result) error {
	prefix := ""
	if r.Cluster != "" {
		prefix = fmt.Sprintf("clusters.%s.", r.Cluster)
	}

	for name, value := range r.Totals {
		if err := cb.client.Gauge(prefix+name, float64(value), nil, 1.0); err != nil {
			return err
		}
	}

	for queue, counts := range r.Queues {
		prefix := fmt.Sprintf("queues.%s.", queue)
		if r.Cluster != "" {
			prefix = fmt.Sprintf("clusters.%s.queues.%s.", r.Cluster, queue)
		}

		for name, value := range counts {
			if err := cb.client.Gauge(prefix+name, float64(value), nil, 1.0); err != nil {
				return err
			}
		}
	}

	return cb.client.Flush()
}
