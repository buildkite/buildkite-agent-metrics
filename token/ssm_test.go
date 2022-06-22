package token

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/buildkite/buildkite-agent-metrics/token/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

//go:generate mockgen -source ssm.go -mock_names SSMClient=SSMClient -package mock -destination mock/ssm_client.go

const (
	ssmTestParameterName  = "test-param"
	ssmTestParameterValue = "some-value"
)

func TestSSMProvider_Get(t *testing.T) {
	t.Setenv(SSMKeyNameEnvVar, ssmTestParameterName)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewSSMClient(ctrl)

	req := ssm.GetParameterInput{
		Name:           aws.String(ssmTestParameterName),
		WithDecryption: aws.Bool(true),
	}

	res := ssm.GetParameterOutput{
		Parameter: &ssm.Parameter{
			Value: aws.String(ssmTestParameterValue),
		},
	}
	client.EXPECT().GetParameter(gomock.Eq(&req)).Return(&res, nil)

	provider := NewSSM(client)
	token, err := provider.Get()
	assert.NoError(t, err)

	assert.Equal(t, ssmTestParameterValue, token)
}
