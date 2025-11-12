# Buildkite Agent Metrics - GCP Cloud Function

This Google Cloud Function collects Buildkite CI/CD metrics and sends them to Google Cloud Monitoring (Stackdriver) for use in auto-scaling decisions.

## Overview

The function can be configured to monitor one or multiple Buildkite clusters/agent tokens in a single deployment. It periodically fetches the following metrics from the Buildkite API:
- **Job Metrics**: ScheduledJobsCount, RunningJobsCount, UnfinishedJobsCount, WaitingJobsCount
- **Agent Metrics**: IdleAgentCount, BusyAgentCount, TotalAgentCount, BusyAgentPercentage

## Multi-Token Support

The function supports collecting metrics for multiple Buildkite clusters in a single deployment. This is useful for:
- Monitoring multiple clusters within the same organization
- Reducing the number of Cloud Functions to maintain
- Centralizing metrics collection

## Prerequisites

1. **Google Cloud Project** with the following APIs enabled:
   - Cloud Functions API
   - Cloud Monitoring API
   - Cloud Scheduler API (for periodic triggering)
   - Secret Manager API (optional, for secure token storage)

2. **Buildkite Agent Token(s)**: Get these from your Buildkite organization's Agents page for each cluster you want to monitor

3. **gcloud CLI**: Install from https://cloud.google.com/sdk/install

## Configuration

The function uses a simplified configuration with two options for providing tokens:

- **BUILDKITE_AGENT_TOKENS**: Comma-separated Buildkite API tokens (or single token) via environment variable
- **BUILDKITE_AGENT_TOKEN_SECRET_NAMES**: Comma-separated GCP Secret Manager secret IDs (or single secret)

Choose one method based on your security requirements.

## Deployment

### 1. Set up your environment

```bash
# Configure your GCP project
export PROJECT_ID="your-gcp-project-id"
export REGION="us-central1"  # or your preferred region

# Authenticate with GCP
gcloud auth login
gcloud config set project $PROJECT_ID
```

### 2. Deploy the Cloud Function

#### Option A: Using Environment Variables

For single or multiple tokens directly in environment variables:

```bash
# Navigate to the cloud_function directory
cd cloud_function

# For a single token
export BUILDKITE_TOKENS="your-buildkite-agent-token"

# For multiple tokens
export BUILDKITE_TOKENS="token1,token2,token3"

gcloud functions deploy buildkite-agent-metrics \
  --gen2 \
  --runtime=go124 \
  --region=$REGION \
  --entry-point=buildkite-agent-metrics \
  --trigger-http \
  --allow-unauthenticated \
  --set-env-vars="BUILDKITE_AGENT_TOKENS=$BUILDKITE_TOKENS,GCP_PROJECT_ID=$PROJECT_ID" \
  --memory=256MB \
  --timeout=60s \
  --max-instances=10
```

#### Option B: Using Secret Manager (Recommended for Production)
```bash
# Create a secret for a single agent token
echo -n "$BUILDKITE_TOKEN" | gcloud secrets create buildkite-token --data-file=-

# Deploy the function with Secret Manager reference
gcloud functions deploy buildkite-agent-metrics \
  --gen2 \
  --runtime=go124 \
  --region=$REGION \
  --entry-point=buildkite-agent-metrics \
  --trigger-http \
  --allow-unauthenticated \
  --set-env-vars="BUILDKITE_AGENT_TOKEN_SECRET_NAMES=projects/${PROJECT_ID}/secrets/buildkite-token/versions/latest,GCP_PROJECT_ID=$PROJECT_ID" \
  --memory=256MB \
  --timeout=60s \
  --max-instances=10
```

##### Multiple Tokens:
```bash
# Create secrets for multiple agent tokens
echo -n "token1" | gcloud secrets create buildkite-token-cluster1 --data-file=-
echo -n "token2" | gcloud secrets create buildkite-token-cluster2 --data-file=-
echo -n "token3" | gcloud secrets create buildkite-token-cluster3 --data-file=-

# Deploy with multiple Secret Manager references (comma-separated)
gcloud functions deploy buildkite-agent-metrics \
  --gen2 \
  --runtime=go124 \
  --region=$REGION \
  --entry-point=buildkite-agent-metrics \
  --trigger-http \
  --allow-unauthenticated \
  --set-env-vars="BUILDKITE_AGENT_TOKEN_SECRET_NAMES=projects/${PROJECT_ID}/secrets/buildkite-token-cluster1/versions/latest,projects/${PROJECT_ID}/secrets/buildkite-token-cluster2/versions/latest,projects/${PROJECT_ID}/secrets/buildkite-token-cluster3/versions/latest,GCP_PROJECT_ID=$PROJECT_ID" \
  --memory=256MB \
  --timeout=60s \
  --max-instances=10
```

#### Optional environment variables:

```bash
# Monitor specific queues within the cluster (comma-separated)
--set-env-vars="BUILDKITE_QUEUE=backend-deploy,frontend-deploy"

# Enable quiet mode (only log errors)
--set-env-vars="BUILDKITE_QUIET=true"

# Enable debug mode (verbose logging)
--set-env-vars="BUILDKITE_DEBUG=true"

# Enable HTTP debug mode (log HTTP requests/responses)
--set-env-vars="BUILDKITE_AGENT_METRICS_DEBUG_HTTP=true"

# Use a custom Buildkite API endpoint
--set-env-vars="BUILDKITE_AGENT_ENDPOINT=https://custom-api.buildkite.com/v3"

# Configure HTTP client settings
--set-env-vars="BUILDKITE_AGENT_METRICS_TIMEOUT=30"  # seconds
--set-env-vars="BUILDKITE_AGENT_METRICS_MAX_IDLE_CONNS=50"
```

### 3. Set up Cloud Scheduler for periodic execution

Create a scheduler job for each deployed function:

```bash
# Create a Cloud Scheduler job to trigger the function every minute
gcloud scheduler jobs create http buildkite-agent-metrics-${CLUSTER_NAME}-scheduler \
  --location=$REGION \
  --schedule="* * * * *" \
  --http-method=POST \
  --uri=$(gcloud functions describe buildkite-agent-metrics-${CLUSTER_NAME} --region=$REGION --format='value(serviceConfig.uri)') \
  --attempt-deadline=60s
```

For different schedules, modify the `--schedule` parameter:
- Every 30 seconds: `*/30 * * * * *` (requires App Engine with automatic scaling)
- Every 5 minutes: `*/5 * * * *`
- Every hour: `0 * * * *`

## Testing

### Manual Testing

1. **Test the deployed function**:

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

2. **Check Cloud Function logs**:

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
