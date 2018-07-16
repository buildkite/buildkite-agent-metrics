package backend

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/buildkite/buildkite-metrics/collector"
)

// CloudWatchBackend sends metrics to AWS CloudWatch
type CloudWatchBackend struct {
	dimension string
}

// NewCloudWatchBackend returns a new CloudWatchBackend with an
// optional Dimension in the form of Key=Value
func NewCloudWatchBackend(dimension string) *CloudWatchBackend {
	return &CloudWatchBackend{dimension: dimension}
}

func (cb *CloudWatchBackend) Collect(r *collector.Result) error {
	svc := cloudwatch.New(session.New())

	var dimensionKey, dimensionValue string

	// Support a custom dimension, needs to be parsed from Key=Value
	if cb.dimension != "" {
		dimensionParts := strings.SplitN(cb.dimension, "=", 2)
		if len(dimensionParts) != 2 {
			return fmt.Errorf("Failed to parse dimension of %q", cb.dimension)
		}
		dimensionKey = dimensionParts[0]
		dimensionValue = dimensionParts[1]

		log.Printf("Using custom Cloudwatch dimension of [ %s = %s ]", dimensionKey, dimensionValue)
	}

	metrics := []*cloudwatch.MetricDatum{}
	metrics = append(metrics, cloudwatchMetrics(r.Totals, nil)...)

	for name, c := range r.Queues {
		dimensions := []*cloudwatch.Dimension{}

		// Add custom dimension if provided
		if dimensionKey != "" {
			dimensions = append(dimensions, &cloudwatch.Dimension{
				Name: aws.String(dimensionKey), Value: aws.String(dimensionValue),
			})
		}

		dimensions = append(dimensions, &cloudwatch.Dimension{
			Name: aws.String("Queue"), Value: aws.String(name),
		})

		metrics = append(metrics, cloudwatchMetrics(c, dimensions)...)
	}

	log.Printf("Extracted %d cloudwatch metrics from results", len(metrics))

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

func cloudwatchMetrics(counts map[string]int, dimensions []*cloudwatch.Dimension) []*cloudwatch.MetricDatum {
	m := []*cloudwatch.MetricDatum{}

	for k, v := range counts {
		m = append(m, &cloudwatch.MetricDatum{
			MetricName: aws.String(k),
			Dimensions: dimensions,
			Value:      aws.Float64(float64(v)),
			Unit:       aws.String("Count"),
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
