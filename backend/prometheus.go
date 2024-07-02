package backend

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/buildkite/buildkite-agent-metrics/v5/collector"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	camel = regexp.MustCompile("(^[^A-Z0-9]*|[A-Z0-9]*)([A-Z0-9][^A-Z]+|$)")
)

type Prometheus struct {
	totals    map[string]*prometheus.GaugeVec
	queues    map[string]*prometheus.GaugeVec
	pipelines map[string]*prometheus.GaugeVec
}

func NewPrometheusBackend() *Prometheus {
	return &Prometheus{
		totals:    make(map[string]*prometheus.GaugeVec),
		queues:    make(map[string]*prometheus.GaugeVec),
		pipelines: make(map[string]*prometheus.GaugeVec),
	}
}

// Serve runs a Prometheus metrics HTTP server.
func (p *Prometheus) Serve(path, addr string) {
	m := http.NewServeMux()
	m.Handle(path, promhttp.Handler())
	log.Fatal(http.ListenAndServe(addr, m))
}

func (p *Prometheus) Collect(r *collector.Result) error {
	// Clear the gauges to prevent stale values from persisting forever.
	for _, gauge := range p.queues {
		gauge.Reset()
	}

	for name, value := range r.Totals {
		labelNames := []string{"cluster"}
		gauge, ok := p.totals[name]
		if !ok {
			gauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: fmt.Sprintf("buildkite_total_%s", camelToUnderscore(name)),
				Help: fmt.Sprintf("Buildkite Total: %s", name),
			}, labelNames)
			prometheus.MustRegister(gauge)
			p.totals[name] = gauge
		}

		// note that r.Cluster will be empty for unclustered agents, this label will be dropped by promethues
		gauge.WithLabelValues(r.Cluster).Set(float64(value))
	}

	for queue, counts := range r.Queues {
		for name, value := range counts {
			gauge, ok := p.queues[name]
			if !ok {
				labelNames := []string{"queue", "cluster"}
				gauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
					Name: fmt.Sprintf("buildkite_queues_%s", camelToUnderscore(name)),
					Help: fmt.Sprintf("Buildkite Queues: %s", name),
				}, labelNames)
				prometheus.MustRegister(gauge)
				p.queues[name] = gauge
			}

			// note that r.Cluster will be empty for unclustered agents, this label will be dropped by promethues
			gauge.WithLabelValues(queue, r.Cluster).Set(float64(value))
		}
	}

	return nil
}

func camelToUnderscore(s string) string {
	var a []string
	for _, sub := range camel.FindAllStringSubmatch(s, -1) {
		if sub[1] != "" {
			a = append(a, sub[1])
		}
		if sub[2] != "" {
			a = append(a, sub[2])
		}
	}
	return strings.ToLower(strings.Join(a, "_"))
}
