package backend

import "github.com/buildkite/buildkite-metrics/collector"

// Backend is a receiver of metrics
type Backend interface {
	Collect(r *collector.Result) error
}
