package token

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/buildkite/buildkite-agent-metrics/v5/token/mock"
	"go.uber.org/mock/gomock"
)

//go:generate go tool mockgen -source ssm.go -mock_names SSMClient=SSMClient -package mock -destination mock/ssm_client.go

const (
	ssmTestParameterName  = "test-param"
	ssmTestParameterValue = "some-value"
)

func TestSSMProvider_New_WithErroringOpt(t *testing.T) {
	expectedErr := fmt.Errorf("some-error")

	errFunc := func(provider *ssmProvider) error {
		return expectedErr
	}

	_, err := NewSSM(nil, ssmTestParameterName, errFunc)

	if err == nil {
		t.Fatalf("expected error to be '%s' but found 'nil'", expectedErr.Error())
	}

	if err != expectedErr {
		t.Fatalf("expected error to be '%s' but found '%s'", expectedErr.Error(), err.Error())
	}
}

func TestSSMProvider_Get(t *testing.T) {
	req := ssm.GetParameterInput{
		Name:           aws.String(ssmTestParameterName),
		WithDecryption: aws.Bool(true),
	}

	res := ssm.GetParameterOutput{
		Parameter: &ssm.Parameter{
			Value: aws.String(ssmTestParameterValue),
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewSSMClient(ctrl)
	client.EXPECT().GetParameter(gomock.Eq(&req)).Return(&res, nil)

	provider, err := NewSSM(client, ssmTestParameterName)
	if err != nil {
		t.Fatalf("failed to create SSMProvider: %v", err)
	}

	token, err := provider.Get()
	if err != nil {
		t.Fatalf("failed to call 'Get()' on SSMProvider: %v", err)
	}

	if token != ssmTestParameterValue {
		t.Fatalf("expected token to be '%s' but found '%s'", ssmTestParameterValue, token)
	}
}
