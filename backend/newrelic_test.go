package backend

import (
	"reflect"
	"testing"
)

func TestToCustomEvent(t *testing.T) {
	tcs := []struct {
		queueName string                 // input queue
		metrics   map[string]int         // input metrics
		expected  map[string]interface{} // output shaped data
	}{
		// test 1 partial data
		{
			queueName: "partial-data-test",
			metrics: map[string]int{
				"BusyAgentCount":      0,
				"BusyAgentPercentage": 0,
				"IdleAgentCount":      3,
				"TotalAgentCount":     3,
				"RunningJobsCount":    0,
			},
			expected: map[string]interface{}{
				"Queue":               "partial-data-test",
				"BusyAgentCount":      0,
				"BusyAgentPercentage": 0,
				"IdleAgentCount":      3,
				"TotalAgentCount":     3,
				"RunningJobsCount":    0,
			},
		},
		// test 2 complete data
		{
			queueName: "complete-data-test",
			metrics: map[string]int{
				"BusyAgentCount":      2,
				"BusyAgentPercentage": 20,
				"IdleAgentCount":      8,
				"TotalAgentCount":     10,
				"RunningJobsCount":    2,
				"ScheduledJobsCount":  0,
				"WaitingJobsCount":    0,
			},
			expected: map[string]interface{}{
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

	for n, tc := range tcs {
		got := toCustomEvent(tc.queueName, tc.metrics)

		if !reflect.DeepEqual(got, tc.expected) {
			t.Errorf("toCustomEvent test #%d failed, result %+v did not equal expected %+v", n, got, tc.expected)
		}
	}

}
