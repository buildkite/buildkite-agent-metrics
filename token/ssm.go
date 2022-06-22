package token

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
)

const SSMKeyNameEnvVar = "BUILDKITE_AGENT_TOKEN_SSM_KEY"

// SSMService represents the minimal subset of interactions required to retrieve a Buildkite API token from
// AWS Systems Manager parameter store.
type SSMClient interface {
	GetParameter(*ssm.GetParameterInput) (*ssm.GetParameterOutput, error)
}

type ssmProvider struct {
	Client SSMClient
}

// NewEnvironment constructs a Buildkite API token provider backed by AWS Systems Manager parameter store.
func NewSSM(client SSMClient) Provider {
	return &ssmProvider{Client: client}
}

func (p ssmProvider) Get() (string, error) {
	keyName := os.Getenv(SSMKeyNameEnvVar)

	if keyName == "" {
		return "", errNotUsable
	}

	output, err := p.Client.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(keyName),
		WithDecryption: aws.Bool(true),
	})

	if err != nil {
		return "", fmt.Errorf("failed to retrieve Buildkite token (%s) from AWS SSM: %v", keyName, err)
	}

	return *output.Parameter.Value, nil
}

func (p ssmProvider) String() string {
	return "SSMProvider"
}
