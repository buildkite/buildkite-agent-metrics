package collector

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptrace"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	ScheduledJobsCount  = "ScheduledJobsCount"
	RunningJobsCount    = "RunningJobsCount"
	UnfinishedJobsCount = "UnfinishedJobsCount"
	WaitingJobsCount    = "WaitingJobsCount"
	IdleAgentCount      = "IdleAgentCount"
	BusyAgentCount      = "BusyAgentCount"
	TotalAgentCount     = "TotalAgentCount"
	BusyAgentPercentage = "BusyAgentPercentage"

	PollDurationHeader = `Buildkite-Agent-Metrics-Poll-Duration`
)

var (
	ErrUnauthorized = errors.New("unauthorized")
)

type Collector struct {
	Endpoint  string
	Token     string
	UserAgent string
	Queues    []string
	Quiet     bool
	Debug     bool
	DebugHttp bool
	Timeout   int
}

type Result struct {
	Totals       map[string]int
	Queues       map[string]map[string]int
	Org          string
	Cluster      string
	PollDuration time.Duration
}

type organizationResponse struct {
	Slug string `json:"slug"`
}

type clusterResponse struct {
	Name string `json:"name"`
}

type metricsAgentsResponse struct {
	Idle  int `json:"idle"`
	Busy  int `json:"busy"`
	Total int `json:"total"`
}

type metricsJobsResponse struct {
	Scheduled int `json:"scheduled"`
	Running   int `json:"running"`
	Waiting   int `json:"waiting"`
	Total     int `json:"total"`
}

type queueMetricsResponse struct {
	Agents       metricsAgentsResponse `json:"agents"`
	Jobs         metricsJobsResponse   `json:"jobs"`
	Organization organizationResponse  `json:"organization"`
	Cluster      clusterResponse       `json:"cluster"`
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
	Cluster      clusterResponse          `json:"cluster"`
}

func (c *Collector) Collect() (*Result, error) {
	result := &Result{
		Totals: map[string]int{},
		Queues: map[string]map[string]int{},
	}

	httpClient := &http.Client{
		Timeout: time.Duration(c.Timeout) * time.Second,
	}

	if len(c.Queues) == 0 {
		if err := c.collectAllQueues(httpClient, result); err != nil {
			return nil, err
		}
	} else {
		for _, queue := range c.Queues {
			if err := c.collectQueue(httpClient, result, queue); err != nil {
				return nil, err
			}
		}
	}

	if !c.Quiet {
		result.Dump()
	}

	return result, nil
}

func (c *Collector) collectAllQueues(httpClient *http.Client, result *Result) error {
	log.Println("Collecting agent metrics for all queues")

	endpoint, err := url.Parse(c.Endpoint)
	if err != nil {
		return err
	}

	endpoint.Path += "/metrics"

	req, err := http.NewRequest("GET", endpoint.String(), nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.Token))

	if c.DebugHttp {
		trace := &httptrace.ClientTrace{
			DNSStart: func(_ httptrace.DNSStartInfo) {
				fmt.Printf("dns start: %v\n", time.Now())
			},
			DNSDone: func(_ httptrace.DNSDoneInfo) {
				fmt.Printf("dns done: %v\n", time.Now())
			},
			ConnectStart: func(_, _ string) {
				fmt.Printf("connection start: %v\n", time.Now())
			},
			ConnectDone: func(_, _ string, _ error) {
				fmt.Printf("connection done: %v\n", time.Now())
			},
			TLSHandshakeStart: func() {
				fmt.Printf("TLS Handshake start: %v\n", time.Now())
			},
			TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
				fmt.Printf("TLS Handshake done: %v\n", time.Now())
			},
			WroteHeaders: func() {
				fmt.Printf("wrote headers: %v\n", time.Now())
			},
			WroteRequest: func(_ httptrace.WroteRequestInfo) {
				fmt.Printf("wrote request: %v\n", time.Now())
			},
		}

		req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
		if dump, err := httputil.DumpRequest(req, true); err == nil {
			log.Printf("DEBUG request uri=%s\n%s\n", req.URL, dump)
		}
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == 401 {
		return fmt.Errorf("http 401 response received %w", ErrUnauthorized)
	}

	if c.DebugHttp {
		if dump, err := httputil.DumpResponse(res, true); err == nil {
			log.Printf("DEBUG response uri=%s\n%s\n", req.URL, dump)
		}
	}

	// Handle any errors
	if res.StatusCode != http.StatusOK {
		// If it's json response, show the error message
		if strings.HasPrefix(res.Header.Get("Content-Type"), "application/json") {
			var errStruct struct {
				Message string `json:"message"`
			}
			err := json.NewDecoder(res.Body).Decode(&errStruct)
			if err == nil {
				return errors.New(errStruct.Message)
			} else {
				log.Printf("Failed to decode error: %v", err)
			}
		}

		return fmt.Errorf("Request failed with %s (%d)", res.Status, res.StatusCode)
	}

	var allMetrics allMetricsResponse

	// Check if we get a poll duration header from server
	if pollSeconds := res.Header.Get(PollDurationHeader); pollSeconds != "" {
		pollSecondsInt, err := strconv.ParseInt(pollSeconds, 10, 64)
		if err != nil {
			log.Printf("Failed to parse %s header: %v", PollDurationHeader, err)
		} else {
			result.PollDuration = time.Duration(pollSecondsInt) * time.Second
		}
	}

	err = json.NewDecoder(res.Body).Decode(&allMetrics)
	if err != nil {
		return err
	}

	if allMetrics.Organization.Slug == "" {
		return fmt.Errorf("No organization slug was found in the metrics response")
	}

	log.Printf("Found organization %q, cluster %q", allMetrics.Organization.Slug, allMetrics.Cluster.Name)
	result.Org = allMetrics.Organization.Slug
	result.Cluster = allMetrics.Cluster.Name

	result.Totals[ScheduledJobsCount] = allMetrics.Jobs.Scheduled
	result.Totals[RunningJobsCount] = allMetrics.Jobs.Running
	result.Totals[UnfinishedJobsCount] = allMetrics.Jobs.Total
	result.Totals[WaitingJobsCount] = allMetrics.Jobs.Waiting
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
		result.Queues[queueName][WaitingJobsCount] = queueJobMetrics.Waiting
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

	return nil
}

func (c *Collector) collectQueue(httpClient *http.Client, result *Result, queue string) error {
	log.Printf("Collecting agent metrics for queue '%s'", queue)

	endpoint, err := url.Parse(c.Endpoint)
	if err != nil {
		return err
	}

	endpoint.Path += "/metrics/queue"
	endpoint.RawQuery = url.Values{"name": {queue}}.Encode()

	req, err := http.NewRequest("GET", endpoint.String(), nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.Token))

	if c.DebugHttp {
		trace := &httptrace.ClientTrace{
			DNSStart: func(_ httptrace.DNSStartInfo) {
				fmt.Printf("dns start: %v\n", time.Now())
			},
			DNSDone: func(_ httptrace.DNSDoneInfo) {
				fmt.Printf("dns done: %v\n", time.Now())
			},
			ConnectStart: func(_, _ string) {
				fmt.Printf("connection start: %v\n", time.Now())
			},
			ConnectDone: func(_, _ string, _ error) {
				fmt.Printf("connection done: %v\n", time.Now())
			},
			TLSHandshakeStart: func() {
				fmt.Printf("TLS Handshake start: %v\n", time.Now())
			},
			TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
				fmt.Printf("TLS Handshake done: %v\n", time.Now())
			},
			WroteHeaders: func() {
				fmt.Printf("wrote headers: %v\n", time.Now())
			},
			WroteRequest: func(_ httptrace.WroteRequestInfo) {
				fmt.Printf("wrote request: %v\n", time.Now())
			},
		}

		req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
		if dump, err := httputil.DumpRequest(req, true); err == nil {
			log.Printf("DEBUG request uri=%s\n%s\n", req.URL, dump)
		}
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == 401 {
		return fmt.Errorf("http 401 response received %w", ErrUnauthorized)
	}

	if c.DebugHttp {
		if dump, err := httputil.DumpResponse(res, true); err == nil {
			log.Printf("DEBUG response uri=%s\n%s\n", req.URL, dump)
		}
	}

	// Handle any errors
	if res.StatusCode != http.StatusOK {
		// If it's json response, show the error message
		if strings.HasPrefix(res.Header.Get("Content-Type"), "application/json") {
			var errStruct struct {
				Message string `json:"message"`
			}
			err := json.NewDecoder(res.Body).Decode(&errStruct)
			if err == nil {
				return errors.New(errStruct.Message)
			} else {
				log.Printf("Failed to decode error: %v", err)
			}
		}

		return fmt.Errorf("Request failed with %s (%d)", res.Status, res.StatusCode)
	}

	var queueMetrics queueMetricsResponse
	err = json.NewDecoder(res.Body).Decode(&queueMetrics)
	if err != nil {
		return err
	}

	if queueMetrics.Organization.Slug == "" {
		return fmt.Errorf("No organization slug was found in the metrics response")
	}

	log.Printf("Found organization %q, cluster %q", queueMetrics.Organization.Slug, queueMetrics.Cluster.Name)
	result.Org = queueMetrics.Organization.Slug
	result.Cluster = queueMetrics.Cluster.Name

	result.Queues[queue] = map[string]int{
		ScheduledJobsCount:  queueMetrics.Jobs.Scheduled,
		RunningJobsCount:    queueMetrics.Jobs.Running,
		UnfinishedJobsCount: queueMetrics.Jobs.Total,
		WaitingJobsCount:    queueMetrics.Jobs.Waiting,
		IdleAgentCount:      queueMetrics.Agents.Idle,
		BusyAgentCount:      queueMetrics.Agents.Busy,
		TotalAgentCount:     queueMetrics.Agents.Total,
		BusyAgentPercentage: busyAgentPercentage(queueMetrics.Agents),
	}
	return nil
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
