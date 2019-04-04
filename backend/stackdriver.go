package backend

import (
	"context"
	"fmt"
	"github.com/buildkite/buildkite-agent-metrics/collector"
	"github.com/golang/protobuf/ptypes/timestamp"
	"time"

	"cloud.google.com/go/monitoring/apiv3"
	"google.golang.org/genproto/googleapis/api/metric"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

const (
	metricTotalPrefix = "custom.googleapis.com/buildlkite/total/%s"
	metricQueuePrefix = "custom.googleapis.com/buildlkite/queue/%s/%s"
)

// StackDriverBackend sends metrics to GCP Stackdriver
type StackDriverBackend struct {
	projectId		string
	client 			*monitoring.MetricClient
	queues 			map[string]string
}

// NewStackdriverBackend returns a new StackDriverBackend for the specified project
func NewStackdriverBackend(projectID string) (*StackDriverBackend, error){
	ctx := context.Background()
	c, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not create stackdriver client: %v", err)
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
		if _, err = createCustomMetric(c, &projectID, &mt); err != nil {
			return nil, err
		}
	}

	return &StackDriverBackend{
		projectId: 		projectID,
		client: 		c,
		queues: 		make(map[string]string),
	}, nil
}

func (sd *StackDriverBackend) Collect(r *collector.Result) error {
	now := &timestamp.Timestamp{
		Seconds: time.Now().Unix(),
	}
	for name, value := range r.Totals {
		mt := fmt.Sprintf(metricTotalPrefix, name)
		if err := writeTimeSeriesValue(sd.client, &sd.projectId, &mt, value, now); err != nil {
			return err
		}
	}

	for queue, counts := range r.Queues {
		for name, value := range counts {
			mt := fmt.Sprintf(metricQueuePrefix, queue, name)
			if _, ok := sd.queues[mt]; !ok {
				if _, err := createCustomMetric(sd.client, &sd.projectId, &mt); err != nil {
					return err
				}
				sd.queues[mt] = queue
			}
			if err := writeTimeSeriesValue(sd.client, &sd.projectId, &mt, value, now); err != nil {
				return err
			}
		}
	}

	return nil
}

// createCustomMetric creates a custom metric specified by the metric type.
func createCustomMetric(c *monitoring.MetricClient, projectID *string, metricType *string) (*metricpb.MetricDescriptor, error) {
	ctx := context.Background()
	md := &metric.MetricDescriptor{
		Name: *metricType,
		Type: *metricType,
		MetricKind:  metric.MetricDescriptor_GAUGE,
		ValueType:   metric.MetricDescriptor_INT64,
		Description: fmt.Sprintf("Buildkite metric: [%s]" + *metricType),
		DisplayName: *metricType,
	}
	req := &monitoringpb.CreateMetricDescriptorRequest{
		Name:             "projects/" + *projectID,
		MetricDescriptor: md,
	}
	m, err := c.CreateMetricDescriptor(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("could not create custom metric [%s]: %v", *metricType, err)
	}

	return m, nil
}

// writeTimeSeriesValue writes a value to Stackdriver for the specified metric
func writeTimeSeriesValue(c *monitoring.MetricClient, projectID *string, metricType *string, value int, time *timestamp.Timestamp) error {
	ctx := context.Background()
	req := &monitoringpb.CreateTimeSeriesRequest{
		Name: "projects/" + *projectID,
		TimeSeries: []*monitoringpb.TimeSeries{{
			Metric: &metricpb.Metric{
				Type: *metricType,
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

	err := c.CreateTimeSeries(ctx, req)
	if err != nil {
		return fmt.Errorf("could not write metric [%s] value [%d], %v ", *metricType, value, err)
	}
	return nil
}
