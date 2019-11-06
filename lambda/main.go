package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
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
	var b backend.Backend
	var provider token.Provider
	var bkToken string
	var err error

	awsRegion := os.Getenv("AWS_REGION")
	backendOpt := os.Getenv("BUILDKITE_BACKEND")
	queue := os.Getenv("BUILDKITE_QUEUE")
	clwDimensions := os.Getenv("BUILDKITE_CLOUDWATCH_DIMENSIONS")
	quietString := os.Getenv("BUILDKITE_QUIET")
	quiet := quietString == "1" || quietString == "true"

	if quiet {
		log.SetOutput(ioutil.Discard)
	}

	t := time.Now()

	if !nextPollTime.IsZero() && nextPollTime.After(t) {
		log.Printf("Skipping polling, next poll time is in %v",
			nextPollTime.Sub(t))
		return "", nil
	}

	provider, err = initTokenProvider(awsRegion)
	if err != nil {
		return "", err
	}

	bkToken, err = provider.Get()
	if err != nil {
		return "", err
	}

	queues := []string{}
	if queue != "" {
		queues = strings.Split(queue, ",")
	}

	userAgent := fmt.Sprintf("buildkite-agent-metrics/%s buildkite-agent-metrics-lambda", version.Version)

	c := collector.Collector{
		UserAgent: userAgent,
		Endpoint:  "https://agent.buildkite.com/v3",
		Token:     bkToken,
		Queues:    queues,
		Quiet:     quiet,
		Debug:     false,
		DebugHttp: false,
	}

	switch strings.ToLower(backendOpt) {
	case "statsd":
		statsdHost := os.Getenv("STATSD_HOST")
		statsdTags := strings.EqualFold(os.Getenv("STATSD_TAGS"), "true")
		b, err = backend.NewStatsDBackend(statsdHost, statsdTags)
		if err != nil {
			return "", err
		}
	case "newrelic":
		nrAppName := os.Getenv("NEWRELIC_APP_NAME")
		nrLicenseKey := os.Getenv("NEWRELIC_LICENSE_KEY")
		b, err = backend.NewNewRelicBackend(nrAppName, nrLicenseKey)
		if err != nil {
			fmt.Printf("Error starting New Relic client: %v\n", err)
			os.Exit(1)
		}
	default:
		dimensions, err := backend.ParseCloudWatchDimensions(clwDimensions)
		if err != nil {
			return "", err
		}
		b = backend.NewCloudWatchBackend(awsRegion, dimensions)
	}

	res, err := c.Collect()
	if err != nil {
		return "", err
	}

	res.Dump()

	err = b.Collect(res)
	if err != nil {
		return "", err
	}

	original, ok := b.(backend.Closer)
	if ok {
		err := original.Close()
		if err != nil {
			return "", err
		}
	}

	log.Printf("Finished in %s", time.Now().Sub(t))

	// Store the next acceptable poll time in global state
	nextPollTime = time.Now().Add(res.PollDuration)

	return "", nil
}

func initTokenProvider(awsRegion string) (token.Provider, error) {
	mutuallyExclusiveEnvVars := []string{
		BKAgentTokenEnvVar,
		BKAgentTokenSSMKeyEnvVar,
		BKAgentTokenSecretsManagerSecretIDEnvVar,
	}

	if err := checkMutuallyExclusiveEnvVars(mutuallyExclusiveEnvVars...); err != nil {
		return nil, err
	}

	if bkToken := os.Getenv(BKAgentTokenEnvVar); bkToken != "" {
		return token.NewInMemory(bkToken)
	}

	if ssmKey := os.Getenv(BKAgentTokenSSMKeyEnvVar); ssmKey != "" {
		sess, err := session.NewSession(&aws.Config{Region: aws.String(awsRegion)})
		if err != nil {
			return nil, err
		}
		client := ssm.New(sess)
		return token.NewSSM(client, ssmKey)
	}

	if secretsManagerSecretID := os.Getenv(BKAgentTokenSecretsManagerSecretIDEnvVar); secretsManagerSecretID != "" {
		jsonKey := os.Getenv(BKAgentTokenSecretsManagerJSONKeyEnvVar)
		sess, err := session.NewSession(&aws.Config{Region: aws.String(awsRegion)})
		if err != nil {
			return nil, err
		}
		client := secretsmanager.New(sess)
		if jsonKey == "" {
			return token.NewSecretsManager(client, secretsManagerSecretID, token.WithSecretsManagerJSONSecret(jsonKey))
		} else {
			return token.NewSecretsManager(client, secretsManagerSecretID)
		}
	}

	return nil, fmt.Errorf("failed to initialize Buildkite token provider: one of the [%s] environment variables "+
		"must be provided", strings.Join(mutuallyExclusiveEnvVars, ","))
}

func checkMutuallyExclusiveEnvVars(varNames ...string) error {
	foundVars := make([]string, 0)
	for _, varName := range varNames {
		value := os.Getenv(varName)
		if value != "" {
			foundVars = append(foundVars, value)
		}
	}
	if len(foundVars) > 1 {
		return fmt.Errorf("the environment variables [%s] are mutually exclusive", strings.Join(foundVars, ","))
	}
	return nil
}
