package backend

import (
	"github.com/DataDog/datadog-go/statsd"
	"github.com/buildkite/buildkite-metrics/collector"
)

// StatsD sends metrics to StatsD (Datadog spec)
type StatsD struct {
	client        *statsd.Client
	tagsSupported bool
}

func NewStatsDBackend(host string, tagsSupported bool) (*StatsD, error) {
	c, err := statsd.NewBuffered(host, 100)
	if err != nil {
		return nil, err
	}
	// prefix every metric with the app name
	c.Namespace = "buildkite."
	return &StatsD{
		client:        c,
		tagsSupported: tagsSupported,
	}, nil
}

func (cb *StatsD) Collect(r *collector.Result) error {
	for name, value := range r.Totals {
		if err := cb.client.Gauge(name, float64(value), []string{}, 1.0); err != nil {
			return err
		}
	}

	for queue, counts := range r.Queues {
		for name, value := range counts {
			var finalName string
			tags := []string{}
			if cb.tagsSupported {
				finalName = "queues." + name
				tags = []string{"queue:" + queue}
			} else {
				finalName = "queues." + queue + "." + name
			}
			if err := cb.client.Gauge(finalName, float64(value), tags, 1.0); err != nil {
				return err
			}
		}
	}

	if err := cb.client.Flush(); err != nil {
		return err
	}

	return nil
}
