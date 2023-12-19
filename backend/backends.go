package backend

import "github.com/buildkite/buildkite-agent-metrics/v5/collector"

// Backend is a receiver of metrics
type Backend interface {
	Collect(r *collector.Result) error
}

// Closer is an interface for backends that need to dispose of resources
type Closer interface {
	Close() error
}
