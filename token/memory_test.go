package token

import (
	"testing"
)

const (
	inMemoryTestToken = "some-token"
)

func TestInMemoryProvider_Get(t *testing.T) {
	provider, err := NewInMemory(inMemoryTestToken)
	if err != nil {
		t.Fatalf("no errors were expected to be returned by NewInMemory but got: %v", err)
	}
	token, err := provider.Get()
	if err != nil {
		t.Fatalf("no errors were expected to be returned by InMemoryProvider.Get() but got: %v", err)
	}

	if token != inMemoryTestToken {
		t.Fatalf("expecting '%s' to be returned by InMemoryProvider.Get() but got: %s", inMemoryTestToken, token)
	}
}
