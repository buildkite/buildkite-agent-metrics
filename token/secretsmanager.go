package token

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

const (
	SecretsManagerKeyEnvVar     = "BUILDKITE_AGENT_SECRETS_MANAGER_SECRET_ID"
	SecretsManagerJSONKeyEnvVar = "BUILDKITE_AGENT_SECRETS_MANAGER_JSON_KEY"
)

// SecretsManagerClient represents the minimal interactions required to retrieve a Buildkite API token from
// AWS Secrets Manager.
type SecretsManagerClient interface {
	GetSecretValue(*secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error)
}

type secretsManagerProvider struct {
	Client SecretsManagerClient
}

// NewSecretsManager constructs a Buildkite API token provider backed by AWS Secrets Manager.
func NewSecretsManager(client SecretsManagerClient) Provider {
	return &secretsManagerProvider{
		Client: client,
	}
}

func (p secretsManagerProvider) Get() (string, error) {
	secretID := os.Getenv(SecretsManagerKeyEnvVar)

	if secretID == "" {
		return "", errNotUsable
	}

	res, err := p.Client.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretID),
	})

	if err != nil {
		return "", fmt.Errorf("failed to retrieve secret '%s' from SecretsManager: %v", secretID, err)
	}

	secret, err := p.parseResponse(res)
	if err != nil {
		return "", fmt.Errorf("failed to parse SecretsManager's response for '%s': %v", secretID, err)
	}

	return secret, nil
}

func (p secretsManagerProvider) IsUsable() bool {
	return os.Getenv(SecretsManagerKeyEnvVar) != ""
}

func (p secretsManagerProvider) String() string {
	return "SecretsManagerProvider"
}

func (p secretsManagerProvider) parseResponse(res *secretsmanager.GetSecretValueOutput) (string, error) {
	var err error
	var secretBytes []byte

	if res.SecretString != nil {
		secretBytes = []byte(*res.SecretString)
	} else {
		secretBytes, err = decodeBase64(res.SecretBinary)
		if err != nil {
			return "", err
		}
	}

	jsonKey := os.Getenv(SecretsManagerJSONKeyEnvVar)
	if jsonKey != "" {
		secret, err := extractStringKeyFromJSON(secretBytes, jsonKey)
		if err != nil {
			return "", err
		}
		return secret, nil
	}

	return string(secretBytes), nil
}

func extractStringKeyFromJSON(data []byte, key string) (string, error) {
	contents := map[string]interface{}{}
	err := json.Unmarshal(data, &contents)
	if err != nil {
		return "", err
	}

	// Checks whether the provided data is a valid JSON, contains the requested key and its corresponding value
	// is a string.
	if secretValue, ok := contents[key].(string); ok {
		return secretValue, nil
	} else {
		return "", fmt.Errorf("key '%s' doesn't exist or isn't a string value", key)
	}
}

func decodeBase64(data []byte) ([]byte, error) {
	decodedBytes := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
	size, err := base64.StdEncoding.Decode(decodedBytes, data)
	if err != nil {
		return nil, err
	}
	return decodedBytes[:size], nil
}
