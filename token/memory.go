package token

type inMemoryProvider struct {
	// The token value to provide on each Get call.
	Token string
}

// NewInMemory constructs a Buildkite API token provider backed by a in-memory string.
func NewInMemory(token string) (Provider, error) {
	return &inMemoryProvider{Token: token}, nil
}

func (p inMemoryProvider) Get() (string, error) {
	return p.Token, nil
}
