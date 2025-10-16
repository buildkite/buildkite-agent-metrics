// Package cloudfunction provides a Google Cloud Function that collects
// Buildkite CI/CD metrics and sends them to Google Cloud Monitoring (Stackdriver)
// for use in auto-scaling decisions.
package cloudfunction

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	// Google Cloud Functions framework
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"

	// Buildkite metrics collection packages from the published module
	"github.com/buildkite/buildkite-agent-metrics/v5/backend"
	"github.com/buildkite/buildkite-agent-metrics/v5/collector"
	"github.com/buildkite/buildkite-agent-metrics/v5/version"
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
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Metrics int    `json:"metrics_collected,omitempty"`
}

// CollectMetrics is the main entry point for the Cloud Function.
// It's triggered via HTTP (typically by Cloud Scheduler) to collect Buildkite
// metrics and send them to Stackdriver for auto-scaling decisions.
//
// Environment variables required:
//   - BUILDKITE_AGENT_TOKEN: Buildkite API token for authentication
//   - GCP_PROJECT_ID: Google Cloud project ID for Stackdriver metrics
//
// Optional environment variables:
//   - BUILDKITE_QUEUE: Comma-separated list of specific queues to monitor
//   - BUILDKITE_AGENT_ENDPOINT: Custom Buildkite API endpoint (defaults to https://agent.buildkite.com/v3)
//   - BUILDKITE_QUIET: Set to "true" or "1" to suppress non-error logs
//   - BUILDKITE_DEBUG: Set to "true" or "1" to enable debug logging
func CollectMetrics(w http.ResponseWriter, r *http.Request) {
	// Set response header to JSON since we always return JSON
	w.Header().Set("Content-Type", "application/json")

	// Initialize our response object
	response := Response{}

	// Read and validate environment variables
	token := os.Getenv("BUILDKITE_AGENT_TOKEN")
	if token == "" {
		// No token provided - this is a configuration error
		response.Success = false
		response.Error = "BUILDKITE_AGENT_TOKEN environment variable is required"
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

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

	// Configure logging based on quiet/debug settings
	if quiet && !debug {
		// In quiet mode (without debug), suppress all non-error logs
		log.SetOutput(nullWriter{})
	}

	// Log the start of execution
	log.Printf("Starting Buildkite metrics collection for project: %s", projectID)

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
	stackdriverBackend, err := backend.NewStackDriverBackend(projectID)
	if err != nil {
		response.Success = false
		response.Error = fmt.Sprintf("Failed to create Stackdriver backend: %v", err)
		log.Printf("ERROR: %s", response.Error)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Create HTTP client with reasonable timeouts for API calls
	// Using 15 seconds timeout to match the Lambda implementation
	httpClient := collector.NewHTTPClient(15, 100)

	// Build the User-Agent string to identify our client
	userAgent := fmt.Sprintf("buildkite-agent-metrics/%s gcp-cloud-function", version.Version)

	// Create the collector that will fetch metrics from Buildkite API
	// The collector handles all the API communication and data parsing
	bkCollector := &collector.Collector{
		Client:    httpClient,
		UserAgent: userAgent,
		Endpoint:  endpoint,
		Token:     token,
		Queues:    queues,
		Quiet:     quiet,
		Debug:     debug,
		DebugHttp: false, // Set to true if you need to debug HTTP requests
	}

	// Collect metrics from Buildkite API
	// This makes HTTP calls to fetch current agent and job statistics
	log.Println("Fetching metrics from Buildkite API...")
	result, err := bkCollector.Collect()
	if err != nil {
		response.Success = false
		response.Error = fmt.Sprintf("Failed to collect metrics from Buildkite: %v", err)
		log.Printf("ERROR: %s", response.Error)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Log what we collected (if not in quiet mode)
	if !quiet {
		log.Printf("Collected metrics for organization: %s", result.Org)
		if result.Cluster != "" {
			log.Printf("Cluster: %s", result.Cluster)
		}
		log.Printf("Total metrics collected: %d organization-wide, %d queue-specific",
			len(result.Totals), countQueueMetrics(result))
	}

	// Send the collected metrics to Stackdriver
	// The backend handles creating custom metrics and writing time series data
	log.Println("Sending metrics to Stackdriver...")
	err = stackdriverBackend.Collect(result)
	if err != nil {
		response.Success = false
		response.Error = fmt.Sprintf("Failed to send metrics to Stackdriver: %v", err)
		log.Printf("ERROR: %s", response.Error)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Success! Build and return the success response
	totalMetrics := len(result.Totals) + countQueueMetrics(result)
	response.Success = true
	response.Message = fmt.Sprintf("Successfully collected and sent %d metrics to Stackdriver", totalMetrics)
	response.Metrics = totalMetrics

	log.Printf("SUCCESS: %s", response.Message)

	// Return HTTP 200 OK with our JSON response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// getEndpoint returns the Buildkite API endpoint to use.
// It checks for a custom endpoint in environment variables,
// otherwise returns the default production endpoint.
func getEndpoint() string {
	if endpoint := os.Getenv("BUILDKITE_AGENT_ENDPOINT"); endpoint != "" {
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
