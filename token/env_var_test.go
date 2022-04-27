package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testToken = "some-token"
)

func TestEnvVarProvider_Get(t *testing.T) {
	t.Setenv(TokenEnvVarKey, testToken)

	provider := NewEnvVar()
	token, err := provider.Get()
	assert.NoError(t, err)

	assert.Equal(t, testToken, token)
}

func TestEnvVarProvider_WithNoEnvVar_IsUnusable(t *testing.T) {
	provider := NewEnvVar()
	_, err := provider.Get()

	assert.Error(t, err)
	assert.ErrorIs(t, err, errNotUsable)
}
