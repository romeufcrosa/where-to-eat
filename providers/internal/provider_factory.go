package internal

import (
	"errors"
	"sync"
)

// ErrProviderNotRegistered error for unregistered providers
var ErrProviderNotRegistered = errors.New("provider not registered")

// Registration function to call when registrating a provider
type Registration func() (interface{}, error)

// Factory provides registration and loading of providers
type Factory struct {
	registration Registration
	provider     interface{}
}

// ProviderFactory provides access to reloading of factories
type ProviderFactory struct {
	mu      sync.Mutex
	factory Factory
}

// NewProviderFactory returns a new ProviderFactory
func NewProviderFactory(registration Registration) ProviderFactory {
	return ProviderFactory{
		factory: Factory{
			registration: registration,
		},
	}
}

// Register registers the given registration with the given name
func (factory *ProviderFactory) Register() error {
	factory.mu.Lock()
	defer factory.mu.Unlock()

	provider, err := factory.factory.registration()
	if err != nil {
		return err
	}

	factory.factory.provider = provider
	return nil
}

// Load loads the given registry
func (factory *ProviderFactory) Load() (interface{}, error) {
	if factory.factory.provider == nil {
		return nil, ErrProviderNotRegistered
	}

	return factory.factory.provider, nil
}
