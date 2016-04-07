package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"gopkg.in/buildkite/go-buildkite.v2/buildkite"
)

// Version is passed in via ldflags
var Version string

var queuePattern *regexp.Regexp

func init() {
	queuePattern = regexp.MustCompile(`(?i)^queue=(.+?)$`)
}

// Generates:

// Buildkite > RunningBuildsCount
// Buildkite > RunningJobsCount
// Buildkite > ScheduledBuildsCount
// Buildkite > ScheduledJobsCount
// Buildkite > IdleAgentsCount
// Buildkite > BusyAgentsCount
// Buildkite > TotalAgentsCount

// Buildkite > (Queue) > RunningBuildsCount
// Buildkite > (Queue) > RunningJobsCount
// Buildkite > (Queue) > ScheduledBuildsCount
// Buildkite > (Queue) > ScheduledJobsCount
// Buildkite > (Queue) > IdleAgentsCount
// Buildkite > (Queue) > BusyAgentsCount
// Buildkite > (Queue) > TotalAgentsCount

// Buildkite > (Pipeline) > RunningBuildsCount
// Buildkite > (Pipeline) > RunningJobsCount
// Buildkite > (Pipeline) > ScheduledBuildsCount
// Buildkite > (Pipeline) > ScheduledJobsCount

func main() {
	var (
		accessToken = flag.String("token", "", "A Buildkite API Access Token")
		orgSlug     = flag.String("org", "", "A Buildkite Organization Slug")
		interval    = flag.Duration("interval", 0, "Update metrics every interval, rather than once")
		debug       = flag.Bool("debug", false, "Show API debugging output")
		version     = flag.Bool("version", false, "Show the version")
	)

	flag.Parse()

	if *version {
		fmt.Printf("buildkite-cloudwatch-metrics %s", Version)
		os.Exit(0)
	}

	if *accessToken == "" {
		log.Fatal("Must provide a value for -token")
	}

	if *orgSlug == "" {
		log.Fatal("Must provide a value for -org")
	}

	config, err := buildkite.NewTokenConfig(*accessToken, false)
	if err != nil {
		log.Fatalf("client config failed: %s", err)
	}

	client := buildkite.NewClient(config.Client())
	buildkite.SetHttpDebug(*debug)

	if err := runCollector(client, *orgSlug, time.Hour*24); err != nil {
		log.Fatal(err)
	}

	if *interval > 0 {
		for _ = range time.NewTicker(*interval).C {
			if err := runCollector(client, *orgSlug, time.Hour); err != nil {
				log.Println(err)
			}
		}
	}
}

func runCollector(client *buildkite.Client, orgSlug string, historical time.Duration) error {
	svc := cloudwatch.New(session.New())

	res := &result{
		totals:    newCounts(),
		queues:    map[string]counts{},
		pipelines: map[string]counts{},
	}

	log.Printf("Collecting buildkite metrics from org %s", orgSlug)
	if err := res.addBuildAndJobMetrics(client, orgSlug, historical); err != nil {
		return err
	}

	if err := res.addAgentMetrics(client, orgSlug, historical); err != nil {
		return err
	}

	metrics := res.toMetrics()
	log.Printf("Extracted %d cloudwatch metrics from results", len(metrics))

	for _, metric := range metrics {
		ds := []string{}
		for _, d := range metric.Dimensions {
			ds = append(ds, *d.Name+"="+*d.Value)
		}

		path := []string{"Buildkite"}
		if len(ds) > 0 {
			path = append(path, strings.Join(ds, ","))
		}

		log.Printf("%s > %s = %.0f",
			strings.Join(path, " > "), *metric.MetricName, *metric.Value)
	}

	for _, chunk := range chunkMetricData(10, metrics) {
		log.Printf("Submitting chunk of %d metrics to Cloudwatch", len(chunk))
		if err := putMetricData(svc, chunk); err != nil {
			if err != nil {
				return err
			}
		}
	}

	return nil
}

const (
	runningBuildsCount   = "RunningBuildsCount"
	runningJobsCount     = "RunningJobsCount"
	scheduledBuildsCount = "ScheduledBuildsCount"
	scheduledJobsCount   = "ScheduledJobsCount"
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
	}
}

func (c counts) toMetrics(dimensions []*cloudwatch.Dimension) []*cloudwatch.MetricDatum {
	m := []*cloudwatch.MetricDatum{}

	for k, v := range c {
		m = append(m, &cloudwatch.MetricDatum{
			MetricName: aws.String(k),
			Dimensions: dimensions,
			Value:      aws.Float64(float64(v)),
			Unit:       aws.String("Count"),
		})
	}

	return m
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

func (r *result) toMetrics() []*cloudwatch.MetricDatum {
	data := []*cloudwatch.MetricDatum{}
	data = append(data, r.totals.toMetrics(nil)...)

	for name, c := range r.queues {
		data = append(data, c.toMetrics([]*cloudwatch.Dimension{
			{Name: aws.String("Queue"), Value: aws.String(name)},
		})...)
	}

	for name, c := range r.pipelines {
		data = append(data, c.toMetrics([]*cloudwatch.Dimension{
			{Name: aws.String("Pipeline"), Value: aws.String(name)},
		})...)
	}

	return data
}

func (r *result) addBuildAndJobMetrics(client *buildkite.Client, orgSlug string, historical time.Duration) error {
	// Algorithm:
	// Get Builds with finished_from = 24 hours ago
	// Build results with zero values for pipelines/queues
	// Get all running and scheduled builds, add to results

	finishedBuilds := listBuildsByOrg(client.Builds, orgSlug, buildkite.BuildsListOptions{
		FinishedFrom: time.Now().UTC().Add(historical * -1),
	})

	err := finishedBuilds.Pages(func(v interface{}, lastPage bool) bool {
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
	if err != nil {
		return err
	}

	currentBuilds := listBuildsByOrg(client.Builds, orgSlug, buildkite.BuildsListOptions{
		State: []string{"scheduled", "running"},
	})

	return currentBuilds.Pages(func(v interface{}, lastPage bool) bool {
		for _, build := range v.([]buildkite.Build) {
			log.Printf("Adding build to stats (id=%q, pipeline=%q, branch=%q, state=%q)",
				*build.ID, *build.Pipeline.Name, *build.Branch, *build.State)

			if _, ok := r.pipelines[*build.Pipeline.Name]; !ok {
				r.pipelines[*build.Pipeline.Name] = newCounts()
			}

			switch *build.State {
			case "running":
				r.totals[runningBuildsCount]++
				r.pipelines[*build.Pipeline.Name][runningBuildsCount]++

			case "scheduled":
				r.totals[scheduledBuildsCount]++
				r.pipelines[*build.Pipeline.Name][scheduledBuildsCount]++
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

				log.Printf("Adding job to stats (id=%q, pipeline=%q, queue=%q, type=%q, state=%q)",
					*job.ID, *build.Pipeline.Name, queue(job), *job.Type, state)

				if _, ok := r.queues[queue(job)]; !ok {
					r.queues[queue(job)] = newCounts()
				}

				switch state {
				case "running":
					r.totals[runningJobsCount]++
					r.queues[queue(job)][runningJobsCount]++

				case "scheduled":
					r.totals[scheduledJobsCount]++
					r.queues[queue(job)][scheduledJobsCount]++
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

func (r *result) addAgentMetrics(client *buildkite.Client, orgSlug string, historical time.Duration) error {
	p := &pager{
		lister: func(page int) (interface{}, int, error) {
			agents, resp, err := client.Agents.List(orgSlug, &buildkite.AgentListOptions{
				ListOptions: buildkite.ListOptions{
					Page: page,
				},
			})
			log.Printf("Agents page %d has %d agents, next page is %d",
				page,
				len(agents),
				resp.NextPage,
			)
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

	err := p.Pages(func(v interface{}, lastPage bool) bool {
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

			log.Printf("Adding agent to stats (name=%q, queue=%q, job=%#v)",
				*agent.Name, queue, agent.Job != nil)

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

func (p *pager) Pages(f func(v interface{}, lastPage bool) bool) error {
	page := 1
	for {
		val, nextPage, err := p.lister(page)
		if err != nil {
			return err
		}
		if !f(val, nextPage == 0) || nextPage == 0 {
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
			log.Printf("Builds page %d has %d builds, next page is %d", page, len(builds), resp.NextPage)
			return builds, resp.NextPage, err
		},
	}
}

func chunkMetricData(size int, data []*cloudwatch.MetricDatum) [][]*cloudwatch.MetricDatum {
	var chunks = [][]*cloudwatch.MetricDatum{}
	for i := 0; i < len(data); i += size {
		end := i + size
		if end > len(data) {
			end = len(data)
		}
		chunks = append(chunks, data[i:end])
	}
	return chunks
}

func putMetricData(svc *cloudwatch.CloudWatch, data []*cloudwatch.MetricDatum) error {
	_, err := svc.PutMetricData(&cloudwatch.PutMetricDataInput{
		MetricData: data,
		Namespace:  aws.String("Buildkite"),
	})
	if err != nil {
		return err
	}

	return nil
}
