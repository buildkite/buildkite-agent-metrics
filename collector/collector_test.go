package collector

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCollectorWithEmptyResponseForAllQueues(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, `{"organization": {"slug": "org-name"}}`)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	c := &Collector{
		Endpoint: s.URL,
		Token: "abc123",
		UserAgent: "some-client/1.2.3",
	}
	res, err := c.Collect()
	if err != nil {
		t.Fatal(err)
	}
	testCases := []struct {
		Group    string
		Counts   map[string]int
		Key      string
		Expected int
	}{
		{"Totals", res.Totals, RunningJobsCount, 0},
		{"Totals", res.Totals, ScheduledJobsCount, 0},
		{"Totals", res.Totals, UnfinishedJobsCount, 0},
		{"Totals", res.Totals, TotalAgentCount, 0},
		{"Totals", res.Totals, BusyAgentCount, 0},
		{"Totals", res.Totals, IdleAgentCount, 0},
		{"Totals", res.Totals, BusyAgentPercentage, 0},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s/%s", tc.Group, tc.Key), func(t *testing.T) {
			if tc.Counts[tc.Key] != tc.Expected {
				t.Fatalf("%s was %d; want %d", tc.Key, tc.Counts[tc.Key], tc.Expected)
			}
		})
	}

	if len(res.Queues) > 0 {
		t.Fatalf("Unexpected queues in response: %v", res.Queues)
	}
}

func TestCollectorWithNoJobsForAllQueues(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, `{"organization": {"slug": "org-name"}, "jobs":{"scheduled":0,"running":0,"all":0,"queues":{}},"agents":{"idle":0,"busy":0,"all":0,"queues":{}}}`)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	c := &Collector{
		Endpoint: s.URL,
		Token: "abc123",
		UserAgent: "some-client/1.2.3",
	}
	res, err := c.Collect()
	if err != nil {
		t.Fatal(err)
	}
	testCases := []struct {
		Group    string
		Counts   map[string]int
		Key      string
		Expected int
	}{
		{"Totals", res.Totals, RunningJobsCount, 0},
		{"Totals", res.Totals, ScheduledJobsCount, 0},
		{"Totals", res.Totals, UnfinishedJobsCount, 0},
		{"Totals", res.Totals, TotalAgentCount, 0},
		{"Totals", res.Totals, BusyAgentCount, 0},
		{"Totals", res.Totals, IdleAgentCount, 0},
		{"Totals", res.Totals, BusyAgentPercentage, 0},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s/%s", tc.Group, tc.Key), func(t *testing.T) {
			if tc.Counts[tc.Key] != tc.Expected {
				t.Fatalf("%s was %d; want %d", tc.Key, tc.Counts[tc.Key], tc.Expected)
			}
		})
	}

	if len(res.Queues) > 0 {
		t.Fatalf("Unexpected queues in response: %v", res.Queues)
	}
}

func TestCollectorWithSomeJobsAndAgentsForAllQueues(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, `{"organization": {"slug": "org-name"}, "jobs":{"scheduled":3,"running":1,"total":4,"queues":{"default":{"scheduled":2,"running":1,"total":3},"deploy":{"scheduled":1,"running":0,"total":1}}},"agents":{"idle":0,"busy":1,"total":1,"queues":{"default":{"idle":0,"busy":1,"total":1}}}}`)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	c := &Collector{
		Endpoint: s.URL,
		Token: "abc123",
		UserAgent: "some-client/1.2.3",
	}
	res, err := c.Collect()
	if err != nil {
		t.Fatal(err)
	}
	testCases := []struct {
		Group    string
		Counts   map[string]int
		Key      string
		Expected int
	}{
		{"Totals", res.Totals, RunningJobsCount, 1},
		{"Totals", res.Totals, ScheduledJobsCount, 3},
		{"Totals", res.Totals, UnfinishedJobsCount, 4},
		{"Totals", res.Totals, TotalAgentCount, 1},
		{"Totals", res.Totals, BusyAgentCount, 1},
		{"Totals", res.Totals, IdleAgentCount, 0},
		{"Totals", res.Totals, BusyAgentPercentage, 100},

		{"Queue.default", res.Queues["default"], RunningJobsCount, 1},
		{"Queue.default", res.Queues["default"], ScheduledJobsCount, 2},
		{"Queue.default", res.Queues["default"], UnfinishedJobsCount, 3},
		{"Queue.default", res.Queues["default"], TotalAgentCount, 1},
		{"Queue.default", res.Queues["default"], BusyAgentCount, 1},
		{"Queue.default", res.Queues["default"], IdleAgentCount, 0},

		{"Queue.deploy", res.Queues["deploy"], RunningJobsCount, 0},
		{"Queue.deploy", res.Queues["deploy"], ScheduledJobsCount, 1},
		{"Queue.deploy", res.Queues["deploy"], UnfinishedJobsCount, 1},
		{"Queue.deploy", res.Queues["deploy"], TotalAgentCount, 0},
		{"Queue.deploy", res.Queues["deploy"], BusyAgentCount, 0},
		{"Queue.deploy", res.Queues["deploy"], IdleAgentCount, 0},
	}

	for queue, _ := range res.Queues {
		switch queue {
		case "default", "deploy":
			continue
		default:
			t.Fatalf("Unexpected queue %s", queue)
		}
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s/%s", tc.Group, tc.Key), func(t *testing.T) {
			if tc.Counts[tc.Key] != tc.Expected {
				t.Fatalf("%s was %d; want %d", tc.Key, tc.Counts[tc.Key], tc.Expected)
			}
		})
	}
}

func TestCollectorWithSomeJobsAndAgentsForAQueue(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics/queue" && r.URL.Query().Get("name") == "deploy" {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, `{"organization": {"slug": "org-name"}, "jobs":{"scheduled":3,"running":1,"total":4},"agents":{"idle":0,"busy":1,"total":1}}`)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	c := &Collector{
		Endpoint: s.URL,
		Token: "abc123",
		UserAgent: "some-client/1.2.3",
		Queue: "deploy",
	}
	res, err := c.Collect()
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Totals) > 0 {
		t.Fatalf("Expected no Totals but found: %v", res.Totals)
	}
	testCases := []struct {
		Group    string
		Counts   map[string]int
		Key      string
		Expected int
	}{
		{"Queue.deploy", res.Queues["deploy"], RunningJobsCount, 1},
		{"Queue.deploy", res.Queues["deploy"], ScheduledJobsCount, 3},
		{"Queue.deploy", res.Queues["deploy"], UnfinishedJobsCount, 4},
		{"Queue.deploy", res.Queues["deploy"], TotalAgentCount, 1},
		{"Queue.deploy", res.Queues["deploy"], BusyAgentCount, 1},
		{"Queue.deploy", res.Queues["deploy"], IdleAgentCount, 0},
		{"Queue.deploy", res.Queues["deploy"], BusyAgentPercentage, 100},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s/%s", tc.Group, tc.Key), func(t *testing.T) {
			if tc.Counts[tc.Key] != tc.Expected {
				t.Fatalf("%s was %d; want %d", tc.Key, tc.Counts[tc.Key], tc.Expected)
			}
		})
	}
}
