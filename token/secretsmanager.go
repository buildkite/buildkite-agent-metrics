package token

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

// SecretsManagerOpt represents a configuration option for the AWS SecretsManager Buildkite token provider.
type SecretsManagerOpt func(opts *secretsManagerProvider) error

// SecretsManagerClient represents the minimal interactions required to retrieve a Buildkite API token from
// AWS Secrets Manager.
type SecretsManagerClient interface {
	GetSecretValue(*secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error)
}

type secretsManagerProvider struct {
	Client   SecretsManagerClient
	SecretID string
	JSONKey  string
}

// WithSecretsManagerJSONSecret instructs SecretsManager Buidlkite token provider that the token is stored within a JSON
// payload. The key parameter specifies the JSON field holding the secret value within the JSON blob.
//
// This configuration option works for both AWS supported secret formats (SecretString and SecretBinary). However, for
// the later case, the binary payload must be a valid JSON document containing the 'key' field.
func WithSecretsManagerJSONSecret(key string) SecretsManagerOpt {
	return func(provider *secretsManagerProvider) error {
		provider.JSONKey = key
		return nil
	}
}

// NewSecretsManager constructs a Buildkite API token provider backed by AWS Secrets Manager.
func NewSecretsManager(client SecretsManagerClient, secretID string, opts ...SecretsManagerOpt) (Provider, error) {
	provider := &secretsManagerProvider{
		Client:   client,
		SecretID: secretID,
	}

	for _, opt := range opts {
		err := opt(provider)
		if err != nil {
			return nil, err
		}
	}

	return provider, nil
}

func (p secretsManagerProvider) Get() (string, error) {
	res, err := p.Client.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: aws.String(p.SecretID),
	})

	if err != nil {
		return "", fmt.Errorf("failed to retrieve secret '%s' from SecretsManager: %v", p.SecretID, err)
	}

	secret, err := p.parseResponse(res)
	if err != nil {
		return "", fmt.Errorf("failed to parse SecretsManager's response for '%s': %v", p.SecretID, err)
	}

	return secret, nil
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

	if p.JSONKey != "" {
		secret, err := extractStringKeyFromJSON(secretBytes, p.JSONKey)
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
