package providers

import (
	"errors"
	"sync"

	domain "github.com/romeufcrosa/where-to-eat/domain/entities"
	"github.com/romeufcrosa/where-to-eat/domain/services"
	"github.com/romeufcrosa/where-to-eat/gateways/google"
)

var (
	googleInteractor = Provider("gateways/google")
	once             = sync.Once{}
	locate           services.Locate
)

// RegisterGatewayProviders ...
func RegisterGatewayProviders() {
	Register(googleInteractor, func() (provider interface{}, err error) {
		googleGeo, err := domain.NewGoogleGeo()
		if err != nil {
			return nil, err
		}
		googleMapsGateway := google.NewGoogleGateway(googleGeo.Client)

		return googleMapsGateway, nil
	})
}

// GetLocator returns the geo locator
func GetLocator() (locator services.Locate, err error) {
	var googleProvider interface{}

	googleProvider, err = Get(googleInteractor)
	if err != nil {
		return locator, errors.New("Provider is missing")
	}

	googleGateway := googleProvider.(google.GeoGateway)

	locator = services.NewGeolocatorWith(&googleGateway)

	return
}
