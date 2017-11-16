package collector

import (
	"log"
	"regexp"
	"time"

	"golang.org/x/net/idna"
	bk "gopkg.in/buildkite/go-buildkite.v2/buildkite"
)

const (
	RunningBuildsCount   = "RunningBuildsCount"
	RunningJobsCount     = "RunningJobsCount"
	ScheduledBuildsCount = "ScheduledBuildsCount"
	ScheduledJobsCount   = "ScheduledJobsCount"
	UnfinishedJobsCount  = "UnfinishedJobsCount"
	TotalAgentCount      = "TotalAgentCount"
	BusyAgentCount       = "BusyAgentCount"
	IdleAgentCount       = "IdleAgentCount"
)

const recordsPerPage = 100

type Opts struct {
	OrgSlug    string
	Historical time.Duration
	Queue      string
	Debug      bool
}

type Collector struct {
	Opts

	buildService interface {
		ListByOrg(org string, opt *bk.BuildsListOptions) ([]bk.Build, *bk.Response, error)
	}
	agentService interface {
		List(org string, opt *bk.AgentListOptions) ([]bk.Agent, *bk.Response, error)
	}
}

func New(c *bk.Client, opts Opts) *Collector {
	return &Collector{
		Opts:         opts,
		buildService: c.Builds,
		agentService: c.Agents,
	}
}

func (c *Collector) Collect() (*Result, error) {
	res := &Result{
		Totals:    newCounts(),
		Queues:    map[string]map[string]int{},
		Pipelines: map[string]map[string]int{},
	}

	if c.Opts.Queue == "" {
		log.Println("Collecting historical metrics")
		if err := c.addHistoricalMetrics(res); err != nil {
			return nil, err
		}
	}

	log.Println("Collecting running and scheduled build and job metrics")
	if err := c.addBuildAndJobMetrics(res); err != nil {
		return nil, err
	}

	log.Println("Collecting agent metrics")
	if err := c.addAgentMetrics(res); err != nil {
		return nil, err
	}

	return res, nil
}

func newCounts() map[string]int {
	return map[string]int{
		RunningBuildsCount:   0,
		ScheduledBuildsCount: 0,
		RunningJobsCount:     0,
		ScheduledJobsCount:   0,
		UnfinishedJobsCount:  0,
	}
}

type Result struct {
	Totals            map[string]int
	Queues, Pipelines map[string]map[string]int
}

func (res Result) Dump() {
	for name, c := range res.Totals {
		log.Printf("Buildkite > %s = %d", name, c)
	}

	for name, c := range res.Queues {
		for k, v := range c {
			log.Printf("Buildkite > [queue = %s] > %s = %d", name, k, v)
		}
	}

	for name, c := range res.Pipelines {
		for k, v := range c {
			log.Printf("Buildkite > [pipeline = %s] > %s = %d", name, k, v)
		}
	}
}

var queuePattern = regexp.MustCompile(`(?i)^queue=(.+?)$`)

func queue(j *bk.Job) string {
	for _, m := range j.AgentQueryRules {
		if match := queuePattern.FindStringSubmatch(m); match != nil {
			return match[1]
		}
	}
	return "default"
}

func getBuildQueues(builds ...bk.Build) []string {
	queueMap := map[string]struct{}{}
	for _, b := range builds {
		for _, j := range b.Jobs {
			queueMap[queue(j)] = struct{}{}
		}
	}

	queues := []string{}
	for q := range queueMap {
		queues = append(queues, q)
	}

	return queues
}

func (c *Collector) addHistoricalMetrics(r *Result) error {
	finishedBuilds := c.listBuildsByOrg(c.Opts.OrgSlug, bk.BuildsListOptions{
		FinishedFrom: time.Now().UTC().Add(c.Opts.Historical * -1),
		ListOptions: bk.ListOptions{
			PerPage: recordsPerPage,
		},
	})

	return finishedBuilds.Pages(func(v interface{}) bool {
		for _, build := range v.([]bk.Build) {
			queues := c.filterQueues(getBuildQueues(v.([]bk.Build)...)...)

			if len(queues) == 0 {
				log.Printf("Skipping build, no jobs match queue filter %v", c.Queue)
				continue
			}

			for _, queue := range queues {
				if _, ok := r.Queues[queue]; !ok {
					r.Queues[queue] = newCounts()
				}
			}

			r.Pipelines[*build.Pipeline.Name] = newCounts()
		}
		return true
	})
}

func (c *Collector) filterQueues(queues ...string) []string {
	if c.Queue == "" {
		return queues
	}
	var filtered = []string{}
	for _, queue := range queues {
		if queue == c.Queue {
			filtered = append(filtered, queue)
		}
	}
	return filtered
}

func (c *Collector) addBuildAndJobMetrics(r *Result) error {
	currentBuilds := c.listBuildsByOrg(c.Opts.OrgSlug, bk.BuildsListOptions{
		State: []string{"scheduled", "running"},
		ListOptions: bk.ListOptions{
			PerPage: recordsPerPage,
		},
	})

	return currentBuilds.Pages(func(v interface{}) bool {
		for _, build := range v.([]bk.Build) {
			if c.Opts.Debug {
				log.Printf("Processing build (id=%q, pipeline=%q, branch=%q, state=%q)",
					*build.ID, *build.Pipeline.Name, *build.Branch, *build.State)
			}

			if filtered := c.filterQueues(getBuildQueues(build)...); len(filtered) == 0 {
				log.Printf("Skipping build, no jobs match queue filter %v", c.Queue)
				continue
			}

			pipeline, ucErr := idna.ToASCII(*build.Pipeline.Name)

			if ucErr != nil {
				log.Printf("Error converting pipeline name '%s' to ASCII: %s", *build.Pipeline.Name, ucErr)
				continue
			}

			if _, ok := r.Pipelines[pipeline]; !ok {
				r.Pipelines[pipeline] = newCounts()
			}

			switch *build.State {
			case "running":
				r.Totals[RunningBuildsCount]++
				r.Pipelines[pipeline][RunningBuildsCount]++

			case "scheduled":
				r.Totals[ScheduledBuildsCount]++
				r.Pipelines[pipeline][ScheduledBuildsCount]++
			}

			var buildQueues = map[string]int{}

			for _, job := range build.Jobs {
				if job.Type != nil && *job.Type == "waiter" {
					continue
				}

				state := ""
				if job.State != nil {
					state = *job.State
				}

				if c.Opts.Debug {
					log.Printf("Adding job to stats (id=%q, pipeline=%q, queue=%q, type=%q, state=%q)",
						*job.ID, *build.Pipeline.Name, queue(job), *job.Type, state)
				}

				if filtered := c.filterQueues(queue(job)); len(filtered) == 0 {
					log.Printf("Skipping job, doesn't match queue filter %v", c.Queue)
					continue
				}

				if _, ok := r.Queues[queue(job)]; !ok {
					r.Queues[queue(job)] = newCounts()
				}

				if state == "running" || state == "scheduled" {
					switch state {
					case "running":
						r.Totals[RunningJobsCount]++
						r.Queues[queue(job)][RunningJobsCount]++
						r.Pipelines[pipeline][RunningJobsCount]++

					case "scheduled":
						r.Totals[ScheduledJobsCount]++
						r.Queues[queue(job)][ScheduledJobsCount]++
						r.Pipelines[pipeline][ScheduledJobsCount]++
					}

					r.Totals[UnfinishedJobsCount]++
					r.Queues[queue(job)][UnfinishedJobsCount]++
					r.Pipelines[pipeline][UnfinishedJobsCount]++
				}

				buildQueues[queue(job)]++
			}

			// add build metrics to queues
			if len(buildQueues) > 0 {
				for queue := range buildQueues {
					switch *build.State {
					case "running":
						r.Queues[queue][RunningBuildsCount]++

					case "scheduled":
						r.Queues[queue][ScheduledBuildsCount]++
					}
				}
			}
		}
		return true
	})
}

func (c *Collector) addAgentMetrics(r *Result) error {
	p := &pager{
		lister: func(page int) (interface{}, int, error) {
			agents, resp, err := c.agentService.List(c.Opts.OrgSlug, &bk.AgentListOptions{
				ListOptions: bk.ListOptions{
					Page: page,
				},
			})
			if err != nil {
				return nil, 0, err
			}
			return agents, resp.NextPage, err
		},
	}

	r.Totals[BusyAgentCount] = 0
	r.Totals[IdleAgentCount] = 0
	r.Totals[TotalAgentCount] = 0

	for queue := range r.Queues {
		if filtered := c.filterQueues(queue); len(filtered) > 0 {
			r.Queues[queue][BusyAgentCount] = 0
			r.Queues[queue][IdleAgentCount] = 0
			r.Queues[queue][TotalAgentCount] = 0
		}
	}

	err := p.Pages(func(v interface{}) bool {
		agents := v.([]bk.Agent)

		for _, agent := range agents {
			queue := "default"
			for _, m := range agent.Metadata {
				if match := queuePattern.FindStringSubmatch(m); match != nil {
					queue = match[1]
					break
				}
			}

			if filtered := c.filterQueues(queue); len(filtered) == 0 {
				log.Printf("Skipping agent, doesn't match queue filter %v", c.Queue)
				continue
			}

			if _, ok := r.Queues[queue]; !ok {
				r.Queues[queue] = newCounts()
				r.Queues[queue][BusyAgentCount] = 0
				r.Queues[queue][IdleAgentCount] = 0
				r.Queues[queue][TotalAgentCount] = 0
			}

			if c.Opts.Debug {
				log.Printf("Adding agent to stats (name=%q, queue=%q, job=%#v)",
					*agent.Name, queue, agent.Job != nil)
			}

			if agent.Job != nil {
				r.Totals[BusyAgentCount]++
				r.Queues[queue][BusyAgentCount]++
			} else {
				r.Totals[IdleAgentCount]++
				r.Queues[queue][IdleAgentCount]++
			}

			r.Totals[TotalAgentCount]++
			r.Queues[queue][TotalAgentCount]++
		}

		return true
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *Collector) listBuildsByOrg(orgSlug string, opts bk.BuildsListOptions) *pager {
	return &pager{
		lister: func(page int) (interface{}, int, error) {
			opts.ListOptions = bk.ListOptions{
				Page: page,
			}
			builds, resp, err := c.buildService.ListByOrg(orgSlug, &opts)
			if err != nil {
				return nil, 0, err
			}
			return builds, resp.NextPage, err
		},
	}
}

type pager struct {
	lister func(page int) (v interface{}, nextPage int, err error)
}

func (p *pager) Pages(f func(v interface{}) bool) error {
	page := 1
	for {
		val, nextPage, err := p.lister(page)
		if err != nil {
			return err
		}
		if !f(val) || nextPage == 0 {
			break
		}
		page = nextPage
	}
	return nil
}
