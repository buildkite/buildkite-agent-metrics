package backend

import (
	"context"
	"testing"

	"github.com/buildkite/buildkite-agent-metrics/v5/collector"
	"github.com/google/go-cmp/cmp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/trace/noop"
)

// newTestOTelBackend builds an OpenTelemetry backend wired to an in-memory
// manual reader so tests can inspect the exported metric data.
func newTestOTelBackend(t *testing.T) (*OpenTelemetryBackend, *sdkmetric.ManualReader) {
	t.Helper()

	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))

	b, err := newOpenTelemetryBackend(noop.NewTracerProvider().Tracer("test"), provider.Meter("test"))
	if err != nil {
		t.Fatalf("newOpenTelemetryBackend() = %v", err)
	}
	return b, reader
}

// collectMetric gathers all exported metrics and returns the one named name.
func collectMetric(t *testing.T, reader *sdkmetric.ManualReader, name string) metricdata.Metrics {
	t.Helper()

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatalf("reader.Collect() = %v", err)
	}
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == name {
				return m
			}
		}
	}
	t.Fatalf("metric %q not found in exported metrics", name)
	return metricdata.Metrics{}
}

type otelPoint struct {
	Attributes map[string]string
	Value      int64
}

// gaugePoints returns the int64 gauge data points of the named metric, keyed by
// the value of their "queue" attribute ("" for the cluster-total series). It
// fails the test if the metric is not an Int64 gauge, which is the regression
// guard for issue #541: counter instruments export as metricdata.Sum instead.
func gaugePoints(t *testing.T, reader *sdkmetric.ManualReader, name string) map[string]otelPoint {
	t.Helper()

	m := collectMetric(t, reader, name)
	gauge, ok := m.Data.(metricdata.Gauge[int64])
	if !ok {
		t.Fatalf("metric %q has data type %T, want metricdata.Gauge[int64]", name, m.Data)
	}

	points := make(map[string]otelPoint, len(gauge.DataPoints))
	for _, dp := range gauge.DataPoints {
		attrs := make(map[string]string)
		for _, kv := range dp.Attributes.ToSlice() {
			attrs[string(kv.Key)] = kv.Value.AsString()
		}
		points[attrs["queue"]] = otelPoint{Attributes: attrs, Value: dp.Value}
	}
	return points
}

// gauges maps each exported metric name to the collector key it is recorded
// from. All eight are point-in-time gauges.
var gauges = map[string]string{
	"buildkite.jobs.scheduled":         collector.ScheduledJobsCount,
	"buildkite.jobs.running":           collector.RunningJobsCount,
	"buildkite.jobs.unfinished":        collector.UnfinishedJobsCount,
	"buildkite.jobs.waiting":           collector.WaitingJobsCount,
	"buildkite.agents.idle":            collector.IdleAgentCount,
	"buildkite.agents.busy":            collector.BusyAgentCount,
	"buildkite.agents.total":           collector.TotalAgentCount,
	"buildkite.agents.busy_percentage": collector.BusyAgentPercentage,
}

// TestOpenTelemetryCollect drives Collect end to end against an in-memory reader
// and checks that every metric is exported as an Int64 gauge with the expected
// value and attributes for the cluster total and each queue.
func TestOpenTelemetryCollect(t *testing.T) {
	b, reader := newTestOTelBackend(t)
	if err := b.Collect(newTestResult(t)); err != nil {
		t.Fatalf("Collect() = %v", err)
	}

	for name, key := range gauges {
		t.Run(name, func(t *testing.T) {
			want := map[string]otelPoint{
				"": {
					Attributes: map[string]string{"org": "", "cluster": "test_cluster"},
					Value:      int64(fakeTotals[key]),
				},
				"default": {
					Attributes: map[string]string{"org": "", "cluster": "test_cluster", "queue": "default"},
					Value:      int64(fakeDefaultQueue[key]),
				},
				"deploy": {
					Attributes: map[string]string{"org": "", "cluster": "test_cluster", "queue": "deploy"},
					Value:      int64(fakeDeployQueue[key]),
				},
			}

			got := gaugePoints(t, reader, name)
			if diff := cmp.Diff(got, want); diff != "" {
				t.Errorf("%s data points diff (-got +want):\n%s", name, diff)
			}
		})
	}
}

// TestOpenTelemetryJobMetricsUseLastValue records each job metric across two
// collection intervals and asserts the exported value is the most recent one,
// not the sum of both. A counter instrument would accumulate; a gauge does not.
// Regression test for issue #541.
func TestOpenTelemetryJobMetricsUseLastValue(t *testing.T) {
	jobMetrics := []string{
		"buildkite.jobs.scheduled",
		"buildkite.jobs.running",
		"buildkite.jobs.unfinished",
		"buildkite.jobs.waiting",
	}

	for _, name := range jobMetrics {
		t.Run(name, func(t *testing.T) {
			b, reader := newTestOTelBackend(t)
			key := gauges[name]

			for _, v := range []int{5, 3} {
				r := &collector.Result{
					Cluster: "test_cluster",
					Totals:  map[string]int{key: v},
				}
				if err := b.Collect(r); err != nil {
					t.Fatalf("Collect() = %v", err)
				}
			}

			points := gaugePoints(t, reader, name)
			if len(points) != 1 {
				t.Fatalf("len(points) = %d, want 1", len(points))
			}
			if got, want := points[""].Value, int64(3); got != want {
				t.Errorf("%s = %d, want %d (last value, not accumulated)", name, got, want)
			}
		})
	}
}
