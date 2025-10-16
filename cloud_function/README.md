# Buildkite Agent Metrics - GCP Cloud Function

This Google Cloud Function collects Buildkite CI/CD metrics and sends them to Google Cloud Monitoring (Stackdriver) for use in auto-scaling decisions.

## Overview

The function periodically fetches the following metrics from Buildkite API:
- **Job Metrics**: ScheduledJobsCount, RunningJobsCount, UnfinishedJobsCount, WaitingJobsCount
- **Agent Metrics**: IdleAgentCount, BusyAgentCount, TotalAgentCount, BusyAgentPercentage

These metrics appear in Cloud Monitoring under: `custom.googleapis.com/buildkite/{org}/{metric_name}`

## Prerequisites

1. **Google Cloud Project** with the following APIs enabled:
   - Cloud Functions API
   - Cloud Monitoring API
   - Cloud Scheduler API (for periodic triggering)

2. **Buildkite Agent Token**: Get this from your Buildkite organization's Agents page

3. **gcloud CLI**: Install from https://cloud.google.com/sdk/install

## Deployment

### 1. Set up your environment

```bash
# Configure your GCP project
export PROJECT_ID="your-gcp-project-id"
export REGION="us-central1"  # or your preferred region
export BUILDKITE_TOKEN="your-buildkite-agent-token"

# Authenticate with GCP
gcloud auth login
gcloud config set project $PROJECT_ID
```

### 2. Deploy the Cloud Function

```bash
# Navigate to the cloud_function directory
cd cloud_function

# Deploy the function
gcloud functions deploy buildkite-agent-metrics \
  --gen2 \
  --runtime=go124 \
  --region=$REGION \
  --entry-point=buildkite-agent-metrics \
  --trigger-http \
  --allow-unauthenticated \
  --set-env-vars="BUILDKITE_AGENT_TOKEN=$BUILDKITE_TOKEN,GCP_PROJECT_ID=$PROJECT_ID" \
  --memory=256MB \
  --timeout=60s \
  --max-instances=10
```

#### Optional environment variables:

```bash
# Monitor specific queues (comma-separated)
--set-env-vars="BUILDKITE_QUEUE=backend-deploy,frontend-deploy"

# Enable quiet mode (only log errors)
--set-env-vars="BUILDKITE_QUIET=true"

# Enable debug mode (verbose logging)
--set-env-vars="BUILDKITE_DEBUG=true"

# Use a custom Buildkite API endpoint
--set-env-vars="BUILDKITE_AGENT_ENDPOINT=https://custom-api.buildkite.com/v3"
```

### 3. Set up Cloud Scheduler for periodic execution

```bash
# Create a Cloud Scheduler job to trigger the function every minute
gcloud scheduler jobs create http buildkite-metrics-collector \
  --location=$REGION \
  --schedule="* * * * *" \
  --http-method=POST \
  --uri=$(gcloud functions describe buildkite-agent-metrics --region=$REGION --format='value(serviceConfig.uri)') \
  --attempt-deadline=60s
```

For different schedules, modify the `--schedule` parameter:
- Every 30 seconds: `*/30 * * * * *` (requires App Engine with automatic scaling)
- Every 5 minutes: `*/5 * * * *`
- Every hour: `0 * * * *`

## Testing

### Manual Testing

1. **Test the function locally** (requires Functions Framework):

```bash
# Install dependencies
go mod download

# Run locally
export FUNCTION_TARGET=buildkite-agent-metrics
export BUILDKITE_AGENT_TOKEN="your-token"
export GCP_PROJECT_ID="your-project-id"
go run cmd/main.go

# In another terminal, trigger the function
curl -X POST http://localhost:8080
```

2. **Test the deployed function**:

```bash
# Get the function URL
FUNCTION_URL=$(gcloud functions describe buildkite-agent-metrics \
  --region=$REGION \
  --format='value(serviceConfig.uri)')

# Trigger the function manually
curl -X POST $FUNCTION_URL

# Check the response (should return JSON with success status)
# Example response:
# {
#   "success": true,
#   "message": "Successfully collected and sent 16 metrics to Stackdriver",
#   "metrics_collected": 16
# }
```

3. **Check Cloud Function logs**:

```bash
# View recent logs
gcloud functions logs read buildkite-agent-metrics \
  --region=$REGION \
  --limit=50

# Stream logs in real-time
gcloud functions logs read buildkite-agent-metrics \
  --region=$REGION \
  --tail
```

### Verify Metrics in Cloud Monitoring

1. Go to the [Cloud Console Metrics Explorer](https://console.cloud.google.com/monitoring/metrics-explorer)

2. In the "Find resource type and metric" field, search for:
   - Resource type: `Global`
   - Metric: `custom.googleapis.com/buildkite/`

3. You should see metrics like:
   - `custom.googleapis.com/buildkite/YOUR_ORG/ScheduledJobsCount`
   - `custom.googleapis.com/buildkite/YOUR_ORG/IdleAgentCount`
   - etc.

4. Create a dashboard with these metrics for monitoring

## Setting Up Auto-scaling

### For GKE (Google Kubernetes Engine)

Use the custom metrics to scale your Buildkite agent deployments:

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: buildkite-agent-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: buildkite-agent
  minReplicas: 1
  maxReplicas: 10
  metrics:
  - type: External
    external:
      metric:
        name: custom.googleapis.com|buildkite|YOUR_ORG|UnfinishedJobsCount
        selector:
          matchLabels:
            Queue: "default"
      target:
        type: Value
        value: "2"  # Scale up when more than 2 jobs are unfinished
```

### For Compute Engine Managed Instance Groups

1. Create a custom metric-based autoscaling policy in the Cloud Console
2. Use metrics like `UnfinishedJobsCount` or `BusyAgentPercentage`
3. Set appropriate target values based on your workload

## Troubleshooting

### Common Issues

1. **"Permission denied" errors**:
   - Ensure the function's service account has the `monitoring.metricWriter` role:
   ```bash
   gcloud projects add-iam-policy-binding $PROJECT_ID \
     --member="serviceAccount:$PROJECT_ID@appspot.gserviceaccount.com" \
     --role="roles/monitoring.metricWriter"
   ```

2. **"Token invalid" errors**:
   - Verify your Buildkite token is correct
   - Check it has the necessary permissions (read access to agents and builds)

3. **No metrics appearing**:
   - Check function logs for errors
   - Ensure the function is being triggered
   - Verify the project ID is correct
   - Wait 2-3 minutes for metrics to appear in Cloud Monitoring

4. **Function timeout**:
   - Increase timeout: `--timeout=120s`
   - Check network connectivity to Buildkite API

### Debug Mode

Enable debug logging to troubleshoot issues:

```bash
gcloud functions deploy buildkite-agent-metrics \
  --update-env-vars="BUILDKITE_DEBUG=true" \
  --region=$REGION
```

## Security Considerations

1. **Use Secret Manager for tokens** (recommended for production):

```bash
# Store token in Secret Manager
echo -n "$BUILDKITE_TOKEN" | gcloud secrets create buildkite-token --data-file=-

# Grant function access to the secret
gcloud secrets add-iam-policy-binding buildkite-token \
  --member="serviceAccount:$PROJECT_ID@appspot.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"

# Deploy function with secret
gcloud functions deploy buildkite-agent-metrics \
  --set-secrets="BUILDKITE_AGENT_TOKEN=buildkite-token:latest" \
  --region=$REGION
```

2. **Restrict function access**:
   - Remove `--allow-unauthenticated` flag
   - Use Cloud Scheduler with a service account

3. **Enable VPC connector** if Buildkite API is behind a firewall

## Cost Estimation

Approximate monthly costs (based on 1-minute execution frequency):
- Cloud Functions: ~$2-5 (43,200 invocations/month)
- Cloud Monitoring: Free tier usually sufficient
- Cloud Scheduler: Free tier (up to 3 jobs)

Total: **~$2-5/month** for basic setup

## Local Development

For local development using the parent project's packages:

1. Uncomment the `replace` directive in `go.mod`:
```go
replace github.com/buildkite/buildkite-agent-metrics/v5 => ../
```

2. Run `go mod tidy` to update dependencies

3. Create a local test file `cmd/main.go`:

```go
package main

import (
	"log"
	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	_ "github.com/buildkite/buildkite-agent-metrics/cloud_function"
)

func main() {
	if err := funcframework.Start("8080"); err != nil {
		log.Fatalf("funcframework.Start: %v\n", err)
	}
}
```

4. Run: `go run cmd/main.go`

## Contributing

When making changes:
1. Update the function code in `main.go`
2. Run `go mod tidy` to update dependencies
3. Test locally and in a development environment
4. Update this README if needed
5. Deploy to production only after testing

## License

This Cloud Function is part of the buildkite-agent-metrics project and follows the same license.
