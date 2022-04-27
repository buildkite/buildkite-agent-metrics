package token

import (
	"errors"
)

// Provider represents the behaviour of obtaining a Buildkite token.
type Provider interface {
	Get() (string, error)
	String() string
}

var errNotUsable = errors.New("provider not usable")
