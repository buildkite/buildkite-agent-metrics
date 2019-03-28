package collector

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

const (
	ScheduledJobsCount  = "ScheduledJobsCount"
	RunningJobsCount    = "RunningJobsCount"
	UnfinishedJobsCount = "UnfinishedJobsCount"
	IdleAgentCount      = "IdleAgentCount"
	BusyAgentCount      = "BusyAgentCount"
	TotalAgentCount     = "TotalAgentCount"
	BusyAgentPercentage = "BusyAgentPercentage"
)

type Collector struct {
	Endpoint  string
	Token     string
	UserAgent string
	Queue     string
	Quiet     bool
	Debug     bool
	DebugHttp bool
}

type Result struct {
	Totals map[string]int
	Queues map[string]map[string]int
	Org    string
}

type organizationResponse struct {
	Slug string `json:"slug"`
}

type metricsAgentsResponse struct {
	Idle  int `json:"idle"`
	Busy  int `json:"busy"`
	Total int `json:"total"`
}

type metricsJobsResponse struct {
	Scheduled int `json:"scheduled"`
	Running   int `json:"running"`
	Total     int `json:"total"`
}

type queueMetricsResponse struct {
	Agents       metricsAgentsResponse `json:"agents"`
	Jobs         metricsJobsResponse   `json:"jobs"`
	Organization organizationResponse  `json:"organization"`
}

type allMetricsAgentsResponse struct {
	metricsAgentsResponse
	Queues map[string]metricsAgentsResponse `json:"queues"`
}

type allMetricsJobsResponse struct {
	metricsJobsResponse
	Queues map[string]metricsJobsResponse `json:"queues"`
}

type allMetricsResponse struct {
	Agents       allMetricsAgentsResponse `json:"agents"`
	Jobs         allMetricsJobsResponse   `json:"jobs"`
	Organization organizationResponse     `json:"organization"`
}

func (c *Collector) Collect() (*Result, error) {
	result := &Result{
		Totals: map[string]int{},
		Queues: map[string]map[string]int{},
	}

	if c.Queue == "" {
		log.Println("Collecting agent metrics for all queues")

		endpoint, err := url.Parse(c.Endpoint)
		if err != nil {
			return nil, err
		}

		endpoint.Path += "/metrics"

		req, err := http.NewRequest("GET", endpoint.String(), nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("User-Agent", c.UserAgent)
		req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.Token))

		if c.DebugHttp {
			if dump, err := httputil.DumpRequest(req, true); err == nil {
				log.Printf("DEBUG request uri=%s\n%s\n", req.URL, dump)
			}
		}

		httpClient := &http.Client{
			Timeout: 5 * time.Second,
		}

		res, err := httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		if c.DebugHttp {
			if dump, err := httputil.DumpResponse(res, true); err == nil {
				log.Printf("DEBUG response uri=%s\n%s\n", req.URL, dump)
			}
		}

		var allMetrics allMetricsResponse
		defer res.Body.Close()
		err = json.NewDecoder(res.Body).Decode(&allMetrics)
		if err != nil {
			return nil, err
		}

		if allMetrics.Organization.Slug == "" {
			return nil, fmt.Errorf("No organization slug was found in the metrics response")
		}

		log.Printf("Found organization %q", allMetrics.Organization.Slug)
		result.Org = allMetrics.Organization.Slug

		result.Totals[ScheduledJobsCount] = allMetrics.Jobs.Scheduled
		result.Totals[RunningJobsCount] = allMetrics.Jobs.Running
		result.Totals[UnfinishedJobsCount] = allMetrics.Jobs.Total
		result.Totals[IdleAgentCount] = allMetrics.Agents.Idle
		result.Totals[BusyAgentCount] = allMetrics.Agents.Busy
		result.Totals[TotalAgentCount] = allMetrics.Agents.Total
		result.Totals[BusyAgentPercentage] = busyAgentPercentage(allMetrics.Agents.metricsAgentsResponse)

		for queueName, queueJobMetrics := range allMetrics.Jobs.Queues {
			if _, ok := result.Queues[queueName]; !ok {
				result.Queues[queueName] = map[string]int{}
			}
			result.Queues[queueName][ScheduledJobsCount] = queueJobMetrics.Scheduled
			result.Queues[queueName][RunningJobsCount] = queueJobMetrics.Running
			result.Queues[queueName][UnfinishedJobsCount] = queueJobMetrics.Total
		}

		for queueName, queueAgentMetrics := range allMetrics.Agents.Queues {
			if _, ok := result.Queues[queueName]; !ok {
				result.Queues[queueName] = map[string]int{}
			}
			result.Queues[queueName][IdleAgentCount] = queueAgentMetrics.Idle
			result.Queues[queueName][BusyAgentCount] = queueAgentMetrics.Busy
			result.Queues[queueName][TotalAgentCount] = queueAgentMetrics.Total
			result.Queues[queueName][BusyAgentPercentage] = busyAgentPercentage(queueAgentMetrics)
		}
	} else {
		log.Printf("Collecting agent metrics for queue '%s'", c.Queue)

		endpoint, err := url.Parse(c.Endpoint)
		if err != nil {
			return nil, err
		}

		endpoint.Path += "/metrics/queue"
		endpoint.RawQuery = url.Values{"name": {c.Queue}}.Encode()

		req, err := http.NewRequest("GET", endpoint.String(), nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("User-Agent", c.UserAgent)
		req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.Token))

		if c.DebugHttp {
			if dump, err := httputil.DumpRequest(req, true); err == nil {
				log.Printf("DEBUG request uri=%s\n%s\n", req.URL, dump)
			}
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		if c.DebugHttp {
			if dump, err := httputil.DumpResponse(res, true); err == nil {
				log.Printf("DEBUG response uri=%s\n%s\n", req.URL, dump)
			}
		}

		var queueMetrics queueMetricsResponse
		defer res.Body.Close()
		err = json.NewDecoder(res.Body).Decode(&queueMetrics)
		if err != nil {
			return nil, err
		}

		if queueMetrics.Organization.Slug == "" {
			return nil, fmt.Errorf("No organization slug was found in the metrics response")
		}

		log.Printf("Found organization %q", queueMetrics.Organization.Slug)
		result.Org = queueMetrics.Organization.Slug

		result.Queues[c.Queue] = map[string]int{
			ScheduledJobsCount:  queueMetrics.Jobs.Scheduled,
			RunningJobsCount:    queueMetrics.Jobs.Running,
			UnfinishedJobsCount: queueMetrics.Jobs.Total,
			IdleAgentCount:      queueMetrics.Agents.Idle,
			BusyAgentCount:      queueMetrics.Agents.Busy,
			TotalAgentCount:     queueMetrics.Agents.Total,
			BusyAgentPercentage: busyAgentPercentage(queueMetrics.Agents),
		}
	}

	if !c.Quiet {
		result.Dump()
	}

	return result, nil
}

func busyAgentPercentage(agents metricsAgentsResponse) int {
	if agents.Total > 0 {
		return int(100 * agents.Busy / agents.Total)
	}
	return 0
}

func (r Result) Dump() {
	for name, c := range r.Totals {
		log.Printf("Buildkite > Org=%s > %s=%d", r.Org, name, c)
	}

	for name, c := range r.Queues {
		for k, v := range c {
			log.Printf("Buildkite > Org=%s > Queue=%s > %s=%d", r.Org, name, k, v)
		}
	}
}
