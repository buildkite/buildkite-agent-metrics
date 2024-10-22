package backend

import (
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/buildkite/buildkite-agent-metrics/v5/collector"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var camelCaseRE = regexp.MustCompile("(^[^A-Z0-9]*|[A-Z0-9]*)([A-Z0-9][^A-Z]+|$)")

// Prometheus this holds a list of prometheus gauges which have been created,
// one for each metric that we want to expose. These are created and registered
// in NewPrometheusBackend.
//
// Note: these metrics are not unique to a cluster / queue, as these labels are
// added to the value when it is set.
type Prometheus struct {
	totals    map[string]*prometheus.GaugeVec
	queues    map[string]*prometheus.GaugeVec
	oldQueues map[string]map[string]struct{} // cluster -> set of queues in cluster from last collect
}

var (
	promSingletonOnce sync.Once
	promSingleton     *Prometheus
)

// NewPrometheusBackend creates an instance of Prometheus and creates and
// registers all the metrics gauges. Because Prometheus metrics must be unique,
// it manages a singleton instance rather than creating a new backend for each
// call.
func NewPrometheusBackend() *Prometheus {
	promSingletonOnce.Do(createPromSingleton)
	return promSingleton
}

func createPromSingleton() {
	promSingleton = &Prometheus{
		totals:    make(map[string]*prometheus.GaugeVec),
		queues:    make(map[string]*prometheus.GaugeVec),
		oldQueues: make(map[string]map[string]struct{}),
	}

	for _, name := range collector.AllMetrics {
		gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "buildkite_total_" + camelToUnderscore(name),
			Help: "Buildkite Total: " + name,
		}, []string{"cluster"})
		prometheus.MustRegister(gauge)
		promSingleton.totals[name] = gauge

		gauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "buildkite_queues_" + camelToUnderscore(name),
			Help: "Buildkite Queues: " + name,
		}, []string{"queue", "cluster"})
		prometheus.MustRegister(gauge)
		promSingleton.queues[name] = gauge
	}
}

// Serve runs a Prometheus metrics HTTP server.
func (p *Prometheus) Serve(path, addr string) {
	m := http.NewServeMux()
	m.Handle(path, promhttp.Handler())
	log.Fatal(http.ListenAndServe(addr, m))
}

// Collect receives a set of metrics from the agent and updates the gauges.
//
// Note: This is called once per agent token per interval
func (p *Prometheus) Collect(r *collector.Result) error {

	// Ranging over all gauges and searching Totals / Queues for values ensures
	// that metrics that are not in this collection are reset to 0.

	for name, gauge := range p.totals {
		value := r.Totals[name] // 0 if missing

		// note that r.Cluster will be empty for unclustered agents, this label
		// will be dropped by prometheus
		gauge.With(prometheus.Labels{
			"cluster": r.Cluster,
		}).Set(float64(value))
	}

	currentQueues := make(map[string]struct{})
	oldQueues := p.oldQueues[r.Cluster]
	for queue, counts := range r.Queues {
		currentQueues[queue] = struct{}{}
		delete(oldQueues, queue) // still current

		for name, gauge := range p.queues {
			value := counts[name] // 0 if missing

			// note that r.Cluster will be empty for unclustered agents, this
			// label will be dropped by prometheus
			gauge.With(prometheus.Labels{
				"cluster": r.Cluster,
				"queue":   queue,
			}).Set(float64(value))
		}
	}

	// oldQueues contains queues that were in the previous collector result, but
	// are no longer present.
	// This is to prevent accumulating label values for deleted queues.
	for queue := range oldQueues {
		for _, gauge := range p.queues {
			gauge.Delete(prometheus.Labels{
				"cluster": r.Cluster,
				"queue":   queue,
			})
		}
	}
	p.oldQueues[r.Cluster] = currentQueues

	return nil
}

func camelToUnderscore(s string) string {
	var a []string
	for _, sub := range camelCaseRE.FindAllStringSubmatch(s, -1) {
		if sub[1] != "" {
			a = append(a, sub[1])
		}
		if sub[2] != "" {
			a = append(a, sub[2])
		}
	}
	return strings.ToLower(strings.Join(a, "_"))
}
