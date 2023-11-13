package backend

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/buildkite/buildkite-agent-metrics/collector"
	"google.golang.org/genproto/googleapis/api/label"
	"google.golang.org/protobuf/types/known/timestamppb"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"google.golang.org/genproto/googleapis/api/metric"
)

const (
	metricTypeFmt      = "custom.googleapis.com/buildkite/%s/%s"
	clusterLabelKey    = "Cluster"
	clusterDescription = "Name of the Buildkite Cluster, or empty"
	queueLabelKey      = "Queue"
	queueDescription   = "Name of the Queue"
	totalMetricsQueue  = "Total"
)

// Stackdriver does not allow dashes in metric names
var dashReplacer = strings.NewReplacer("-", "_")

// StackDriverBackend sends metrics to GCP Stackdriver
type StackDriverBackend struct {
	projectID   string
	client      *monitoring.MetricClient
	metricTypes map[string]string
}

// NewStackDriverBackend returns a new StackDriverBackend for the specified project
func NewStackDriverBackend(gcpProjectID string) (*StackDriverBackend, error) {
	ctx := context.Background()
	c, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("[NewStackDriverBackend] could not create stackdriver client: %v", err)
	}

	return &StackDriverBackend{
		projectID:   gcpProjectID,
		client:      c,
		metricTypes: make(map[string]string),
	}, nil
}

// Collect metrics
func (sd *StackDriverBackend) Collect(r *collector.Result) error {
	ctx := context.Background()
	now := &timestamppb.Timestamp{
		Seconds: time.Now().Unix(),
	}
	orgName := dashReplacer.Replace(r.Org)
	metricTypeFunc := func(name string) string {
		return fmt.Sprintf(metricTypeFmt, orgName, name)
	}

	for name, value := range r.Totals {
		mt, present := sd.metricTypes[name]
		if !present {
			mt = metricTypeFunc(name)
			metricReq := createCustomMetricRequest(&sd.projectID, &mt)
			_, err := sd.client.CreateMetricDescriptor(ctx, metricReq)
			if err != nil {
				retErr := fmt.Errorf("[Collect] could not create custom metric [%s]: %v", mt, err)
				log.Println(retErr)
				return retErr
			}
			log.Printf("[Collect] created custom metric [%s]", mt)
			sd.metricTypes[name] = mt
		}
		req := createTimeSeriesValueRequest(&sd.projectID, &mt, r.Cluster, totalMetricsQueue, value, now)
		err := sd.client.CreateTimeSeries(ctx, req)
		if err != nil {
			retErr := fmt.Errorf("[Collect] could not write metric [%s] value [%d], %v ", mt, value, err)
			log.Println(retErr)
			return retErr
		}
	}

	for queue, counts := range r.Queues {
		for name, value := range counts {
			mt := metricTypeFunc(name)
			req := createTimeSeriesValueRequest(&sd.projectID, &mt, r.Cluster, queue, value, now)
			err := sd.client.CreateTimeSeries(ctx, req)
			if err != nil {
				retErr := fmt.Errorf("[Collect] could not write metric [%s] value [%d], %v ", mt, value, err)
				log.Println(retErr)
				return retErr
			}
		}
	}

	return nil
}

// createCustomMetricRequest creates a custom metric request as specified by the metric type.
func createCustomMetricRequest(projectID *string, metricType *string) *monitoringpb.CreateMetricDescriptorRequest {
	clusterLabel := &label.LabelDescriptor{
		Key:         clusterLabelKey,
		ValueType:   label.LabelDescriptor_STRING,
		Description: clusterDescription,
	}
	queueLabel := &label.LabelDescriptor{
		Key:         queueLabelKey,
		ValueType:   label.LabelDescriptor_STRING,
		Description: queueDescription,
	}
	labels := []*label.LabelDescriptor{
		clusterLabel,
		queueLabel,
	}
	md := &metric.MetricDescriptor{
		Name:        *metricType,
		Type:        *metricType,
		MetricKind:  metric.MetricDescriptor_GAUGE,
		ValueType:   metric.MetricDescriptor_INT64,
		Description: fmt.Sprintf("Buildkite metric: [%s]", *metricType),
		DisplayName: *metricType,
		Labels:      labels,
	}
	req := &monitoringpb.CreateMetricDescriptorRequest{
		Name:             "projects/" + *projectID,
		MetricDescriptor: md,
	}

	return req
}

// createTimeSeriesValueRequest creates a StackDriver value request for the specified metric
func createTimeSeriesValueRequest(projectID *string, metricType *string, cluster, queue string, value int, time *timestamppb.Timestamp) *monitoringpb.CreateTimeSeriesRequest {
	req := &monitoringpb.CreateTimeSeriesRequest{
		Name: "projects/" + *projectID,
		TimeSeries: []*monitoringpb.TimeSeries{{
			Metric: &metric.Metric{
				Type: *metricType,
				Labels: map[string]string{
					clusterLabelKey: cluster,
					queueLabelKey:   queue,
				},
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
