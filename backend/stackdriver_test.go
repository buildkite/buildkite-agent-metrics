package backend

import (
	"testing"
	"time"

	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/genproto/googleapis/api/label"
	"google.golang.org/genproto/googleapis/api/metric"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_createCustomMetricRequest(t *testing.T) {
	type args struct {
		projectID  string
		metricType string
	}
	tests := []struct {
		name string
		args args
		want *monitoringpb.CreateMetricDescriptorRequest
	}{
		{
			name: "HappyPath",
			args: args{
				projectID:  "test-project-id",
				metricType: "test-metric-type",
			},
			want: &monitoringpb.CreateMetricDescriptorRequest{
				Name: "projects/test-project-id",
				MetricDescriptor: &metric.MetricDescriptor{
					Name:        "test-metric-type",
					Type:        "test-metric-type",
					MetricKind:  metric.MetricDescriptor_GAUGE,
					ValueType:   metric.MetricDescriptor_INT64,
					Description: "Buildkite metric: [test-metric-type]",
					DisplayName: "test-metric-type",
					Labels: []*label.LabelDescriptor{
						{
							Key:         clusterLabelKey,
							ValueType:   label.LabelDescriptor_STRING,
							Description: clusterDescription,
						},
						{
							Key:         queueLabelKey,
							ValueType:   label.LabelDescriptor_STRING,
							Description: queueDescription,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createCustomMetricRequest(&tt.args.projectID, &tt.args.metricType)
			if diff := cmp.Diff(got, tt.want, protocmp.Transform()); diff != "" {
				t.Errorf("createCustomMetricRequest diff (-got +want):\n%s", diff)
			}
		})
	}
}

func Test_createTimeSeriesValueRequest(t *testing.T) {
	now := &timestamppb.Timestamp{
		Seconds: time.Now().Unix(),
	}

	type args struct {
		projectID  string
		metricType string
		cluster    string
		queue      string
		value      int
		time       *timestamppb.Timestamp
	}
	tests := []struct {
		name string
		args args
		want *monitoringpb.CreateTimeSeriesRequest
	}{
		{
			name: "Happy Path",
			args: args{
				projectID:  "test-project-id",
				metricType: "test-metric-type",
				cluster:    "test-cluster",
				queue:      "test-queue",
				value:      13,
				time:       now,
			},
			want: &monitoringpb.CreateTimeSeriesRequest{
				Name: "projects/test-project-id",
				TimeSeries: []*monitoringpb.TimeSeries{{
					Metric: &metric.Metric{
						Type: "test-metric-type",
						Labels: map[string]string{
							clusterLabelKey: "test-cluster",
							queueLabelKey:   "test-queue",
						},
					},
					Points: []*monitoringpb.Point{{
						Interval: &monitoringpb.TimeInterval{
							StartTime: now,
							EndTime:   now,
						},
						Value: &monitoringpb.TypedValue{
							Value: &monitoringpb.TypedValue_Int64Value{
								Int64Value: int64(13),
							},
						},
					}},
				}},
			},
		},
		{
			name: "Bad Queue Name",
			args: args{
				projectID:  "test-project-id",
				metricType: "test-metric-type",
				queue:      "${BUILDKITE_QUEUE:-default}",
				value:      13,
				time:       now,
			},
			want: &monitoringpb.CreateTimeSeriesRequest{
				Name: "projects/test-project-id",
				TimeSeries: []*monitoringpb.TimeSeries{{
					Metric: &metric.Metric{
						Type: "test-metric-type",
						Labels: map[string]string{
							clusterLabelKey: "",
							queueLabelKey:   "${BUILDKITE_QUEUE:-default}",
						},
					},
					Points: []*monitoringpb.Point{{
						Interval: &monitoringpb.TimeInterval{
							StartTime: now,
							EndTime:   now,
						},
						Value: &monitoringpb.TypedValue{
							Value: &monitoringpb.TypedValue_Int64Value{
								Int64Value: int64(13),
							},
						},
					}},
				}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createTimeSeriesValueRequest(&tt.args.projectID, &tt.args.metricType, tt.args.cluster, tt.args.queue, tt.args.value, tt.args.time)
			if diff := cmp.Diff(got, tt.want, protocmp.Transform()); diff != "" {
				t.Errorf("createTimeSeriesValueRequest diff (-got +want):\n%s", diff)
			}
		})
	}
}
