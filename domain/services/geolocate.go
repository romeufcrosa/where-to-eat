package services

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	domain "github.com/romeufcrosa/where-to-eat/domain/entities"
	maps "googlemaps.github.io/maps"
)

// GeoLocator ...
type GeoLocator interface {
	Geolocate(ctx context.Context, accessPoints []maps.WiFiAccessPoint) (*maps.GeolocationResult, error)
	ListRestaurants(ctx context.Context, searchRequest *maps.NearbySearchRequest) (maps.PlacesSearchResponse, error)
	PlaceDetails(ctx context.Context, detailsRequest *maps.PlaceDetailsRequest) (domain.Place, error)
}

// Wifi ...
type Wifi struct {
	BSSID   string
	Channel int
}

// Locate ...
type Locate struct {
	google GeoLocator
}

// NewGeolocatorWith ...
func NewGeolocatorWith(google GeoLocator) Locate {
	return Locate{
		google: google,
	}
}

// FetchLocation ...
func (l Locate) FetchLocation(ctx context.Context, airportRows []string) (*maps.GeolocationResult, error) {
	m := make(map[int]Wifi)

	if (len(airportRows)) < 2 {
		return nil, errors.New("no APs available")
	}

	for i := 0; i < 2; i++ {
		row := strings.Split(airportRows[i], " ")
		if strings.TrimSpace(strings.Join(row, "")) == "" {
			return nil, errors.New("ap row incomplete")
		}
		ch, err := strconv.Atoi(row[1])
		if err != nil {
			log.Printf("conversion error: %s", err.Error())
			return nil, err
		}
		m[i] = Wifi{
			BSSID:   row[0],
			Channel: ch,
		}
	}

	accessPoints := []maps.WiFiAccessPoint{
		{
			Channel:    m[0].Channel,
			MACAddress: m[0].BSSID,
		},
		{
			Channel:    m[1].Channel,
			MACAddress: m[1].BSSID,
		},
	}

	return l.google.Geolocate(ctx, accessPoints)
}

// FetchRestaurant ...
func (l Locate) FetchRestaurant(ctx context.Context, req domain.SearchRequest) (domain.Place, error) {
	currentPosition := &maps.LatLng{
		Lat: req.Lat,
		Lng: req.Lng,
	}

	searchRequest := &maps.NearbySearchRequest{
		Location: currentPosition,
		Radius:   req.Distance,
		Type:     maps.PlaceTypeRestaurant,
	}

	log.Println("Sending request to API")
	response, err := l.google.ListRestaurants(ctx, searchRequest)
	if err != nil {
		log.Printf("error from google API: %s", err.Error())
		return domain.Place{}, err
	}
	log.Printf("Found %d results in response", len(response.Results))

	randomPlace, err := getRandomPlace(response.Results)
	if err != nil {
		log.Printf("Could not get random place, reason: %s", err.Error())
		return domain.Place{}, err
	}

	placeDetail := domain.PlaceFrom(randomPlace)

	return placeDetail, nil
}

func getRandomPlace(places []maps.PlacesSearchResult) (maps.PlacesSearchResult, error) {
	if len(places) == 0 {
		return maps.PlacesSearchResult{}, errors.New("no suitable place found")
	}

	rand.Seed(time.Now().UnixNano())
	var randomPlace maps.PlacesSearchResult
	arrayPos := rand.Intn(len(places))

	if randomPlace = places[arrayPos]; randomPlace.Rating < 1 {
		places = append(places[:arrayPos], places[arrayPos+1:]...)
		return getRandomPlace(places)
	}

	return randomPlace, nil
}

// FormatBool transforms a boolean into a string
func FormatBool(b bool) string {
	if b {
		return "Sim"
	}
	return "Não"
}
