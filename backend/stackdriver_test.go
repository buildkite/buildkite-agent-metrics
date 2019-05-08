package backend

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/genproto/googleapis/api/label"
	"google.golang.org/genproto/googleapis/api/metric"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
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
		{name: "HappyPath", args: args{projectID: "test-project-id", metricType: "test-metric-type"}, want: &monitoringpb.CreateMetricDescriptorRequest{
			Name: "projects/test-project-id",
			MetricDescriptor: &metric.MetricDescriptor{
				Name:        "test-metric-type",
				Type:        "test-metric-type",
				MetricKind:  metric.MetricDescriptor_GAUGE,
				ValueType:   metric.MetricDescriptor_INT64,
				Description: fmt.Sprintf("Buildkite metric: [test-metric-type]"),
				DisplayName: "test-metric-type",
				Labels: []*label.LabelDescriptor{{
					Key:         queueLabelKey,
					ValueType:   label.LabelDescriptor_STRING,
					Description: queueDescription,
				}},
			},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createCustomMetricRequest(&tt.args.projectID, &tt.args.metricType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createCustomMetricRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createTimeSeriesValueRequest(t *testing.T) {
	now := timestamp.Timestamp{
		Seconds: time.Now().Unix(),
	}

	type args struct {
		projectID  string
		metricType string
		queue      string
		value      int
		time       timestamp.Timestamp
	}
	tests := []struct {
		name string
		args args
		want *monitoringpb.CreateTimeSeriesRequest
	}{
		{name: "Happy Path",
			args: args{projectID: "test-project-id", metricType: "test-metric-type", queue: "test-queue", value: 13, time: now},
			want: &monitoringpb.CreateTimeSeriesRequest{
				Name: "projects/test-project-id",
				TimeSeries: []*monitoringpb.TimeSeries{{
					Metric: &metric.Metric{
						Type:   "test-metric-type",
						Labels: map[string]string{queueLabelKey: "test-queue"},
					},
					Points: []*monitoringpb.Point{{
						Interval: &monitoringpb.TimeInterval{
							StartTime: &now,
							EndTime:   &now,
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
		{name: "Bad Queue Name",
			args: args{projectID: "test-project-id", metricType: "test-metric-type", queue: "${BUILDKITE_QUEUE:-default}", value: 13, time: now},
			want: &monitoringpb.CreateTimeSeriesRequest{
				Name: "projects/test-project-id",
				TimeSeries: []*monitoringpb.TimeSeries{{
					Metric: &metric.Metric{
						Type:   "test-metric-type",
						Labels: map[string]string{queueLabelKey: "${BUILDKITE_QUEUE:-default}"},
					},
					Points: []*monitoringpb.Point{{
						Interval: &monitoringpb.TimeInterval{
							StartTime: &now,
							EndTime:   &now,
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
			if got := createTimeSeriesValueRequest(&tt.args.projectID, &tt.args.metricType, tt.args.queue, tt.args.value, &tt.args.time); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createTimeSeriesValueRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
