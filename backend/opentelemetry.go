package backend

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/buildkite/buildkite-agent-metrics/v5/collector"
	"github.com/buildkite/buildkite-agent-metrics/v5/version"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

type OpenTelemetryBackend struct {
	tracer trace.Tracer
	meter  metric.Meter

	// Metrics instruments
	scheduledJobsCounter  metric.Int64Counter
	runningJobsCounter    metric.Int64Counter
	unfinishedJobsCounter metric.Int64Counter
	waitingJobsCounter    metric.Int64Counter
	idleAgentsGauge       metric.Int64Gauge
	busyAgentsGauge       metric.Int64Gauge
	totalAgentsGauge      metric.Int64Gauge
	busyAgentPercentGauge metric.Int64Gauge
	collectionDuration    metric.Float64Histogram

	shutdown func()
}

// NewOpenTelemetryBackend creates a new OpenTelemetry backend.
//
// # Configuration is through standard OpenTelemetry SDK environment variables
//
// https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/
// https://opentelemetry.io/docs/specs/otel/protocol/exporter/#endpoint-urls-for-otlphttp
func NewOpenTelemetryBackend() (*OpenTelemetryBackend, error) {
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "buildkite-agent-metrics"
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(serviceName),
		semconv.ServiceNamespace("buildkite"),
		semconv.ServiceVersion(version.Version),
	)

	// Set up trace exporter based on protocol
	var traceExporter sdktrace.SpanExporter
	var err error
	protocol := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")
	// default to http/protobuf as per spec
	if protocol == "" {
		protocol = "http/protobuf"
	}

	switch protocol {
	case "grpc":
		traceExporter, err = otlptracegrpc.New(context.Background())
	case "http/protobuf", "http":
		traceExporter, err = otlptracehttp.New(context.Background())
	default:
		return nil, fmt.Errorf("unsupported otlp protocol: %s", protocol)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Set up trace provider
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tracerProvider)

	// Set up metric exporter based on protocol
	var metricExporter sdkmetric.Exporter

	switch protocol {
	case "grpc":
		metricExporter, err = otlpmetricgrpc.New(context.Background())
	case "http/protobuf", "http":
		metricExporter, err = otlpmetrichttp.New(context.Background())
	}
	if err != nil {
		tracerProvider.Shutdown(context.Background())
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	// Set up metric provider
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)

	// Set up text map propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Create shutdown function
	otelShutdown := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		tracerProvider.Shutdown(ctx)
		meterProvider.Shutdown(ctx)
	}

	backend := &OpenTelemetryBackend{
		tracer:   otel.Tracer("buildkite-agent-metrics"),
		meter:    otel.Meter("buildkite-agent-metrics"),
		shutdown: otelShutdown,
	}

	// Initialize metrics instruments
	if err := backend.initializeMetrics(); err != nil {
		otelShutdown()
		return nil, fmt.Errorf("error initializing metrics: %w", err)
	}

	log.Println("OpenTelemetry backend initialized successfully")
	return backend, nil
}

func (b *OpenTelemetryBackend) initializeMetrics() error {
	var err error

	b.scheduledJobsCounter, err = b.meter.Int64Counter(
		"buildkite.jobs.scheduled",
		metric.WithDescription("Number of scheduled jobs"),
	)
	if err != nil {
		return err
	}

	b.runningJobsCounter, err = b.meter.Int64Counter(
		"buildkite.jobs.running",
		metric.WithDescription("Number of running jobs"),
	)
	if err != nil {
		return err
	}

	b.unfinishedJobsCounter, err = b.meter.Int64Counter(
		"buildkite.jobs.unfinished",
		metric.WithDescription("Number of unfinished jobs"),
	)
	if err != nil {
		return err
	}

	b.waitingJobsCounter, err = b.meter.Int64Counter(
		"buildkite.jobs.waiting",
		metric.WithDescription("Number of waiting jobs"),
	)
	if err != nil {
		return err
	}

	b.idleAgentsGauge, err = b.meter.Int64Gauge(
		"buildkite.agents.idle",
		metric.WithDescription("Number of idle agents"),
	)
	if err != nil {
		return err
	}

	b.busyAgentsGauge, err = b.meter.Int64Gauge(
		"buildkite.agents.busy",
		metric.WithDescription("Number of busy agents"),
	)
	if err != nil {
		return err
	}

	b.totalAgentsGauge, err = b.meter.Int64Gauge(
		"buildkite.agents.total",
		metric.WithDescription("Total number of agents"),
	)
	if err != nil {
		return err
	}

	b.busyAgentPercentGauge, err = b.meter.Int64Gauge(
		"buildkite.agents.busy_percentage",
		metric.WithDescription("Percentage of busy agents"),
	)
	if err != nil {
		return err
	}

	b.collectionDuration, err = b.meter.Float64Histogram(
		"buildkite.collection.duration",
		metric.WithDescription("Duration of metrics collection"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	return nil
}

// Collect implements the Backend interface
func (b *OpenTelemetryBackend) Collect(r *collector.Result) error {
	ctx := context.Background()
	start := time.Now()

	// Start tracing span for this collection
	ctx, span := b.tracer.Start(ctx, "collect_metrics")
	defer span.End()

	// Add span attributes
	span.SetAttributes(
		attribute.String("org", r.Org),
		attribute.String("cluster", r.Cluster),
		attribute.Int("queues_count", len(r.Queues)),
	)

	commonAttrs := []attribute.KeyValue{
		attribute.String("org", r.Org),
		attribute.String("cluster", r.Cluster),
	}

	// Record total metrics
	if val, ok := r.Totals["ScheduledJobsCount"]; ok {
		b.scheduledJobsCounter.Add(ctx, int64(val), metric.WithAttributes(commonAttrs...))
	}
	if val, ok := r.Totals["RunningJobsCount"]; ok {
		b.runningJobsCounter.Add(ctx, int64(val), metric.WithAttributes(commonAttrs...))
	}
	if val, ok := r.Totals["UnfinishedJobsCount"]; ok {
		b.unfinishedJobsCounter.Add(ctx, int64(val), metric.WithAttributes(commonAttrs...))
	}
	if val, ok := r.Totals["WaitingJobsCount"]; ok {
		b.waitingJobsCounter.Add(ctx, int64(val), metric.WithAttributes(commonAttrs...))
	}
	if val, ok := r.Totals["IdleAgentCount"]; ok {
		b.idleAgentsGauge.Record(ctx, int64(val), metric.WithAttributes(commonAttrs...))
	}
	if val, ok := r.Totals["BusyAgentCount"]; ok {
		b.busyAgentsGauge.Record(ctx, int64(val), metric.WithAttributes(commonAttrs...))
	}
	if val, ok := r.Totals["TotalAgentCount"]; ok {
		b.totalAgentsGauge.Record(ctx, int64(val), metric.WithAttributes(commonAttrs...))
	}
	if val, ok := r.Totals["BusyAgentPercentage"]; ok {
		b.busyAgentPercentGauge.Record(ctx, int64(val), metric.WithAttributes(commonAttrs...))
	}

	// Record per-queue metrics and collect data for events
	queueEvents := make([]map[string]any, 0, len(r.Queues))

	for queueName, queueMetrics := range r.Queues {
		queueAttrs := append(commonAttrs, attribute.String("queue", queueName))

		// Extract values for logging
		scheduledJobs := int64(0)
		runningJobs := int64(0)
		unfinishedJobs := int64(0)
		waitingJobs := int64(0)
		idleAgents := int64(0)
		busyAgents := int64(0)
		totalAgents := int64(0)
		busyPercentage := int64(0)

		if val, ok := queueMetrics["ScheduledJobsCount"]; ok {
			scheduledJobs = int64(val)
			b.scheduledJobsCounter.Add(ctx, scheduledJobs, metric.WithAttributes(queueAttrs...))
		}
		if val, ok := queueMetrics["RunningJobsCount"]; ok {
			runningJobs = int64(val)
			b.runningJobsCounter.Add(ctx, runningJobs, metric.WithAttributes(queueAttrs...))
		}
		if val, ok := queueMetrics["UnfinishedJobsCount"]; ok {
			unfinishedJobs = int64(val)
			b.unfinishedJobsCounter.Add(ctx, unfinishedJobs, metric.WithAttributes(queueAttrs...))
		}
		if val, ok := queueMetrics["WaitingJobsCount"]; ok {
			waitingJobs = int64(val)
			b.waitingJobsCounter.Add(ctx, waitingJobs, metric.WithAttributes(queueAttrs...))
		}
		if val, ok := queueMetrics["IdleAgentCount"]; ok {
			idleAgents = int64(val)
			b.idleAgentsGauge.Record(ctx, idleAgents, metric.WithAttributes(queueAttrs...))
		}
		if val, ok := queueMetrics["BusyAgentCount"]; ok {
			busyAgents = int64(val)
			b.busyAgentsGauge.Record(ctx, busyAgents, metric.WithAttributes(queueAttrs...))
		}
		if val, ok := queueMetrics["TotalAgentCount"]; ok {
			totalAgents = int64(val)
			b.totalAgentsGauge.Record(ctx, totalAgents, metric.WithAttributes(queueAttrs...))
		}
		if val, ok := queueMetrics["BusyAgentPercentage"]; ok {
			busyPercentage = int64(val)
			b.busyAgentPercentGauge.Record(ctx, busyPercentage, metric.WithAttributes(queueAttrs...))
		}

		// Store queue data for event logging
		queueEvents = append(queueEvents, map[string]any{
			"name":       queueName,
			"scheduled":  scheduledJobs,
			"running":    runningJobs,
			"unfinished": unfinishedJobs,
			"waiting":    waitingJobs,
			"idle":       idleAgents,
			"busy":       busyAgents,
			"total":      totalAgents,
			"busy_pct":   busyPercentage,
		})

		// Print to logs for local debugging
		log.Printf("Queue '%s': Scheduled=%d, Running=%d, Unfinished=%d, Waiting=%d, Idle=%d, Busy=%d, Total=%d, Busy%%=%d",
			queueName, scheduledJobs, runningJobs, unfinishedJobs, waitingJobs,
			idleAgents, busyAgents, totalAgents, busyPercentage)
	}

	// Record collection duration
	collectionDuration := time.Since(start)
	b.collectionDuration.Record(ctx, collectionDuration.Seconds(),
		metric.WithAttributes(commonAttrs...),
	)

	// Add individual queue events to span
	for _, queueData := range queueEvents {
		span.AddEvent("queue_metrics", trace.WithAttributes(
			attribute.String("queue_name", queueData["name"].(string)),
			attribute.Int64("scheduled_jobs", queueData["scheduled"].(int64)),
			attribute.Int64("running_jobs", queueData["running"].(int64)),
			attribute.Int64("unfinished_jobs", queueData["unfinished"].(int64)),
			attribute.Int64("waiting_jobs", queueData["waiting"].(int64)),
			attribute.Int64("idle_agents", queueData["idle"].(int64)),
			attribute.Int64("busy_agents", queueData["busy"].(int64)),
			attribute.Int64("total_agents", queueData["total"].(int64)),
			attribute.Int64("busy_agent_percentage", queueData["busy_pct"].(int64)),
		))
	}

	// Add collection info to span with queue details
	span.AddEvent("metrics_collected", trace.WithAttributes(
		attribute.Int("total_metrics", len(r.Totals)),
		attribute.Int("queue_count", len(r.Queues)),
		attribute.Float64("duration_seconds", collectionDuration.Seconds()),
	))

	// Add a summary event with all queue names for debugging
	queueNames := make([]string, 0, len(r.Queues))
	for queueName := range r.Queues {
		queueNames = append(queueNames, queueName)
	}
	span.AddEvent("queue_summary", trace.WithAttributes(
		attribute.StringSlice("queue_names", queueNames),
		attribute.Int("events_added", len(queueEvents)),
	))

	return nil
}

// Close implements the Closer interface
func (b *OpenTelemetryBackend) Close() error {
	if b.shutdown != nil {
		b.shutdown()
	}
	return nil
}
