package backend

import (
	"context"
	"fmt"
	"github.com/buildkite/buildkite-agent-metrics/collector"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/genproto/googleapis/api/label"
	"log"
	"time"

	"cloud.google.com/go/monitoring/apiv3"
	"google.golang.org/genproto/googleapis/api/metric"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

const (
	metricTotalPrefix = "custom.googleapis.com/buildkite/total/%s"
	queueLabelKey = "Queue"
	queueDescription = "Queue Descriptor"
	totalMetricsQueue = "Total"
)

// StackDriverBackend sends metrics to GCP Stackdriver
type StackDriverBackend struct {
	projectId		string
	client 			*monitoring.MetricClient
}

// NewStackdriverBackend returns a new StackDriverBackend for the specified project
func NewStackdriverBackend(gcpProjectID string) (*StackDriverBackend, error) {
	ctx := context.Background()
	c, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("[NewStackdriverBackend] could not create stackdriver client: %v", err)
	}
	countTypes := []string{
		collector.RunningJobsCount,
		collector.ScheduledJobsCount,
		collector.UnfinishedJobsCount,
		collector.TotalAgentCount,
		collector.BusyAgentCount,
		collector.IdleAgentCount,
	}
	for _, name := range countTypes {
		mt := fmt.Sprintf(metricTotalPrefix, name)
		req := createCustomMetricRequest(&gcpProjectID, &mt)
		_, err := c.CreateMetricDescriptor(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("[NewStackdriverBackend] could not create custom metric [%s]: %v", mt, err)
		}
		log.Printf("[NewStackdriverBackend] created custom metric [%s]", mt)
	}

	return &StackDriverBackend{
		projectId: 		gcpProjectID,
		client: 		c,
	}, nil
}

func (sd *StackDriverBackend) Collect(r *collector.Result) error {
	ctx := context.Background()
	now := &timestamp.Timestamp{
		Seconds: time.Now().Unix(),
	}
	for name, value := range r.Totals {
		mt := fmt.Sprintf(metricTotalPrefix, name)
		req := createTimeSeriesValueRequest(&sd.projectId, &mt, totalMetricsQueue, value, now)
		err := sd.client.CreateTimeSeries(ctx, req)
		if err != nil {
			return fmt.Errorf("[Collect] could not write metric [%s] value [%d], %v ", mt, value, err)
		}
	}

	for queue, counts := range r.Queues {
		for name, value := range counts {
			mt := fmt.Sprintf(metricTotalPrefix, name)
			req := createTimeSeriesValueRequest(&sd.projectId, &mt, queue, value, now)
			err := sd.client.CreateTimeSeries(ctx, req)
			if err != nil {
				return fmt.Errorf("[Collect] could not write metric [%s] value [%d], %v ", mt, value, err)
			}
		}
	}

	return nil
}

// createCustomMetricRequest creates a custom metric request as specified by the metric type.
func createCustomMetricRequest(projectID *string, metricType *string) (*monitoringpb.CreateMetricDescriptorRequest) {
	l := &label.LabelDescriptor{
		Key:				queueLabelKey,
		ValueType: 			label.LabelDescriptor_STRING,
		Description: 		queueDescription,

	}
	labels := []*label.LabelDescriptor{l}
	md := &metric.MetricDescriptor{
		Name: *metricType,
		Type: *metricType,
		MetricKind:  metric.MetricDescriptor_GAUGE,
		ValueType:   metric.MetricDescriptor_INT64,
		Description: fmt.Sprintf("Buildkite metric: [%s]", *metricType),
		DisplayName: *metricType,
		Labels: 	 labels,
	}
	req := &monitoringpb.CreateMetricDescriptorRequest{
		Name:             "projects/" + *projectID,
		MetricDescriptor: md,
	}

	return req
}

// createTimeSeriesValueRequest create a Stackdriver value request for the specified metric
func createTimeSeriesValueRequest(projectID *string, metricType *string, queue string, value int, time *timestamp.Timestamp) *monitoringpb.CreateTimeSeriesRequest {
	req := &monitoringpb.CreateTimeSeriesRequest{
		Name: "projects/" + *projectID,
		TimeSeries: []*monitoringpb.TimeSeries{{
			Metric: &metric.Metric{
				Type: *metricType,
				Labels: map[string]string{queueLabelKey: queue},
			},
			Points: []*monitoringpb.Point{{
				Interval: &monitoringpb.TimeInterval{
					StartTime: time,
					EndTime:   time,
				},
				Value: &monitoringpb.TypedValue{
					Value: &monitoringpb.TypedValue_Int64Value{
						Int64Value: int64(value),
					},
				},
			}},
		}},
	}
	return req
}
