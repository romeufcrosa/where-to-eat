package entities

import (
	"encoding/json"

	"googlemaps.github.io/maps"
)

// GoogleMapsAPI ...
type GoogleMapsAPI struct {
	Client *maps.Client
}

// IsConfigured returns if the service is configured
func IsConfigured() bool {
	return true
}

// NewGoogleGeo ...
func NewGoogleGeo() (*GoogleMapsAPI, error) {
	// TODO: Use WithRateLimit(qps) to ensure it doesn't go above the 50 qps limit
	client, err := maps.NewClient(maps.WithAPIKey("AIzaSyBvE7lMDfzA9hMypDNfIhGi5VtRbk8HgcU"))
	if err != nil {
		return nil, err
	}
	return &GoogleMapsAPI{
		Client: client,
	}, nil
}

// GeolocateError ...
type GeolocateError struct {
	Domain  string
	Reason  string
	Message string
}

// Geometry ...
type Geometry struct {
	Location     Location
	LocationType string `json:"location_type"`
}

// GeolocateResult ...
type GeolocateResult struct {
	FormattedAddress string `json:"formatted_address"`
	Geometry         Geometry
}

// GeolocateResponse ...
type GeolocateResponse struct {
	Results []GeolocateResult
}

// Error ...
type Error struct {
	Code    int
	Message string
	Errors  []GeolocateError
}

// JSONRequest ...
type JSONRequest struct {
	Lat      string `json:"lat"`
	Lng      string `json:"lng"`
	Distance string `json:"distance"`
	Pricing  string `json:"pricing"`
}

// Location ...
type Location struct {
	Lat float64
	Lng float64
}

// SearchRequest ...
type SearchRequest struct {
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	Distance uint    `json:"distance"`
	Pricing  int     `json:"pricing"`
}

// Point ...
type Point struct {
	Lat          float64
	Lng          float64
	Address      string
	LocationType string
}

// NewPoint ...
func NewPoint(lat, lng float64) *Point {
	return &Point{Lat: lat, Lng: lng}
}

// NewFromJSON loads a Location from the request Body
func NewFromJSON(bodyBytes []byte) (sr SearchRequest, err error) {
	if err = json.Unmarshal(bodyBytes, &sr); err != nil {
		return sr, err
	}

	return sr, nil
}
