package token

import (
	"errors"
	"fmt"
)

type ProviderChain struct {
	providers []Provider
}

func NewProviderChain(providers ...Provider) *ProviderChain {
	return &ProviderChain{
		providers: providers,
	}
}

// Append adds a new token provider to the end of the chain, meaning it will be evaluated last
func (p *ProviderChain) Append(provider Provider) {
	p.providers = append(p.providers, provider)
}

// Prepend adds a new token provider to the end of the chain, meaning it will be evaluated first
func (p *ProviderChain) Prepend(provider Provider) {
	newProviders := make([]Provider, 0, len(p.providers)+1)
	newProviders = append(newProviders, provider)
	newProviders = append(newProviders, p.providers...)

	p.providers = newProviders
}

// Resolve iterates through the providers in the chain, looking for one that's usable, and returning the token that the first
// usable provider returns
func (p *ProviderChain) Resolve() (string, error) {
	for _, provider := range p.providers {
		token, err := provider.Get()
		if err != nil {
			if errors.Is(err, errNotUsable) {
				// The provider doesn't have the necessary config to fetch a token
				continue
			}
			return "", err
		}

		return token, nil
	}

	return "", fmt.Errorf("no providers in the chain %v were able to resolve a token", p.providers)
}
