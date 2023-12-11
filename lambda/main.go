package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/ssm"

	"github.com/buildkite/buildkite-agent-metrics/backend"
	"github.com/buildkite/buildkite-agent-metrics/collector"
	"github.com/buildkite/buildkite-agent-metrics/token"
	"github.com/buildkite/buildkite-agent-metrics/version"
)

const (
	BKAgentTokenEnvVar                       = "BUILDKITE_AGENT_TOKEN"
	BKAgentTokenSSMKeyEnvVar                 = "BUILDKITE_AGENT_TOKEN_SSM_KEY"
	BKAgentTokenSecretsManagerSecretIDEnvVar = "BUILDKITE_AGENT_SECRETS_MANAGER_SECRET_ID"
	BKAgentTokenSecretsManagerJSONKeyEnvVar  = "BUILDKITE_AGENT_SECRETS_MANAGER_JSON_KEY"
)

var (
	nextPollTime time.Time
)

func main() {
	if os.Getenv(`DEBUG`) != "" {
		_, err := Handler(context.Background(), json.RawMessage([]byte{}))
		if err != nil {
			log.Fatal(err)
		}
	} else {
		lambda.Start(Handler)
	}
}

func Handler(ctx context.Context, evt json.RawMessage) (string, error) {
	// Where we send metrics
	var metricsBackend backend.Backend

	awsRegion := os.Getenv("AWS_REGION")
	backendOpt := os.Getenv("BUILDKITE_BACKEND")
	queue := os.Getenv("BUILDKITE_QUEUE")
	clwDimensions := os.Getenv("BUILDKITE_CLOUDWATCH_DIMENSIONS")
	quietString := os.Getenv("BUILDKITE_QUIET")
	quiet := quietString == "1" || quietString == "true"
	timeout := os.Getenv("BUILDKITE_AGENT_METRICS_TIMEOUT")

	debugEnvVar := os.Getenv("BUILDKITE_AGENT_METRICS_DEBUG")
	debug := debugEnvVar == "1" || debugEnvVar == "true"

	debugHTTPEnvVar := os.Getenv("BUILDKITE_AGENT_METRICS_DEBUG_HTTP")
	debugHTTP := debugHTTPEnvVar == "1" || debugHTTPEnvVar == "true"

	if quiet {
		log.SetOutput(io.Discard)
	}

	startTime := time.Now()

	if !nextPollTime.IsZero() && nextPollTime.After(startTime) {
		log.Printf("Skipping polling, next poll time is in %v",
			nextPollTime.Sub(startTime))
		return "", nil
	}

	providers, err := initTokenProvider(awsRegion)
	if err != nil {
		return "", err
	}

	tokens := make([]string, 0)
	for _, provider := range providers {
		bkToken, err := provider.Get()
		if err != nil {
			return "", err
		}
		tokens = append(tokens, bkToken)
	}

	queues := []string{}
	if queue != "" {
		queues = strings.Split(queue, ",")
	}

	if timeout == "" {
		timeout = "15"
	}

	configuredTimeout, err := strconv.Atoi(timeout)

	if err != nil {
		return "", err
	}

	userAgent := fmt.Sprintf("buildkite-agent-metrics/%s buildkite-agent-metrics-lambda", version.Version)

	collectors := make([]*collector.Collector, 0, len(tokens))
	for _, token := range tokens {
		collectors = append(collectors, &collector.Collector{
			UserAgent: userAgent,
			Endpoint:  "https://agent.buildkite.com/v3",
			Token:     token,
			Queues:    queues,
			Quiet:     quiet,
			Debug:     debug,
			DebugHttp: debugHTTP,
			Timeout:   configuredTimeout,
		})
	}

	switch strings.ToLower(backendOpt) {
	case "statsd":
		statsdHost := os.Getenv("STATSD_HOST")
		statsdTags := strings.EqualFold(os.Getenv("STATSD_TAGS"), "true")
		metricsBackend, err = backend.NewStatsDBackend(statsdHost, statsdTags)
		if err != nil {
			return "", err
		}

	case "newrelic":
		nrAppName := os.Getenv("NEWRELIC_APP_NAME")
		nrLicenseKey := os.Getenv("NEWRELIC_LICENSE_KEY")
		metricsBackend, err = backend.NewNewRelicBackend(nrAppName, nrLicenseKey)
		if err != nil {
			fmt.Printf("Error starting New Relic client: %v\n", err)
			os.Exit(1)
		}

	default:
		dimensions, err := backend.ParseCloudWatchDimensions(clwDimensions)
		if err != nil {
			return "", err
		}
		metricsBackend = backend.NewCloudWatchBackend(awsRegion, dimensions)
	}

	// minimum res.PollDuration across collectors
	var pollDuration time.Duration

	for _, c := range collectors {
		res, err := c.Collect()
		if err != nil {
			return "", err
		}

		if res.PollDuration > pollDuration {
			pollDuration = res.PollDuration
		}

		res.Dump()

		if err := metricsBackend.Collect(res); err != nil {
			return "", err
		}
	}

	original, ok := metricsBackend.(backend.Closer)
	if ok {
		err := original.Close()
		if err != nil {
			return "", err
		}
	}

	log.Printf("Finished in %s", time.Since(startTime))

	// Store the next acceptable poll time in global state
	nextPollTime = time.Now().Add(pollDuration)

	return "", nil
}

func initTokenProvider(awsRegion string) ([]token.Provider, error) {
	err := checkMutuallyExclusiveEnvVars(
		BKAgentTokenEnvVar,
		BKAgentTokenSSMKeyEnvVar,
		BKAgentTokenSecretsManagerSecretIDEnvVar,
	)
	if err != nil {
		return nil, err
	}

	var providers []token.Provider
	if bkTokenEnvVar := os.Getenv(BKAgentTokenEnvVar); bkTokenEnvVar != "" {
		bkTokens := strings.Split(bkTokenEnvVar, ",")
		for _, bkToken := range bkTokens {
			provider, err := token.NewInMemory(bkToken)
			if err != nil {
				return nil, err
			}
			providers = append(providers, provider)
		}
	}

	if ssmKey := os.Getenv(BKAgentTokenSSMKeyEnvVar); ssmKey != "" {
		sess, err := session.NewSession(&aws.Config{Region: aws.String(awsRegion)})
		if err != nil {
			return nil, err
		}
		client := ssm.New(sess)
		provider, err := token.NewSSM(client, ssmKey)
		if err != nil {
			return nil, err
		}
		providers = append(providers, provider)
	}

	if secretsManagerSecretID := os.Getenv(BKAgentTokenSecretsManagerSecretIDEnvVar); secretsManagerSecretID != "" {
		jsonKey := os.Getenv(BKAgentTokenSecretsManagerJSONKeyEnvVar)
		sess, err := session.NewSession(&aws.Config{Region: aws.String(awsRegion)})
		if err != nil {
			return nil, err
		}
		client := secretsmanager.New(sess)
		if jsonKey == "" {
			secretIDs := strings.Split(secretsManagerSecretID, ",")
			for _, secretID := range secretIDs {
				secretManager, err := token.NewSecretsManager(client, secretID)
				if err != nil {
					return nil, err
				}
				providers = append(providers, secretManager)
			}
		} else {
			secretManager, err := token.NewSecretsManager(client, secretsManagerSecretID, token.WithSecretsManagerJSONSecret(jsonKey))
			if err != nil {
				return nil, err
			}
			providers = append(providers, secretManager)
		}
	}

	if len(providers) == 0 {
		// This should be very unlikely or even impossible (famous last words):
		// - There was exactly one of the mutually-exclusive env vars
		// - If a token provider above failed to use its value, it should error
		// - Otherwise, each if-branch appends to providers, so...
		return nil, fmt.Errorf("no Buildkite token providers could be created")
	}

	return providers, nil
}

func checkMutuallyExclusiveEnvVars(varNames ...string) error {
	foundVars := make([]string, 0)
	for _, varName := range varNames {
		value := os.Getenv(varName)
		if value != "" {
			foundVars = append(foundVars, value)
		}
	}
	switch len(foundVars) {
	case 0:
		return fmt.Errorf("one of the environment variables [%s] must be provided", strings.Join(varNames, ","))

	case 1:
		return nil // that's what we want

	default:
		return fmt.Errorf("the environment variables [%s] are mutually exclusive", strings.Join(foundVars, ","))
	}
}
