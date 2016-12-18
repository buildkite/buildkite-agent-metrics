package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"time"

	"gopkg.in/buildkite/go-buildkite.v2/buildkite"
)

const recordsPerPage = 100

// Version is passed in via ldflags
var Version string

var queuePattern *regexp.Regexp

func init() {
	queuePattern = regexp.MustCompile(`(?i)^queue=(.+?)$`)
}

func main() {
	var (
		accessToken = flag.String("token", "", "A Buildkite API Access Token")
		orgSlug     = flag.String("org", "", "A Buildkite Organization Slug")
		interval    = flag.Duration("interval", 0, "Update metrics every interval, rather than once")
		history     = flag.Duration("history", time.Hour*24, "Historical data to use for finished builds")
		debug       = flag.Bool("debug", false, "Show debug output")
		version     = flag.Bool("version", false, "Show the version")
		quiet       = flag.Bool("quiet", false, "Only print errors")
		dryRun      = flag.Bool("dry-run", false, "Whether to only print metrics")

		// filters
		queue = flag.String("queue", "", "Only include a specific queue")
	)

	flag.Parse()

	if *version {
		fmt.Printf("buildkite-metrics %s\n", Version)
		os.Exit(0)
	}

	if *accessToken == "" {
		fmt.Println("Must provide a value for -token")
		os.Exit(1)
	}

	if *orgSlug == "" {
		fmt.Println("Must provide a value for -org")
		os.Exit(1)
	}

	if *quiet {
		log.SetOutput(ioutil.Discard)
	}

	config, err := buildkite.NewTokenConfig(*accessToken, false)
	if err != nil {
		fmt.Printf("client config failed: %s\n", err)
		os.Exit(1)
	}

	client := buildkite.NewClient(config.Client())
	if *debug && os.Getenv("TRACE_HTTP") != "" {
		buildkite.SetHttpDebug(*debug)
	}

	f := func() error {
		t := time.Now()

		res, err := collectResults(client, collectOpts{
			OrgSlug:    *orgSlug,
			Historical: *history,
			Queue:      *queue,
			Debug:      *debug,
		})
		if err != nil {
			return err
		}

		if !*quiet {
			dumpResults(res)
		}

		if !*dryRun {
			err = cloudwatchSend(res)
			if err != nil {
				return err
			}
		}

		log.Printf("Finished in %s", time.Now().Sub(t))
		return nil
	}

	if err := f(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if *interval > 0 {
		for _ = range time.NewTicker(*interval).C {
			if err := f(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}
}

type collectOpts struct {
	OrgSlug    string
	Historical time.Duration
	Queue      string
	Debug      bool
}

func collectResults(client *buildkite.Client, opts collectOpts) (*result, error) {
	res := &result{
		totals:    newCounts(),
		queues:    map[string]counts{},
		pipelines: map[string]counts{},
	}

	if opts.Queue == "" {
		log.Println("Collecting historical metrics")
		if err := res.addHistoricalMetrics(client, opts); err != nil {
			return nil, err
		}
	}

	log.Println("Collecting running and scheduled build and job metrics")
	if err := res.addBuildAndJobMetrics(client, opts); err != nil {
		return nil, err
	}

	log.Println("Collecting agent metrics")
	if err := res.addAgentMetrics(client, opts); err != nil {
		return nil, err
	}

	if opts.Queue != "" {
		if c, ok := res.queues[opts.Queue]; ok {
			return &result{
				queues: map[string]counts{
					opts.Queue: c,
				},
			}, nil
		}
		return &result{}, nil
	}

	return res, nil
}

func dumpResults(res *result) {
	for name, c := range res.totals {
		log.Printf("Buildkite > %s = %d", name, c)
	}

	for name, c := range res.queues {
		for k, v := range c {
			log.Printf("Buildkite > [queue = %s] > %s = %d", name, k, v)
		}
	}

	for name, c := range res.pipelines {
		for k, v := range c {
			log.Printf("Buildkite > [pipeline = %s] > %s = %d", name, k, v)
		}
	}
}

const (
	runningBuildsCount   = "RunningBuildsCount"
	runningJobsCount     = "RunningJobsCount"
	scheduledBuildsCount = "ScheduledBuildsCount"
	scheduledJobsCount   = "ScheduledJobsCount"
	unfinishedJobsCount  = "UnfinishedJobsCount"
	totalAgentCount      = "TotalAgentCount"
	busyAgentCount       = "BusyAgentCount"
	idleAgentCount       = "IdleAgentCount"
)

type counts map[string]int

func newCounts() counts {
	return counts{
		runningBuildsCount:   0,
		scheduledBuildsCount: 0,
		runningJobsCount:     0,
		scheduledJobsCount:   0,
		unfinishedJobsCount:  0,
	}
}

func queue(j *buildkite.Job) string {
	for _, m := range j.AgentQueryRules {
		if match := queuePattern.FindStringSubmatch(m); match != nil {
			return match[1]
		}
	}
	return "default"
}

func uniqueQueues(builds []buildkite.Build) []string {
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

type result struct {
	totals            counts
	queues, pipelines map[string]counts
}

func (r *result) addHistoricalMetrics(client *buildkite.Client, opts collectOpts) error {
	finishedBuilds := listBuildsByOrg(client.Builds, opts.OrgSlug, buildkite.BuildsListOptions{
		FinishedFrom: time.Now().UTC().Add(opts.Historical * -1),
		ListOptions: buildkite.ListOptions{
			PerPage: recordsPerPage,
		},
	})

	return finishedBuilds.Pages(func(v interface{}) bool {
		for _, queue := range uniqueQueues(v.([]buildkite.Build)) {
			if _, ok := r.queues[queue]; !ok {
				r.queues[queue] = newCounts()
			}
		}
		for _, build := range v.([]buildkite.Build) {
			r.pipelines[*build.Pipeline.Name] = newCounts()
		}
		return true
	})
}

func (r *result) addBuildAndJobMetrics(client *buildkite.Client, opts collectOpts) error {
	currentBuilds := listBuildsByOrg(client.Builds, opts.OrgSlug, buildkite.BuildsListOptions{
		State: []string{"scheduled", "running"},
		ListOptions: buildkite.ListOptions{
			PerPage: recordsPerPage,
		},
	})

	return currentBuilds.Pages(func(v interface{}) bool {
		for _, build := range v.([]buildkite.Build) {
			if opts.Debug {
				log.Printf("Adding build to stats (id=%q, pipeline=%q, branch=%q, state=%q)",
					*build.ID, *build.Pipeline.Name, *build.Branch, *build.State)
			}

			pipeline := *build.Pipeline.Name

			if _, ok := r.pipelines[pipeline]; !ok {
				r.pipelines[pipeline] = newCounts()
			}

			switch *build.State {
			case "running":
				r.totals[runningBuildsCount]++
				r.pipelines[pipeline][runningBuildsCount]++

			case "scheduled":
				r.totals[scheduledBuildsCount]++
				r.pipelines[pipeline][scheduledBuildsCount]++
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

				if opts.Debug {
					log.Printf("Adding job to stats (id=%q, pipeline=%q, queue=%q, type=%q, state=%q)",
						*job.ID, *build.Pipeline.Name, queue(job), *job.Type, state)
				}

				if _, ok := r.queues[queue(job)]; !ok {
					r.queues[queue(job)] = newCounts()
				}

				if state == "running" || state == "scheduled" {
					switch state {
					case "running":
						r.totals[runningJobsCount]++
						r.queues[queue(job)][runningJobsCount]++
						r.pipelines[pipeline][runningJobsCount]++

					case "scheduled":
						r.totals[scheduledJobsCount]++
						r.queues[queue(job)][scheduledJobsCount]++
						r.pipelines[pipeline][scheduledJobsCount]++
					}

					r.totals[unfinishedJobsCount]++
					r.queues[queue(job)][unfinishedJobsCount]++
					r.pipelines[pipeline][unfinishedJobsCount]++
				}

				buildQueues[queue(job)]++
			}

			// add build metrics to queues
			if len(buildQueues) > 0 {
				for queue := range buildQueues {
					switch *build.State {
					case "running":
						r.queues[queue][runningBuildsCount]++

					case "scheduled":
						r.queues[queue][scheduledBuildsCount]++
					}
				}
			}
		}
		return true
	})
}

func (r *result) addAgentMetrics(client *buildkite.Client, opts collectOpts) error {
	p := &pager{
		lister: func(page int) (interface{}, int, error) {
			agents, resp, err := client.Agents.List(opts.OrgSlug, &buildkite.AgentListOptions{
				ListOptions: buildkite.ListOptions{
					Page: page,
				},
			})
			if err != nil {
				return nil, 0, err
			}
			return agents, resp.NextPage, err
		},
	}

	r.totals[busyAgentCount] = 0
	r.totals[idleAgentCount] = 0
	r.totals[totalAgentCount] = 0

	for queue := range r.queues {
		r.queues[queue][busyAgentCount] = 0
		r.queues[queue][idleAgentCount] = 0
		r.queues[queue][totalAgentCount] = 0
	}

	err := p.Pages(func(v interface{}) bool {
		agents := v.([]buildkite.Agent)

		for _, agent := range agents {
			queue := "default"
			for _, m := range agent.Metadata {
				if match := queuePattern.FindStringSubmatch(m); match != nil {
					queue = match[1]
					break
				}
			}

			if _, ok := r.queues[queue]; !ok {
				r.queues[queue] = newCounts()
				r.queues[queue][busyAgentCount] = 0
				r.queues[queue][idleAgentCount] = 0
				r.queues[queue][totalAgentCount] = 0
			}

			if opts.Debug {
				log.Printf("Adding agent to stats (name=%q, queue=%q, job=%#v)",
					*agent.Name, queue, agent.Job != nil)
			}

			if agent.Job != nil {
				r.totals[busyAgentCount]++
				r.queues[queue][busyAgentCount]++
			} else {
				r.totals[idleAgentCount]++
				r.queues[queue][idleAgentCount]++
			}

			r.totals[totalAgentCount]++
			r.queues[queue][totalAgentCount]++
		}

		return true
	})
	if err != nil {
		return err
	}

	return nil
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

func listBuildsByOrg(builds *buildkite.BuildsService, orgSlug string, opts buildkite.BuildsListOptions) *pager {
	return &pager{
		lister: func(page int) (interface{}, int, error) {
			opts.ListOptions = buildkite.ListOptions{
				Page: page,
			}
			builds, resp, err := builds.ListByOrg(orgSlug, &opts)
			if err != nil {
				return nil, 0, err
			}
			return builds, resp.NextPage, err
		},
	}
}
