package token

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/buildkite/buildkite-agent-metrics/v5/token/mock"
	"github.com/golang/mock/gomock"
)

//go:generate mockgen -source secretsmanager.go -mock_names SecretsManagerClient=SecretsManagerClient -package mock -destination mock/secretsmanager_client.go

const (
	secretsManagerSecretID        = "some-secret-id"
	secretsManagerSecretJSONKey   = "some_json_key"
	secretsManagerSecretValue     = "super-secret-value"
	secretsManagerSecretJSONValue = `{"some_json_key" : "super-secret-value"}`
)

func TestSecretsManagerProvider_WithSecretsManagerJSONSecret(t *testing.T) {
	provider := secretsManagerProvider{}

	err := WithSecretsManagerJSONSecret(secretsManagerSecretJSONKey)(&provider)
	if err != nil {
		t.Fatalf("failed to apply WithJSONSecret: %v", err)
	}

	if provider.JSONKey != secretsManagerSecretJSONKey {
		t.Fatalf(
			"expected secretsManagerProvider.JSONKey to be %s but found %s",
			secretsManagerSecretJSONKey, provider.JSONKey)
	}
}

func TestSecretsManagerProvider_New_WithErroringOpt(t *testing.T) {
	expectedErr := fmt.Errorf("some-error")

	errFunc := func(provider *secretsManagerProvider) error {
		return expectedErr
	}

	_, err := NewSecretsManager(nil, secretsManagerSecretID, errFunc)

	if err == nil {
		t.Fatalf("expected error to be '%s' but found 'nil'", expectedErr.Error())
	}

	if err != expectedErr {
		t.Fatalf("expected error to be '%s' but found '%s'", expectedErr.Error(), err.Error())
	}
}

func TestSecretsManagerProvider_Get_WithPlainTextSecret(t *testing.T) {
	req := secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretsManagerSecretID),
	}
	res := secretsmanager.GetSecretValueOutput{
		SecretString: aws.String(secretsManagerSecretValue),
		SecretBinary: nil,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewSecretsManagerClient(ctrl)
	client.EXPECT().GetSecretValue(gomock.Eq(&req)).Return(&res, nil)

	provider, err := NewSecretsManager(client, secretsManagerSecretID)
	if err != nil {
		t.Fatalf("failed to create SecretsManagerProvider: %v", err)
	}

	token, err := provider.Get()
	if err != nil {
		t.Error(err)
	}

	if token != secretsManagerSecretValue {
		t.Fatalf("expected token to be '%s' but found '%s'", secretsManagerSecretValue, token)
	}
}

func TestSecretsManagerProvider_Get_WithBinarySecret(t *testing.T) {
	req := secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretsManagerSecretID),
	}
	res := secretsmanager.GetSecretValueOutput{
		SecretString: nil,
		SecretBinary: stringToBase64(secretsManagerSecretValue),
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewSecretsManagerClient(ctrl)
	client.EXPECT().GetSecretValue(gomock.Eq(&req)).Return(&res, nil)

	provider, err := NewSecretsManager(client, secretsManagerSecretID)
	if err != nil {
		t.Fatalf("failed to create SecretsManagerProvider: %v", err)
	}

	token, err := provider.Get()
	if err != nil {
		t.Error(err)
	}

	if token != secretsManagerSecretValue {
		t.Fatalf("expected token to be '%s' but found '%s'", secretsManagerSecretValue, token)
	}
}

func TestSecretsManagerProvider_Get_WithExistingJSONKey(t *testing.T) {
	req := secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretsManagerSecretID),
	}
	res := secretsmanager.GetSecretValueOutput{
		SecretString: aws.String(secretsManagerSecretJSONValue),
		SecretBinary: nil,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewSecretsManagerClient(ctrl)
	client.EXPECT().GetSecretValue(gomock.Eq(&req)).Return(&res, nil)

	provider, err := NewSecretsManager(client, secretsManagerSecretID,
		WithSecretsManagerJSONSecret(secretsManagerSecretJSONKey))

	if err != nil {
		t.Fatalf("failed to create SecretsManagerProvider: %v", err)
	}

	token, err := provider.Get()
	if err != nil {
		t.Error(err)
	}

	if token != secretsManagerSecretValue {
		t.Fatalf("expected token to be '%s' but found '%s'", secretsManagerSecretValue, token)
	}
}

func TestSecretsManagerProvider_Get_WithBinaryJSONSecret(t *testing.T) {
	req := secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretsManagerSecretID),
	}
	res := secretsmanager.GetSecretValueOutput{
		SecretString: nil,
		SecretBinary: stringToBase64(secretsManagerSecretJSONValue),
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewSecretsManagerClient(ctrl)
	client.EXPECT().GetSecretValue(gomock.Eq(&req)).Return(&res, nil)

	provider, err := NewSecretsManager(client, secretsManagerSecretID,
		WithSecretsManagerJSONSecret(secretsManagerSecretJSONKey))

	if err != nil {
		t.Fatalf("failed to create SecretsManagerProvider: %v", err)
	}

	token, err := provider.Get()
	if err != nil {
		t.Error(err)
	}

	if token != secretsManagerSecretValue {
		t.Fatalf("expected token to be '%s' but found '%s'", secretsManagerSecretValue, token)
	}
}

func TestSecretsManagerProvider_Get_WithNonJSONPayload(t *testing.T) {
	req := secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretsManagerSecretID),
	}
	res := secretsmanager.GetSecretValueOutput{
		SecretString: aws.String("this is not a JSON payload"),
		SecretBinary: nil,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewSecretsManagerClient(ctrl)
	client.EXPECT().GetSecretValue(gomock.Eq(&req)).Return(&res, nil)

	provider, err := NewSecretsManager(client, secretsManagerSecretID,
		WithSecretsManagerJSONSecret(secretsManagerSecretJSONKey))

	if err != nil {
		t.Fatalf("failed to create SecretsManagerProvider: %v", err)
	}

	_, err = provider.Get()
	if err == nil {
		t.Fatalf("expecting error when extracting JSON key from a non-JSON payload")
	}
}

func TestSecretsManagerProvider_Get_WithNonStringValue(t *testing.T) {
	req := secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretsManagerSecretID),
	}
	res := secretsmanager.GetSecretValueOutput{
		SecretString: aws.String(`{ "non_string_value": true }`),
		SecretBinary: nil,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewSecretsManagerClient(ctrl)
	client.EXPECT().GetSecretValue(gomock.Eq(&req)).Return(&res, nil)

	provider, err := NewSecretsManager(client, secretsManagerSecretID,
		WithSecretsManagerJSONSecret("non_string_value"))

	if err != nil {
		t.Fatalf("failed to create SecretsManagerProvider: %v", err)
	}

	_, err = provider.Get()
	if err == nil {
		t.Fatalf("expecting error when extracting a non-string value from JSON payload")
	}
}

func stringToBase64(text string) []byte {
	data := base64.StdEncoding.EncodeToString([]byte(text))
	return []byte(data)
}
