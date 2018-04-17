package backend

import (
	"fmt"
	"testing"

	"github.com/buildkite/buildkite-metrics/collector"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

const (
	wantHaveFmt        = "want %v, have %v"
	runningBuildsCount = iota
	scheduledBuildsCount
	runningJobsCount
	scheduledJobsCount
	unfinishedJobsCount
	idleAgentCount
	busyAgentCount
	totalAgentCount
)

func newTestResult(t *testing.T) *collector.Result {
	t.Helper()
	totals := map[string]int{
		"RunningBuildsCount":   runningBuildsCount,
		"ScheduledBuildsCount": scheduledBuildsCount,
		"RunningJobsCount":     runningJobsCount,
		"ScheduledJobsCount":   scheduledJobsCount,
		"UnfinishedJobsCount":  unfinishedJobsCount,
		"IdleAgentCount":       idleAgentCount,
		"BusyAgentCount":       busyAgentCount,
		"TotalAgentCount":      totalAgentCount,
	}

	res := &collector.Result{
		Totals: totals,
		Queues: map[string]map[string]int{
			"default": totals,
			"deploy":  totals,
		},
	}
	return res
}

func gatherMetrics(t *testing.T) map[string]*dto.MetricFamily {
	t.Helper()

	oldRegisterer := prometheus.DefaultRegisterer
	defer func() {
		prometheus.DefaultRegisterer = oldRegisterer
	}()
	r := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = r

	p := newPrometheus()
	p.Collect(newTestResult(t))

	if mfs, err := r.Gather(); err != nil {
		t.Fatal(err)
		return nil
	} else {
		mfsm := make(map[string]*dto.MetricFamily)
		for _, mf := range mfs {
			mfsm[*mf.Name] = mf
		}
		return mfsm
	}
}

func TestCollect(t *testing.T) {
	mfs := gatherMetrics(t)

	if want, have := 16, len(mfs); want != have {
		t.Errorf("wanted %d Prometheus metrics, have: %d", want, have)
	}

	tcs := []struct {
		Group      string
		PromName   string
		PromHelp   string
		PromLabels []string
		PromValue  float64
		PromType   dto.MetricType
	}{
		{
			"Total",
			"buildkite_total_running_jobs_count",
			"Buildkite Total: RunningJobsCount",
			[]string{},
			runningJobsCount,
			dto.MetricType_GAUGE,
		},
		{
			"Total",
			"buildkite_total_scheduled_jobs_count",
			"Buildkite Total: ScheduledJobsCount",
			[]string{},
			scheduledJobsCount,
			dto.MetricType_GAUGE,
		},
		{
			"Queues",
			"buildkite_queues_scheduled_builds_count",
			"Buildkite Queues: ScheduledBuildsCount",
			[]string{"default", "deploy"},
			scheduledBuildsCount,
			dto.MetricType_GAUGE,
		},
		{
			"Queues",
			"buildkite_queues_idle_agent_count",
			"Buildkite Queues: IdleAgentCount",
			[]string{"default", "deploy"},
			idleAgentCount,
			dto.MetricType_GAUGE,
		},
	}

	for _, tc := range tcs {
		t.Run(fmt.Sprintf("%s/%s", tc.Group, tc.PromName), func(t *testing.T) {
			mf, ok := mfs[tc.PromName]
			if !ok {
				t.Errorf("no metric found for name %s", tc.PromName)
			}

			if want, have := tc.PromHelp, mf.GetHelp(); want != have {
				t.Errorf(wantHaveFmt, want, have)
			}

			if want, have := tc.PromType, mf.GetType(); want != have {
				t.Errorf(wantHaveFmt, want, have)
			}

			ms := mf.GetMetric()
			for i, m := range ms {
				if want, have := tc.PromValue, m.GetGauge().GetValue(); want != have {
					t.Errorf(wantHaveFmt, want, have)
				}

				if len(tc.PromLabels) > 0 {
					if want, have := tc.PromLabels[i], m.Label[0].GetValue(); want != have {
						t.Errorf(wantHaveFmt, want, have)
					}
				}
			}

		})
	}
}

func TestCamelToUnderscore(t *testing.T) {
	tcs := []struct {
		Camel      string
		Underscore string
	}{
		{"TotalAgentCount", "total_agent_count"},
		{"Total@#4JobsCount", "total@#4_jobs_count"},
		{"BuildkiteQueuesIdleAgentCount1_11", "buildkite_queues_idle_agent_count1_11"},
	}

	for _, tc := range tcs {
		if want, have := tc.Underscore, camelToUnderscore(tc.Camel); want != have {
			t.Errorf(wantHaveFmt, want, have)
		}
	}
}
