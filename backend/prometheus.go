package backend

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/buildkite/buildkite-metrics/collector"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	camel = regexp.MustCompile("(^[^A-Z0-9]*|[A-Z0-9]*)([A-Z0-9][^A-Z]+|$)")
)

type Prometheus struct {
	totals    map[string]prometheus.Gauge
	queues    map[string]*prometheus.GaugeVec
	pipelines map[string]*prometheus.GaugeVec
}

func NewPrometheusBackend(path, addr string) *Prometheus {
	go func() {
		http.Handle(path, promhttp.Handler())
		log.Fatal(http.ListenAndServe(addr, nil))
	}()

	return newPrometheus()
}

func newPrometheus() *Prometheus {
	return &Prometheus{
		totals:    make(map[string]prometheus.Gauge),
		queues:    make(map[string]*prometheus.GaugeVec),
		pipelines: make(map[string]*prometheus.GaugeVec),
	}
}

func (p *Prometheus) Collect(r *collector.Result) error {
	for name, value := range r.Totals {
		gauge, ok := p.totals[name]
		if !ok {
			gauge = prometheus.NewGauge(prometheus.GaugeOpts{
				Name: fmt.Sprintf("buildkite_total_%s", camelToUnderscore(name)),
				Help: fmt.Sprintf("Buildkite Total: %s", name),
			})
			prometheus.MustRegister(gauge)
			p.totals[name] = gauge
		}
		gauge.Set(float64(value))
	}

	for queue, counts := range r.Queues {
		for name, value := range counts {
			gauge, ok := p.queues[name]
			if !ok {
				gauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
					Name: fmt.Sprintf("buildkite_queues_%s", camelToUnderscore(name)),
					Help: fmt.Sprintf("Buildkite Queues: %s", name),
				}, []string{"queue"})
				prometheus.MustRegister(gauge)
				p.queues[name] = gauge
			}
			gauge.WithLabelValues(queue).Set(float64(value))
		}
	}

	for pipeline, counts := range r.Pipelines {
		for name, value := range counts {
			gauge, ok := p.pipelines[name]
			if !ok {
				gauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
					Name: fmt.Sprintf("buildkite_pipelines_%s", camelToUnderscore(name)),
					Help: fmt.Sprintf("Buildkite Pipelines: %s", name),
				}, []string{"pipeline"})
				prometheus.MustRegister(gauge)
				p.pipelines[name] = gauge
			}
			gauge.WithLabelValues(pipeline).Set(float64(value))
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
