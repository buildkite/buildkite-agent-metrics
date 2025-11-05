// Package cloudfunction provides a Google Cloud Function that collects
// Buildkite CI/CD metrics and sends them to Google Cloud Monitoring (Stackdriver)
// for use in auto-scaling decisions.
package cloudfunction

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	// Google Cloud Functions framework
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"

	// Google Cloud Secret Manager
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"

	// Buildkite metrics collection packages from the published module
	"github.com/buildkite/buildkite-agent-metrics/v5/backend"
	"github.com/buildkite/buildkite-agent-metrics/v5/collector"
	"github.com/buildkite/buildkite-agent-metrics/v5/version"
)

// Environment variable constants for configuration
const (
	BKAgentTokensEnvVar           = "BUILDKITE_AGENT_TOKENS"             // Comma-separated tokens (or single token)
	BKAgentTokenSecretNamesEnvVar = "BUILDKITE_AGENT_TOKEN_SECRET_NAMES" // Comma-separated secret names (or single secret)
	BKAgentEndpointEnvVar         = "BUILDKITE_AGENT_ENDPOINT"
)

// Poll duration tracking to respect Buildkite API rate limits.
// These package-level variables persist between invocations within the same
// Cloud Function container instance (similar to Lambda behavior).
// When using multiple tokens, nextPollTime applies globally to all tokens.
var (
	nextPollTime time.Time
	lastPollTime time.Time
)

// init registers the HTTP function with the Functions Framework.
// This function is called once when the Cloud Function container starts.
func init() {
	// Register our HTTP handler function with the name "buildkite-agent-metrics"
	// This name must match the --entry-point flag when deploying
	functions.HTTP("buildkite-agent-metrics", CollectMetrics)
}

// Response represents the JSON response returned by the Cloud Function
type Response struct {
	Success         bool               `json:"success"`
	Message         string             `json:"message,omitempty"`
	Error           string             `json:"error,omitempty"`
	Metrics         int                `json:"metrics_collected,omitempty"`
	TokensProcessed int                `json:"tokens_processed,omitempty"`
	TokenErrors     []TokenErrorDetail `json:"token_errors,omitempty"`
}

// TokenErrorDetail provides details about errors for specific tokens
type TokenErrorDetail struct {
	TokenIndex int    `json:"token_index"`
	Cluster    string `json:"cluster,omitempty"`
	Error      string `json:"error"`
}

// tokenProvider represents the behavior of obtaining Buildkite tokens.
// Updated to support multiple tokens.
type tokenProvider interface {
	GetAll() ([]string, error)
}

// multiTokenProvider handles multiple token providers
type multiTokenProvider struct {
	providers []singleTokenProvider
}

// singleTokenProvider wraps a single token provider
type singleTokenProvider interface {
	Get() (string, error)
}

// GetAll returns all tokens from all providers
func (m *multiTokenProvider) GetAll() ([]string, error) {
	var tokens []string
	var errors []error

	for _, provider := range m.providers {
		token, err := provider.Get()
		if err != nil {
			errors = append(errors, err)
			continue
		}
		tokens = append(tokens, token)
	}

	if len(errors) > 0 && len(tokens) == 0 {
		// All providers failed
		return nil, fmt.Errorf("all token providers failed: %v", errors)
	}

	return tokens, nil
}

// memoryTokenProvider stores a token in memory (from environment variable)
type memoryTokenProvider struct {
	token string
}

// Get returns the stored token
func (m *memoryTokenProvider) Get() (string, error) {
	return m.token, nil
}

// secretManagerTokenProvider fetches a token from GCP Secret Manager
type secretManagerTokenProvider struct {
	client   *secretmanager.Client
	secretID string
}

// Get retrieves the token from Secret Manager
func (s *secretManagerTokenProvider) Get() (string, error) {
	ctx := context.Background()

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: s.secretID,
	}

	result, err := s.client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret %s: %w", s.secretID, err)
	}

	return string(result.Payload.Data), nil
}

// CollectMetrics is the main entry point for the Cloud Function.
// It's triggered via HTTP (typically by Cloud Scheduler) to collect Buildkite
// metrics and send them to Stackdriver for auto-scaling decisions.
//
// Poll Duration Tracking:
// The function tracks poll duration from the Buildkite API to respect rate limits.
// If called before the next allowed poll time, it returns early with a success response.
// When using multiple tokens, the function uses the maximum poll duration across all tokens,
// which means all tokens will wait for the longest duration before being polled again.
//
// Token configuration (choose one):
//   - BUILDKITE_AGENT_TOKENS: Comma-separated Buildkite API tokens or single token
//   - BUILDKITE_AGENT_TOKEN_SECRET_NAMES: Comma-separated GCP Secret Manager secret IDs or single secret
//
// Required environment variables:
//   - GCP_PROJECT_ID or GOOGLE_CLOUD_PROJECT: Google Cloud project ID for Stackdriver metrics
//
// Optional environment variables:
//   - BUILDKITE_QUEUE: Comma-separated list of specific queues to monitor
//   - BUILDKITE_AGENT_ENDPOINT: Custom Buildkite API endpoint (defaults to https://agent.buildkite.com/v3)
//   - BUILDKITE_QUIET: Set to "true" or "1" to suppress non-error logs
//   - BUILDKITE_DEBUG: Set to "true" or "1" to enable debug logging
//   - BUILDKITE_AGENT_METRICS_DEBUG_HTTP: Set to "true" or "1" to enable HTTP request/response debugging
//   - BUILDKITE_AGENT_METRICS_TIMEOUT: HTTP client timeout in seconds (default: 15)
//   - BUILDKITE_AGENT_METRICS_MAX_IDLE_CONNS: Max idle connections (default: 100)
func CollectMetrics(w http.ResponseWriter, r *http.Request) {
	// Set response header to JSON since we always return JSON
	w.Header().Set("Content-Type", "application/json")

	// Initialize our response object
	response := Response{}

	// Get project ID from environment
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		// For Cloud Functions, we can try to get the project ID from metadata
		// But it's more reliable to explicitly set it
		projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
		if projectID == "" {
			response.Success = false
			response.Error = "GCP_PROJECT_ID or GOOGLE_CLOUD_PROJECT environment variable is required"
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	// Parse optional configuration from environment
	queuesEnv := os.Getenv("BUILDKITE_QUEUE")
	endpoint := getEndpoint()
	quiet := isQuietMode()
	debug := isDebugMode()
	debugHTTP := isDebugHTTPMode()

	// Configure logging based on quiet/debug settings
	if quiet && !debug {
		// In quiet mode (without debug), suppress all non-error logs
		log.SetOutput(nullWriter{})
	}

	// Log the start of execution
	log.Printf("Starting Buildkite metrics collection for project: %s", projectID)

	// Check if we should skip this poll based on the last poll duration.
	// This applies globally to all tokens when using multiple tokens.
	startTime := time.Now()
	if !nextPollTime.IsZero() && nextPollTime.After(startTime) {
		timeUntilNextPoll := nextPollTime.Sub(startTime)
		log.Printf("Skipping polling, next poll time is in %v", timeUntilNextPoll)

		response.Success = true
		response.Message = fmt.Sprintf("Skipping polling, next poll time is in %v", timeUntilNextPoll)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Initialize token provider (supports both env vars and Secret Manager)
	provider, err := initTokenProvider()
	if err != nil {
		response.Success = false
		response.Error = fmt.Sprintf("Failed to initialize token provider: %v", err)
		log.Printf("ERROR: %s", response.Error)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get all tokens from provider
	tokens, err := provider.GetAll()
	if err != nil {
		response.Success = false
		response.Error = fmt.Sprintf("Failed to get tokens: %v", err)
		log.Printf("ERROR: %s", response.Error)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Printf("Successfully retrieved %d agent token(s)", len(tokens))

	// Parse the queue list if provided
	// If empty, we'll collect metrics for all queues in the organization
	var queues []string
	if queuesEnv != "" {
		// Split comma-separated queue names and trim whitespace
		for _, q := range strings.Split(queuesEnv, ",") {
			queue := strings.TrimSpace(q)
			if queue != "" {
				queues = append(queues, queue)
			}
		}
		log.Printf("Monitoring specific queues: %v", queues)
	} else {
		log.Println("Monitoring all queues in the organization")
	}

	// Create the Stackdriver backend for sending metrics
	// This establishes a connection to Google Cloud Monitoring API
	// Use backend.Backend interface type to allow type assertion for Closer
	var metricsBackend backend.Backend
	metricsBackend, err = backend.NewStackDriverBackend(projectID)
	if err != nil {
		response.Success = false
		response.Error = fmt.Sprintf("Failed to create Stackdriver backend: %v", err)
		log.Printf("ERROR: %s", response.Error)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Parse HTTP client configuration (matching Lambda implementation)
	timeout := os.Getenv("BUILDKITE_AGENT_METRICS_TIMEOUT")
	maxIdleConns := os.Getenv("BUILDKITE_AGENT_METRICS_MAX_IDLE_CONNS")

	configuredTimeout, err := toIntWithDefault(timeout, 15)
	if err != nil {
		response.Success = false
		response.Error = fmt.Sprintf("Invalid timeout value: %v", err)
		log.Printf("ERROR: %s", response.Error)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	configuredMaxIdleConns, err := toIntWithDefault(maxIdleConns, 100)
	if err != nil {
		response.Success = false
		response.Error = fmt.Sprintf("Invalid max idle connections value: %v", err)
		log.Printf("ERROR: %s", response.Error)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Create HTTP client with configurable timeout and connections
	httpClient := collector.NewHTTPClient(configuredTimeout, configuredMaxIdleConns)

	// Build the User-Agent string to identify our client
	userAgent := fmt.Sprintf("buildkite-agent-metrics/%s gcp-cloud-function", version.Version)

	// Track the maximum poll duration across all tokens.
	// NOTE: This means ALL tokens will wait for the longest duration returned.
	// For example, if token A returns 30s and token B returns 60s, we wait 60s
	// before polling either token again. This is conservative but may delay some
	// tokens unnecessarily.
	// TODO: Consider per-token poll duration tracking for more efficient polling.
	var pollDuration time.Duration

	// Collect metrics for each token
	totalMetrics := 0
	successfulTokens := 0
	var tokenErrors []TokenErrorDetail

	for i, token := range tokens {
		log.Printf("Processing token %d of %d", i+1, len(tokens))

		// Create collector for this token
		bkCollector := &collector.Collector{
			Client:    httpClient,
			UserAgent: userAgent,
			Endpoint:  endpoint,
			Token:     token,
			Queues:    queues,
			Quiet:     quiet,
			Debug:     debug,
			DebugHttp: debugHTTP,
		}

		// Collect metrics from Buildkite API
		log.Printf("Fetching metrics from Buildkite API for token %d...", i+1)

		result, err := bkCollector.Collect()
		if err != nil {
			// Log the error but continue with other tokens
			errorDetail := TokenErrorDetail{
				TokenIndex: i + 1,
				Error:      fmt.Sprintf("Failed to collect metrics: %v", err),
			}
			tokenErrors = append(tokenErrors, errorDetail)
			log.Printf("ERROR for token %d: %s", i+1, errorDetail.Error)
			continue
		}

		// Log what we collected (if not in quiet mode)
		if !quiet {
			log.Printf("Token %d - Collected metrics for organization: %s", i+1, result.Org)
			if result.Cluster != "" {
				log.Printf("Token %d - Cluster: %s", i+1, result.Cluster)
			}
			log.Printf("Token %d - Metrics collected: %d organization-wide, %d queue-specific",
				i+1, len(result.Totals), countQueueMetrics(result))
			// Dump detailed metrics for debugging (matching Lambda implementation)
			result.Dump()
		}

		// Send the collected metrics to Stackdriver
		log.Printf("Sending metrics to Stackdriver for token %d...", i+1)
		err = metricsBackend.Collect(result)
		if err != nil {
			// Log the error but continue with other tokens
			errorDetail := TokenErrorDetail{
				TokenIndex: i + 1,
				Cluster:    result.Cluster,
				Error:      fmt.Sprintf("Failed to send metrics to Stackdriver: %v", err),
			}
			tokenErrors = append(tokenErrors, errorDetail)
			log.Printf("ERROR for token %d: %s", i+1, errorDetail.Error)
			continue
		}

		// Update poll duration - take the maximum across all tokens
		if result.PollDuration > pollDuration {
			pollDuration = result.PollDuration
		}

		// Track successful collections
		successfulTokens++
		tokenMetricsCount := len(result.Totals) + countQueueMetrics(result)
		totalMetrics += tokenMetricsCount

		log.Printf("SUCCESS for token %d: Sent %d metrics to Stackdriver", i+1, tokenMetricsCount)
	}

	// Clean up backend resources if it implements the Closer interface
	// (matching Lambda implementation for proper resource management)
	if closer, ok := metricsBackend.(backend.Closer); ok {
		if err := closer.Close(); err != nil {
			log.Printf("WARNING: Failed to close backend: %v", err)
		}
	}

	// Update poll time tracking after successful collection.
	// This applies globally to all tokens.
	lastPollTime = time.Now()
	nextPollTime = time.Now().Add(pollDuration)
	if pollDuration > 0 {
		log.Printf("Next poll allowed in %v", pollDuration)
	}

	// Prepare response based on results
	response.TokensProcessed = len(tokens)
	response.Metrics = totalMetrics
	response.TokenErrors = tokenErrors

	if successfulTokens == 0 {
		// All tokens failed
		response.Success = false
		response.Error = fmt.Sprintf("All %d tokens failed to collect metrics", len(tokens))
		log.Printf("ERROR: %s", response.Error)
		w.WriteHeader(http.StatusInternalServerError)
	} else if len(tokenErrors) > 0 {
		// Partial success
		response.Success = true
		response.Message = fmt.Sprintf("Successfully processed %d of %d tokens, collected %d total metrics. %d token(s) had errors.",
			successfulTokens, len(tokens), totalMetrics, len(tokenErrors))
		log.Printf("PARTIAL SUCCESS: %s", response.Message)
		w.WriteHeader(http.StatusOK)
	} else {
		// Complete success
		response.Success = true
		response.Message = fmt.Sprintf("Successfully processed all %d tokens and collected %d total metrics",
			len(tokens), totalMetrics)
		log.Printf("SUCCESS: %s", response.Message)
		w.WriteHeader(http.StatusOK)
	}

	// Return our JSON response
	json.NewEncoder(w).Encode(response)
}

// initTokenProvider initializes a token provider based on environment variables.
// Token sources are checked in order:
// 1. BUILDKITE_AGENT_TOKENS (comma-separated tokens or single token via env var)
// 2. BUILDKITE_AGENT_TOKEN_SECRET_NAMES (comma-separated secret names or single secret via Secret Manager)
func initTokenProvider() (tokenProvider, error) {
	tokensEnv := os.Getenv(BKAgentTokensEnvVar)
	secretNames := os.Getenv(BKAgentTokenSecretNamesEnvVar)

	// Check that only one method is configured
	if tokensEnv != "" && secretNames != "" {
		return nil, fmt.Errorf("cannot specify both %s and %s. Use only one token configuration method",
			BKAgentTokensEnvVar, BKAgentTokenSecretNamesEnvVar)
	}

	if tokensEnv == "" && secretNames == "" {
		return nil, fmt.Errorf("must specify either %s or %s",
			BKAgentTokensEnvVar, BKAgentTokenSecretNamesEnvVar)
	}

	// Handle tokens from environment variable
	if tokensEnv != "" {
		var providers []singleTokenProvider
		tokens := strings.Split(tokensEnv, ",")
		for _, token := range tokens {
			token = strings.TrimSpace(token)
			if token == "" {
				continue
			}
			providers = append(providers, &memoryTokenProvider{token: token})
		}
		if len(providers) == 0 {
			return nil, fmt.Errorf("no valid tokens found in %s", BKAgentTokensEnvVar)
		}
		return &multiTokenProvider{providers: providers}, nil
	}

	// Handle secrets from Secret Manager
	if secretNames != "" {
		ctx := context.Background()
		smClient, err := secretmanager.NewClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create secretmanager client: %w", err)
		}

		var providers []singleTokenProvider
		secrets := strings.Split(secretNames, ",")
		for _, secret := range secrets {
			secret = strings.TrimSpace(secret)
			if secret == "" {
				continue
			}
			providers = append(providers, &secretManagerTokenProvider{
				client:   smClient,
				secretID: secret,
			})
		}
		if len(providers) == 0 {
			return nil, fmt.Errorf("no valid secrets found in %s", BKAgentTokenSecretNamesEnvVar)
		}
		return &multiTokenProvider{providers: providers}, nil
	}

	// This should never be reached due to earlier checks
	return nil, fmt.Errorf("no valid token configuration found")
}

// toIntWithDefault parses a string to int with a default value
// (matching Lambda implementation for consistency)
func toIntWithDefault(s string, defaultValue int) (int, error) {
	if s == "" {
		return defaultValue, nil
	}

	val, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("failed to parse '%s' as integer: %w", s, err)
	}

	return val, nil
}

// getEndpoint returns the Buildkite API endpoint to use.
// It checks for a custom endpoint in environment variables,
// otherwise returns the default production endpoint.
func getEndpoint() string {
	if endpoint := os.Getenv(BKAgentEndpointEnvVar); endpoint != "" {
		return endpoint
	}
	// Default to the production Buildkite API
	return "https://agent.buildkite.com/v3"
}

// isQuietMode checks if quiet mode is enabled via environment variable.
// In quiet mode, only error messages are logged.
func isQuietMode() bool {
	quiet := os.Getenv("BUILDKITE_QUIET")
	return quiet == "1" || quiet == "true"
}

// isDebugMode checks if debug mode is enabled via environment variable.
// In debug mode, additional diagnostic information is logged.
func isDebugMode() bool {
	debug := os.Getenv("BUILDKITE_DEBUG")
	return debug == "1" || debug == "true"
}

// isDebugHTTPMode checks if HTTP debug mode is enabled via environment variable.
// In HTTP debug mode, HTTP requests and responses are logged for troubleshooting.
func isDebugHTTPMode() bool {
	debugHTTP := os.Getenv("BUILDKITE_AGENT_METRICS_DEBUG_HTTP")
	return debugHTTP == "1" || debugHTTP == "true"
}

// countQueueMetrics counts the total number of metrics across all queues.
// This is used for reporting how many metrics were collected.
func countQueueMetrics(result *collector.Result) int {
	count := 0
	for _, queueMetrics := range result.Queues {
		count += len(queueMetrics)
	}
	return count
}

// nullWriter is a writer that discards all data written to it.
// Used to suppress logs in quiet mode.
type nullWriter struct{}

// Write implements io.Writer interface but discards all data
func (nullWriter) Write([]byte) (int, error) {
	return 0, nil
}

// HealthCheck is an optional endpoint that can be used to verify the function is deployed.
// You can register this as a separate function if you want a health check endpoint.
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := Response{
		Success: true,
		Message: "Buildkite metrics collector is healthy",
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
