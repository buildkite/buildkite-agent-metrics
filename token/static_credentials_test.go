package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStaticCredentials_Get(t *testing.T) {
	tok := "super secret"

	provider := NewStaticCredentialsProvider(&tok)

	token, err := provider.Get()

	assert.NoError(t, err)
	assert.Equal(t, tok, token)
}

func TestStaticCredentials_Get_WithNilPtr_ReturnsErr(t *testing.T) {
	provider := NewStaticCredentialsProvider(nil)

	_, err := provider.Get()

	assert.Error(t, err)
	assert.ErrorIs(t, err, errNotUsable)
}
