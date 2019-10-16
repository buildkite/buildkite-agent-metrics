package token

import (
	"fmt"
	"os"
)

// Provider represents the behaviour of obtaining a Buildkite token.
type Provider interface {
	Get() (string, error)
}

// Must is a helper function to ensure a Provider object can be successfully instantiated when calling any of the
// constructor functions provided by this package.
//
// This helper is intended to be used at program startup to load the Provider implementation to be used. Such as:
//   var provider := token.Must(token.NewSSMProvider())
func Must(prov Provider, err error) Provider {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to initialize Buildkite token provider: %v", err)
		os.Exit(1)
	}
	return prov
}
