package collector

import (
	"fmt"
	"testing"

	bk "gopkg.in/buildkite/go-buildkite.v2/buildkite"
)

func newTestCollector() *Collector {
	return &Collector{
		Opts: Opts{},
		buildService: &testBuildService{
			[]bk.Build{
				{
					Pipeline: &bk.Pipeline{
						Name: bk.String("llamas"),
					},
					State: bk.String("running"),
					Jobs: []*bk.Job{
						{
							State:           bk.String("running"),
							AgentQueryRules: []string{"queue=default"},
						},
						{
							State:           bk.String("scheduled"),
							AgentQueryRules: []string{"queue=deploy"},
						},
					},
				},
				{
					Pipeline: &bk.Pipeline{
						Name: bk.String("alpacas"),
					},
					State: bk.String("scheduled"),
					Jobs: []*bk.Job{
						{
							State:           bk.String("scheduled"),
							AgentQueryRules: []string{"queue=default"},
						},
					},
				},
			},
		},
		agentService: &testAgentService{
			[]bk.Agent{
				{
					Metadata: []string{"queue=default"},
					Job: &bk.Job{
						State:           bk.String("running"),
						AgentQueryRules: []string{"queue=default"},
					},
				},
			},
		},
	}
}

func TestCollectorWithRunningBuilds(t *testing.T) {
	c := newTestCollector()

	res, err := c.Collect()
	if err != nil {
		t.Fatal(err)
	}

	res.Dump()

	testCases := []struct {
		Group    string
		Counts   map[string]int
		Key      string
		Expected int
	}{
		{"Totals", res.Totals, RunningBuildsCount, 1},
		{"Totals", res.Totals, ScheduledBuildsCount, 1},
		{"Totals", res.Totals, RunningJobsCount, 1},
		{"Totals", res.Totals, ScheduledJobsCount, 2},
		{"Totals", res.Totals, UnfinishedJobsCount, 3},
		{"Totals", res.Totals, TotalAgentCount, 1},
		{"Totals", res.Totals, BusyAgentCount, 1},
		{"Totals", res.Totals, IdleAgentCount, 0},

		{"Queue.default", res.Queues["default"], RunningBuildsCount, 1},
		{"Queue.default", res.Queues["default"], ScheduledBuildsCount, 1},
		{"Queue.default", res.Queues["default"], RunningJobsCount, 1},
		{"Queue.default", res.Queues["default"], ScheduledJobsCount, 1},
		{"Queue.default", res.Queues["default"], UnfinishedJobsCount, 2},
		{"Queue.default", res.Queues["default"], TotalAgentCount, 1},
		{"Queue.default", res.Queues["default"], BusyAgentCount, 1},
		{"Queue.default", res.Queues["default"], IdleAgentCount, 0},

		{"Queue.deploy", res.Queues["deploy"], RunningBuildsCount, 1},
		{"Queue.deploy", res.Queues["deploy"], ScheduledBuildsCount, 0},
		{"Queue.deploy", res.Queues["deploy"], RunningJobsCount, 0},
		{"Queue.deploy", res.Queues["deploy"], ScheduledJobsCount, 1},
		{"Queue.deploy", res.Queues["deploy"], UnfinishedJobsCount, 1},
		{"Queue.deploy", res.Queues["deploy"], TotalAgentCount, 0},
		{"Queue.deploy", res.Queues["deploy"], BusyAgentCount, 0},
		{"Queue.deploy", res.Queues["deploy"], IdleAgentCount, 0},

		{"Pipeline.llamas", res.Pipelines["llamas"], RunningBuildsCount, 1},
		{"Pipeline.llamas", res.Pipelines["llamas"], ScheduledBuildsCount, 0},
		{"Pipeline.llamas", res.Pipelines["llamas"], RunningJobsCount, 1},
		{"Pipeline.llamas", res.Pipelines["llamas"], ScheduledJobsCount, 1},
		{"Pipeline.llamas", res.Pipelines["llamas"], UnfinishedJobsCount, 2},
		{"Pipeline.llamas", res.Pipelines["llamas"], TotalAgentCount, 0},
		{"Pipeline.llamas", res.Pipelines["llamas"], BusyAgentCount, 0},
		{"Pipeline.llamas", res.Pipelines["llamas"], IdleAgentCount, 0},

		{"Pipeline.alpacas", res.Pipelines["alpacas"], RunningBuildsCount, 0},
		{"Pipeline.alpacas", res.Pipelines["alpacas"], ScheduledBuildsCount, 1},
		{"Pipeline.alpacas", res.Pipelines["alpacas"], RunningJobsCount, 0},
		{"Pipeline.alpacas", res.Pipelines["alpacas"], ScheduledJobsCount, 1},
		{"Pipeline.alpacas", res.Pipelines["alpacas"], UnfinishedJobsCount, 1},
		{"Pipeline.alpacas", res.Pipelines["alpacas"], TotalAgentCount, 0},
		{"Pipeline.alpacas", res.Pipelines["alpacas"], BusyAgentCount, 0},
		{"Pipeline.alpacas", res.Pipelines["alpacas"], IdleAgentCount, 0},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s/%s", tc.Group, tc.Key), func(t *testing.T) {
			if tc.Counts[tc.Key] != tc.Expected {
				t.Fatalf("%s was %d; want %d", tc.Key, tc.Counts[tc.Key], tc.Expected)
			}
		})
	}
}

type testBuildService struct {
	Builds []bk.Build
}

func (bs *testBuildService) ListByOrg(org string, opt *bk.BuildsListOptions) ([]bk.Build, *bk.Response, error) {
	nextPage := opt.Page + 1
	if nextPage > len(bs.Builds) {
		nextPage = 0
	}
	return []bk.Build{bs.Builds[opt.Page-1]}, &bk.Response{NextPage: nextPage}, nil
}

type testAgentService struct {
	Agents []bk.Agent
}

func (as *testAgentService) List(org string, opt *bk.AgentListOptions) ([]bk.Agent, *bk.Response, error) {
	return as.Agents, &bk.Response{NextPage: 0}, nil
}
