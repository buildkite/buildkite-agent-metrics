package backend

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/buildkite/buildkite-agent-metrics/v5/collector"
	"github.com/google/go-cmp/cmp"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

var (
	fakeTotals       = make(map[string]int)
	fakeDefaultQueue = make(map[string]int)
	fakeDeployQueue  = make(map[string]int)
)

var _ func() *Prometheus = NewPrometheusBackend

func init() {
	for i, metric := range collector.AllMetrics {
		fakeTotals[metric] = i
		fakeDefaultQueue[metric] = i + 100
		fakeDeployQueue[metric] = i + 200
	}
}

func newTestResult(t *testing.T) *collector.Result {
	t.Helper()

	res := &collector.Result{
		Totals:  fakeTotals,
		Cluster: "test_cluster",
		Queues: map[string]map[string]int{
			"default": fakeDefaultQueue,
			"deploy":  fakeDeployQueue,
		},
	}
	return res
}

// gatherMetrics runs the backend's Prometheus gatherer, and returns the metric
// families grouped by name.
func gatherMetrics(t *testing.T, p *Prometheus) map[string]*dto.MetricFamily {
	t.Helper()

	mfs, err := p.gatherer.Gather()
	if err != nil {
		t.Fatalf("prometheus.Registry.Gather() = %v", err)
		return nil
	}

	mfsm := make(map[string]*dto.MetricFamily)
	for _, mf := range mfs {
		mfsm[*mf.Name] = mf
	}
	return mfsm
}

func TestCollect(t *testing.T) {
	p := NewPrometheusBackendWithoutDefaultMetrics()
	if err := p.Collect(newTestResult(t)); err != nil {
		t.Fatalf("p.Collect() = %v", err)
	}
	metricFamilies := gatherMetrics(t, p)

	if got, want := len(metricFamilies), 16; got != want {
		t.Errorf("len(metricFamilies) = %d, want %d", got, want)
	}

	type promMetric struct {
		Labels map[string]string
		Value  float64
	}

	tcs := []struct {
		group       string
		metricName  string
		wantHelp    string
		wantType    dto.MetricType
		wantMetrics []promMetric
	}{
		{
			group:      "Total",
			metricName: "buildkite_total_running_jobs_count",
			wantHelp:   "Buildkite Total: RunningJobsCount",
			wantType:   dto.MetricType_GAUGE,
			wantMetrics: []promMetric{
				{
					Labels: map[string]string{"cluster": "test_cluster"},
					Value:  float64(fakeTotals[collector.RunningJobsCount]),
				},
			},
		},
		{
			group:      "Total",
			metricName: "buildkite_total_scheduled_jobs_count",
			wantHelp:   "Buildkite Total: ScheduledJobsCount",
			wantType:   dto.MetricType_GAUGE,
			wantMetrics: []promMetric{
				{
					Labels: map[string]string{"cluster": "test_cluster"},
					Value:  float64(fakeTotals[collector.ScheduledJobsCount]),
				},
			},
		},
		{
			group:      "Queues",
			metricName: "buildkite_queues_unfinished_jobs_count",
			wantHelp:   "Buildkite Queues: UnfinishedJobsCount",
			wantType:   dto.MetricType_GAUGE,
			wantMetrics: []promMetric{
				{
					Labels: map[string]string{
						"cluster": "test_cluster",
						"queue":   "default",
					},
					Value: float64(fakeDefaultQueue[collector.UnfinishedJobsCount]),
				},
				{
					Labels: map[string]string{
						"cluster": "test_cluster",
						"queue":   "deploy",
					},
					Value: float64(fakeDeployQueue[collector.UnfinishedJobsCount]),
				},
			},
		},
		{
			group:      "Queues",
			metricName: "buildkite_queues_idle_agent_count",
			wantHelp:   "Buildkite Queues: IdleAgentCount",
			wantType:   dto.MetricType_GAUGE,
			wantMetrics: []promMetric{
				{
					Labels: map[string]string{
						"cluster": "test_cluster",
						"queue":   "default",
					},
					Value: float64(fakeDefaultQueue[collector.IdleAgentCount]),
				},
				{
					Labels: map[string]string{
						"cluster": "test_cluster",
						"queue":   "deploy",
					},
					Value: float64(fakeDeployQueue[collector.IdleAgentCount]),
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(fmt.Sprintf("%s/%s", tc.group, tc.metricName), func(t *testing.T) {
			mf, ok := metricFamilies[tc.metricName]
			if !ok {
				t.Errorf("no metric found for name %s", tc.metricName)
			}

			if got, want := mf.GetHelp(), tc.wantHelp; got != want {
				t.Errorf("mf.GetHelp() = %q, want %q", got, want)
			}

			if got, want := mf.GetType(), tc.wantType; got != want {
				t.Errorf("mf.GetType() = %q, want %q", got, want)
			}

			// Convert the metric family into an easier-to-diff form.
			ms := mf.GetMetric()
			var gotMetrics []promMetric
			for _, m := range ms {
				gotMetric := promMetric{
					Value:  m.Gauge.GetValue(),
					Labels: make(map[string]string),
				}
				for _, label := range m.Label {
					gotMetric.Labels[label.GetName()] = label.GetValue()
				}
				gotMetrics = append(gotMetrics, gotMetric)
			}

			if diff := cmp.Diff(gotMetrics, tc.wantMetrics); diff != "" {
				t.Errorf("metrics diff (-got +want):\n%s", diff)
			}
		})
	}
}

func TestPrometheusDefaultMetrics(t *testing.T) {
	customMetric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "test_custom_metric",
		Help: "A custom metric registered outside the Prometheus backend.",
	})
	prometheus.MustRegister(customMetric)
	t.Cleanup(func() { prometheus.Unregister(customMetric) })

	p := NewPrometheusBackend()
	if err := p.Collect(newTestResult(t)); err != nil {
		t.Fatalf("p.Collect() = %v", err)
	}

	recorder := httptest.NewRecorder()
	p.handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET /metrics status = %d, want %d", recorder.Code, http.StatusOK)
	}

	metricFamilies := gatherMetrics(t, p)
	for _, prefix := range []string{"buildkite_", "go_", "process_", "promhttp_", "test_custom_metric"} {
		if !containsMetricWithPrefix(metricFamilies, prefix) {
			t.Errorf("no metric found with prefix %q", prefix)
		}
	}
}

func TestPrometheusWithoutDefaultMetrics(t *testing.T) {
	p := NewPrometheusBackendWithoutDefaultMetrics()
	if err := p.Collect(newTestResult(t)); err != nil {
		t.Fatalf("p.Collect() = %v", err)
	}

	recorder := httptest.NewRecorder()
	p.handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET /metrics status = %d, want %d", recorder.Code, http.StatusOK)
	}

	metricFamilies := gatherMetrics(t, p)
	if got, want := len(metricFamilies), 16; got != want {
		t.Errorf("len(metricFamilies) = %d, want %d", got, want)
	}
	for name := range metricFamilies {
		if !strings.HasPrefix(name, "buildkite_") {
			t.Errorf("unexpected non-Buildkite metric %q", name)
		}
	}
}

func containsMetricWithPrefix(metricFamilies map[string]*dto.MetricFamily, prefix string) bool {
	for name := range metricFamilies {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func TestCamelToUnderscore(t *testing.T) {
	tcs := []struct {
		input string
		want  string
	}{
		{
			input: "TotalAgentCount",
			want:  "total_agent_count",
		},
		{
			input: "Total@#4JobsCount",
			want:  "total@#4_jobs_count",
		},
		{
			input: "BuildkiteQueuesIdleAgentCount1_11",
			want:  "buildkite_queues_idle_agent_count1_11",
		},
	}

	for _, tc := range tcs {
		if got := camelToUnderscore(tc.input); got != tc.want {
			t.Errorf("camelToUnderscore(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
