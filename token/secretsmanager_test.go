package token

import (
	"encoding/base64"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/buildkite/buildkite-agent-metrics/token/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

//go:generate mockgen -source secretsmanager.go -mock_names SecretsManagerClient=SecretsManagerClient -package mock -destination mock/secretsmanager_client.go

const (
	secretsManagerSecretID        = "some-secret-id"
	secretsManagerSecretJSONKey   = "some_json_key"
	secretsManagerSecretValue     = "super-secret-value"
	secretsManagerSecretJSONValue = `{"some_json_key" : "super-secret-value"}`
)

func TestSecretsManagerProvider_Get_WithPlainTextSecret(t *testing.T) {
	t.Setenv(SecretsManagerKeyEnvVar, secretsManagerSecretID)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewSecretsManagerClient(ctrl)

	req := secretsmanager.GetSecretValueInput{SecretId: aws.String(secretsManagerSecretID)}
	res := secretsmanager.GetSecretValueOutput{
		SecretString: aws.String(secretsManagerSecretValue),
		SecretBinary: nil,
	}

	client.EXPECT().GetSecretValue(gomock.Eq(&req)).Return(&res, nil)

	provider := NewSecretsManager(client)

	token, err := provider.Get()
	assert.NoError(t, err)

	assert.Equal(t, secretsManagerSecretValue, token)
}

func TestSecretsManagerProvider_Get_WithBinarySecret(t *testing.T) {
	t.Setenv(SecretsManagerKeyEnvVar, secretsManagerSecretID)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewSecretsManagerClient(ctrl)
	req := secretsmanager.GetSecretValueInput{SecretId: aws.String(secretsManagerSecretID)}
	res := secretsmanager.GetSecretValueOutput{
		SecretString: nil,
		SecretBinary: stringToBase64(secretsManagerSecretValue),
	}

	client.EXPECT().GetSecretValue(gomock.Eq(&req)).Return(&res, nil)

	provider := NewSecretsManager(client)

	token, err := provider.Get()
	assert.NoError(t, err)

	assert.Equal(t, secretsManagerSecretValue, token)
}

func TestSecretsManagerProvider_Get_WithExistingJSONKey(t *testing.T) {
	t.Setenv(SecretsManagerKeyEnvVar, secretsManagerSecretID)
	t.Setenv(SecretsManagerJSONKeyEnvVar, secretsManagerSecretJSONKey)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewSecretsManagerClient(ctrl)

	req := secretsmanager.GetSecretValueInput{SecretId: aws.String(secretsManagerSecretID)}
	res := secretsmanager.GetSecretValueOutput{
		SecretString: aws.String(secretsManagerSecretJSONValue),
		SecretBinary: nil,
	}

	client.EXPECT().GetSecretValue(gomock.Eq(&req)).Return(&res, nil)

	provider := NewSecretsManager(client)

	token, err := provider.Get()
	assert.NoError(t, err)

	assert.Equal(t, secretsManagerSecretValue, token)
}

func TestSecretsManagerProvider_Get_WithBinaryJSONSecret(t *testing.T) {
	t.Setenv(SecretsManagerKeyEnvVar, secretsManagerSecretID)
	t.Setenv(SecretsManagerJSONKeyEnvVar, secretsManagerSecretJSONKey)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewSecretsManagerClient(ctrl)

	req := secretsmanager.GetSecretValueInput{SecretId: aws.String(secretsManagerSecretID)}
	res := secretsmanager.GetSecretValueOutput{
		SecretString: nil,
		SecretBinary: stringToBase64(secretsManagerSecretJSONValue),
	}

	client.EXPECT().GetSecretValue(gomock.Eq(&req)).Return(&res, nil)

	provider := NewSecretsManager(client)

	token, err := provider.Get()
	assert.NoError(t, err)

	assert.Equal(t, secretsManagerSecretValue, token)
}

func TestSecretsManagerProvider_Get_WithNonJSONPayload(t *testing.T) {
	t.Setenv(SecretsManagerKeyEnvVar, secretsManagerSecretID)
	t.Setenv(SecretsManagerJSONKeyEnvVar, secretsManagerSecretJSONKey)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewSecretsManagerClient(ctrl)

	req := secretsmanager.GetSecretValueInput{SecretId: aws.String(secretsManagerSecretID)}
	res := secretsmanager.GetSecretValueOutput{
		SecretString: aws.String("this is not a JSON payload"),
		SecretBinary: nil,
	}

	client.EXPECT().GetSecretValue(gomock.Eq(&req)).Return(&res, nil)

	provider := NewSecretsManager(client)

	_, err := provider.Get()
	assert.Error(t, err)
}

func TestSecretsManagerProvider_Get_WithNonStringValue(t *testing.T) {
	t.Setenv(SecretsManagerKeyEnvVar, secretsManagerSecretID)
	t.Setenv(SecretsManagerJSONKeyEnvVar, "non_string_value")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewSecretsManagerClient(ctrl)

	req := secretsmanager.GetSecretValueInput{SecretId: aws.String(secretsManagerSecretID)}
	res := secretsmanager.GetSecretValueOutput{
		SecretString: aws.String(`{ "non_string_value": true }`),
		SecretBinary: nil,
	}

	client.EXPECT().GetSecretValue(gomock.Eq(&req)).Return(&res, nil)

	provider := NewSecretsManager(client)

	_, err := provider.Get()
	assert.Error(t, err)
}

func stringToBase64(text string) []byte {
	data := base64.StdEncoding.EncodeToString([]byte(text))
	return []byte(data)
}
