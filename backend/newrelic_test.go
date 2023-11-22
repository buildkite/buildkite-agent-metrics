package backend

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestToCustomEvent(t *testing.T) {
	tcs := []struct {
		desc        string
		clusterName string
		queueName   string         // input queue
		metrics     map[string]int // input metrics
		want        map[string]any // output shaped data
	}{
		{
			desc:        "partial data",
			clusterName: "test-cluster",
			queueName:   "partial-data-test",
			metrics: map[string]int{
				"BusyAgentCount":      0,
				"BusyAgentPercentage": 0,
				"IdleAgentCount":      3,
				"TotalAgentCount":     3,
				"RunningJobsCount":    0,
			},
			want: map[string]any{
				"Cluster":             "test-cluster",
				"Queue":               "partial-data-test",
				"BusyAgentCount":      0,
				"BusyAgentPercentage": 0,
				"IdleAgentCount":      3,
				"TotalAgentCount":     3,
				"RunningJobsCount":    0,
			},
		},
		{
			desc:        "complete data",
			clusterName: "test-cluster",
			queueName:   "complete-data-test",
			metrics: map[string]int{
				"BusyAgentCount":      2,
				"BusyAgentPercentage": 20,
				"IdleAgentCount":      8,
				"TotalAgentCount":     10,
				"RunningJobsCount":    2,
				"ScheduledJobsCount":  0,
				"WaitingJobsCount":    0,
			},
			want: map[string]any{
				"Cluster":             "test-cluster",
				"Queue":               "complete-data-test",
				"BusyAgentCount":      2,
				"BusyAgentPercentage": 20,
				"IdleAgentCount":      8,
				"TotalAgentCount":     10,
				"RunningJobsCount":    2,
				"ScheduledJobsCount":  0,
				"WaitingJobsCount":    0,
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			got := toCustomEvent(tc.clusterName, tc.queueName, tc.metrics)
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("toCustomEvent output diff (-got +want):\n%s", diff)
			}
		})
	}

}
