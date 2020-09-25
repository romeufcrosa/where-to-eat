// Package google provides gateway logic for for interacting with Googles APIs
package google

import (
	"context"
	"log"
	"strings"
	"time"

	domain "github.com/romeufcrosa/where-to-eat/domain/entities"
	"googlemaps.github.io/maps"
)

// GeoGateway ...
type GeoGateway struct {
	client *maps.Client
}

// NewGoogleGateway ...
func NewGoogleGateway(client *maps.Client) GeoGateway {
	gg := GeoGateway{
		client: client,
	}

	return gg
}

// Geolocate ...
func (g *GeoGateway) Geolocate(ctx context.Context, accessPoints []maps.WiFiAccessPoint) (*maps.GeolocationResult, error) {
	gRequest := &maps.GeolocationRequest{
		ConsiderIP:       true,
		WiFiAccessPoints: accessPoints,
	}

	result, err := g.client.Geolocate(ctx, gRequest)
	if err != nil {
		log.Printf("Geolocate failed: %s", err.Error())
		return nil, err
	}

	return result, nil
}

// ListRestaurants ...
func (g *GeoGateway) ListRestaurants(ctx context.Context, searchRequest *maps.NearbySearchRequest) (maps.PlacesSearchResponse, error) {
	restaurants, err := g.client.NearbySearch(context.Background(), searchRequest)
	if err != nil {
		return maps.PlacesSearchResponse{}, err
	}

	return restaurants, nil
}

// PlaceDetails ...
func (g *GeoGateway) PlaceDetails(ctx context.Context, detailsRequest *maps.PlaceDetailsRequest) (domain.Place, error) {
	placeDetails, err := g.client.PlaceDetails(context.Background(), detailsRequest)
	if err != nil {
		return domain.Place{}, err
	}

	response, err := newPlaceResponse(placeDetails)
	if err != nil {
		return domain.Place{}, err
	}
	return response, nil
}

func newPlaceResponse(details maps.PlaceDetailsResult) (domain.Place, error) {
	place := domain.Place{}
	place.Address = details.FormattedAddress
	place.Location = details.Geometry.Location
	place.Name = details.Name
	place.Phone = details.FormattedPhoneNumber
	place.PriceLevel = details.PriceLevel
	place.Rating = details.Rating
	place.Schedule = stringFrom(details.OpeningHours)
	place.Types = strings.Join(details.Types, ",")

	return place, nil
}

func stringFrom(schedule *maps.OpeningHours) string {
	if schedule != nil {
		for _, p := range schedule.Periods {
			if time.Now().Weekday() == p.Open.Day {
				return p.Open.Time
			}
		}
	}
	return ""
}
