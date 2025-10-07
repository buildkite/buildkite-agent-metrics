package token

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
)

// SSMService represents the minimal subset of interactions required to retrieve a Buildkite API token from
// AWS Systems Manager parameter store.
type SSMClient interface {
	GetParameter(*ssm.GetParameterInput) (*ssm.GetParameterOutput, error)
}

type ssmProvider struct {
	Client SSMClient
	Name   string
}

// SSMProviderOpt represents a configuration option for the AWS SSM Buildkite token provider.
type SSMProviderOpt func(prov *ssmProvider) error

// NewEnvironment constructs a Buildkite API token provider backed by AWS Systems Manager parameter store.
func NewSSM(client SSMClient, name string, opts ...SSMProviderOpt) (Provider, error) {
	provider := &ssmProvider{
		Client: client,
		Name:   name,
	}

	for _, opt := range opts {
		err := opt(provider)
		if err != nil {
			return nil, err
		}
	}

	return provider, nil
}

func (p ssmProvider) Get() (string, error) {
	output, err := p.Client.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(p.Name),
		WithDecryption: aws.Bool(true),
	})

	if err != nil {
		return "", fmt.Errorf("failed to retrieve Buildkite token (%s) from AWS SSM: %w", p.Name, err)
	}

	return *output.Parameter.Value, nil
}
