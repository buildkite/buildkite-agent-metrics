package backend

import "github.com/buildkite/buildkite-agent-metrics/collector"

// Backend is a receiver of metrics
type Backend interface {
	Collect(r *collector.Result) error
}
