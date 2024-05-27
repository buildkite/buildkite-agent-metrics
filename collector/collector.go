package collector

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"maps"
	"net/http"
	"net/http/httptrace"
	"net/http/httputil"
	"net/textproto"
	"net/url"
	"os"
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

var ErrUnauthorized = errors.New("unauthorized")

var traceLog = log.New(os.Stderr, "TRACE", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.Lmsgprefix)

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
		req = traceHTTPRequest(req)
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
			traceLog.Printf("response uri=%s\n%s", req.URL, dump)
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
		req = traceHTTPRequest(req)
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
			traceLog.Printf("response uri=%s\n%s", req.URL, dump)
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

func traceHTTPRequest(req *http.Request) *http.Request {
	trace := &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			traceLog.Printf("getting connection to %s\n", hostPort)
		},
		GotConn: func(info httptrace.GotConnInfo) {
			traceLog.Printf("got connection [reused?: %t, idle?: %t (time: %v)]", info.Reused, info.WasIdle, info.IdleTime)
		},
		PutIdleConn: func(err error) {
			if err != nil {
				traceLog.Printf("Failed to put connection idle with error - %v", err)
				return
			}
			traceLog.Printf("connection successfully put idle")
		},
		GotFirstResponseByte: func() {
			traceLog.Printf("received first response byte")
		},
		Got1xxResponse: func(code int, header textproto.MIMEHeader) error {
			traceLog.Printf("received %d with header %v", code, header)
			return nil
		},
		DNSStart: func(_ httptrace.DNSStartInfo) {
			traceLog.Print("dns start")
		},
		DNSDone: func(_ httptrace.DNSDoneInfo) {
			traceLog.Print("dns done")
		},
		ConnectStart: func(_, _ string) {
			traceLog.Print("connection start")
		},
		ConnectDone: func(_, _ string, _ error) {
			traceLog.Print("connection done")
		},
		TLSHandshakeStart: func() {
			traceLog.Print("TLS Handshake start")
		},
		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
			traceLog.Print("TLS Handshake done")
		},
		WroteHeaders: func() {
			traceLog.Print("wrote headers")
		},
		WroteRequest: func(_ httptrace.WroteRequestInfo) {
			traceLog.Print("wrote request")
		},
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	// Don't leak the token in the logs. Temporarily replace the Header
	// with a clone in order to redact the token.
	origHeader := req.Header
	defer func() { req.Header = origHeader }()
	req.Header = maps.Clone(origHeader)
	req.Header.Set("Authorization", "Token <redacted>")

	dump, err := httputil.DumpRequest(req, true)
	if err != nil {
		traceLog.Printf("Couldn't dump request: %v", err)
		return req
	}
	traceLog.Printf("request uri=%s\n%s", req.URL, dump)
	return req
}
