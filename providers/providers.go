package providers

import (
	"errors"
	"os"
	"sync"

	"github.com/romeufcrosa/where-to-eat/providers/internal"
	"googlemaps.github.io/maps"
)

// ErrProviderNotFound error sent when the provider was not found
var (
	ErrProviderNotFound          = errors.New("provider was not found")
	ErrProviderAlreadyRegistered = errors.New("provider already registered")

	env = os.Getenv("ENV")
)

// Provider the provider unique key
type Provider string

// Params parameters to pass onto providers
type Params struct {
	client *maps.Client
}

// NewParams returns a new Params
func NewParams(client *maps.Client) Params {
	return Params{client: client}
}

// ConfigurationUpdateFunc a function that checks if the configurations have changed
type ConfigurationUpdateFunc func() bool

type manufactureFactory struct {
	factory    *internal.ProviderFactory
	hasUpdates ConfigurationUpdateFunc
}

type manufacture struct {
	isConfigured bool
	mu           sync.RWMutex
	hasUpdates   ConfigurationUpdateFunc
	factories    map[Provider]manufactureFactory
	params       Params
}

func noUpdates() bool {
	return false
}

var (
	manufacturer = manufacture{
		isConfigured: false,
		factories:    make(map[Provider]manufactureFactory),
		hasUpdates:   noUpdates,
	}
)

// Configure sets the provider params to be used when initializing them
func Configure(params Params, updates ConfigurationUpdateFunc) {
	if manufacturer.isConfigured && env != "tests" {
		return
	}

	manufacturer.params = params
	manufacturer.hasUpdates = updates
	manufacturer.isConfigured = true
}

// Get gets a given provider name onto provider
func Get(name Provider) (interface{}, error) {
	manufacturer.mu.RLock()
	defer manufacturer.mu.RUnlock()

	manufacture, ok := manufacturer.factories[name]
	if !ok {
		return nil, ErrProviderNotFound
	}

	hasUpdates := manufacturer.hasUpdates
	if manufacture.hasUpdates != nil {
		hasUpdates = manufacture.hasUpdates
	}

	if hasUpdates() {
		err := manufacture.factory.Register()
		if err != nil {
			return nil, err
		}
	}

	return manufacture.factory.Load()
}

// Register registers a new provider with the given name
func Register(name Provider, registration internal.Registration, update ...ConfigurationUpdateFunc) error {
	if _, found := manufacturer.factories[name]; found {
		return ErrProviderAlreadyRegistered
	}

	factory := internal.NewProviderFactory(registration)
	if err := factory.Register(); err != nil {
		// log.WithError(err).Error(context.TODO(), "could not register provider")
		return err
	}

	var updateFunc ConfigurationUpdateFunc
	if len(update) > 0 {
		updateFunc = update[0]
	}

	manufacturer.factories[name] = manufactureFactory{
		factory:    &factory,
		hasUpdates: updateFunc,
	}
	return nil
}
