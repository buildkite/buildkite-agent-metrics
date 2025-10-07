package backend

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/buildkite/buildkite-agent-metrics/v5/collector"
)

// CloudWatchDimension is a dimension to add to metrics
type CloudWatchDimension struct {
	Key   string
	Value string
}

func ParseCloudWatchDimensions(ds string) ([]CloudWatchDimension, error) {
	dimensions := []CloudWatchDimension{}

	if strings.TrimSpace(ds) == "" {
		return dimensions, nil
	}

	for _, dimension := range strings.Split(strings.TrimSpace(ds), ",") {
		parts := strings.SplitN(strings.TrimSpace(dimension), "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("Failed to parse dimension of %q", dimension)
		}
		dimensions = append(dimensions, CloudWatchDimension{
			Key:   parts[0],
			Value: parts[1],
		})
	}

	return dimensions, nil
}

// CloudWatchBackend sends metrics to AWS CloudWatch
type CloudWatchBackend struct {
	region               string
	dimensions           []CloudWatchDimension
	interval             int64
	enableHighResolution bool
}

// NewCloudWatchBackend returns a new CloudWatchBackend with optional dimensions
func NewCloudWatchBackend(region string, dimensions []CloudWatchDimension, interval int64, enableHighResolution bool) *CloudWatchBackend {
	return &CloudWatchBackend{
		region:               region,
		dimensions:           dimensions,
		interval:             interval,
		enableHighResolution: enableHighResolution,
	}
}

func (cb *CloudWatchBackend) Collect(r *collector.Result) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cb.region),
	})
	if err != nil {
		return err
	}

	svc := cloudwatch.New(sess)
	metrics := []*cloudwatch.MetricDatum{}

	// Set the baseline org dimension
	dimensions := []*cloudwatch.Dimension{
		{
			Name:  aws.String("Org"),
			Value: aws.String(r.Org),
		},
	}

	// Add cluster dimension if a cluster token was used
	if r.Cluster != "" {
		dimensions = append(dimensions, &cloudwatch.Dimension{
			Name:  aws.String("Cluster"),
			Value: aws.String(r.Cluster),
		})
	}

	// Add custom dimension if provided
	for _, d := range cb.dimensions {
		log.Printf("Using custom Cloudwatch dimension of [ %s = %s ]", d.Key, d.Value)

		dimensions = append(dimensions, &cloudwatch.Dimension{
			Name: aws.String(d.Key), Value: aws.String(d.Value),
		})
	}

	// Add total metrics
	metrics = append(metrics, cb.cloudwatchMetrics(r.Totals, nil)...)

	for name, c := range r.Queues {
		queueDimensions := dimensions

		// Add an queue dimension
		queueDimensions = append(queueDimensions,
			&cloudwatch.Dimension{Name: aws.String("Queue"), Value: aws.String(name)},
		)

		// Add per-queue metrics
		metrics = append(metrics, cb.cloudwatchMetrics(c, queueDimensions)...)
	}

	log.Printf("Extracted %d cloudwatch metrics from results", len(metrics))

	// Chunk into batches of 10 metrics
	for _, chunk := range chunkCloudwatchMetrics(10, metrics) {
		log.Printf("Submitting chunk of %d metrics to Cloudwatch", len(chunk))
		_, err := svc.PutMetricData(&cloudwatch.PutMetricDataInput{
			MetricData: chunk,
			Namespace:  aws.String("Buildkite"),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (cb *CloudWatchBackend) cloudwatchMetrics(counts map[string]int, dimensions []*cloudwatch.Dimension) []*cloudwatch.MetricDatum {
	m := []*cloudwatch.MetricDatum{}

	var duration int64
	if cb.interval < 60 && cb.enableHighResolution {
		// PutMetricData supports either normal (60s) or high frequency (1s)
		// metrics - other values result in an error.
		duration = 1
	} else {
		duration = 60
	}

	for k, v := range counts {
		m = append(m, &cloudwatch.MetricDatum{
			MetricName:        aws.String(k),
			Dimensions:        dimensions,
			Value:             aws.Float64(float64(v)),
			Unit:              aws.String("Count"),
			StorageResolution: &duration,
		})
	}

	return m
}

func chunkCloudwatchMetrics(size int, data []*cloudwatch.MetricDatum) [][]*cloudwatch.MetricDatum {
	var chunks = [][]*cloudwatch.MetricDatum{}
	for i := 0; i < len(data); i += size {
		end := i + size
		if end > len(data) {
			end = len(data)
		}
		chunks = append(chunks, data[i:end])
	}
	return chunks
}
