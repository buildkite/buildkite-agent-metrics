package backend

import "github.com/buildkite/buildkite-agent-metrics/collector"

// Stdout is a "metrics collection backend" that simply prints metrics to standard output.
// Useful for testing and local debugging
type Stdout struct{}

func NewStdoutBackend() *Stdout {
	return &Stdout{}
}

func (s *Stdout) Collect(r *collector.Result) error {
	r.Dump()
	return nil
}
