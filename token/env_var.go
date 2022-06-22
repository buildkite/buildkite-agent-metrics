package token

import (
	"os"
)

type envVarProvider struct {
	// The token value to provide on each Get call.
	Token string
}

const TokenEnvVarKey = "BUILDKITE_AGENT_TOKEN"

// NewEnvVar constructs a Buildkite API token provider backed by the environment variable BUILDKITE_AGENT_TOKEN
func NewEnvVar() Provider {
	return &envVarProvider{Token: os.Getenv(TokenEnvVarKey)}
}

func (p envVarProvider) Get() (string, error) {
	if p.Token == "" {
		return "", errNotUsable
	}

	return p.Token, nil
}

func (p envVarProvider) String() string {
	return "EnvVarProvider"
}
