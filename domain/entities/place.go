package entities

import (
	"encoding/json"
	"strings"
	"time"

	"googlemaps.github.io/maps"
)

// Place ...
type Place struct {
	Address    string      `json:"address"`
	Location   maps.LatLng `json:"location"`
	Name       string      `json:"name"`
	Phone      string      `json:"phone"`
	Rating     float32     `json:"rating"`
	Schedule   string      `json:"schedule"`
	PriceLevel int         `json:"price_level"`
	Types      string      `json:"types"`
}

// Jsonable an entity that returns a JSON representation of itself
type Jsonable interface {
	ToJSON() (json.RawMessage, error)
}

// ToJSON returns a JSON representation of Place
func (p Place) ToJSON() (json.RawMessage, error) {
	return json.Marshal(p)
}

// PlaceFrom ...
func PlaceFrom(details maps.PlacesSearchResult) Place {
	place := Place{}
	place.Address = details.Vicinity // Hack because of nearbySearchRequest, should be details.FormattedAddress
	place.Location = details.Geometry.Location
	place.Name = details.Name
	place.PriceLevel = details.PriceLevel
	place.Rating = details.Rating
	// FIXME: Make this safe
	place.Schedule = stringFrom(details.OpeningHours)
	place.Types = strings.Join(details.Types, ",")

	return place
}

func stringFrom(schedule *maps.OpeningHours) string {
	for _, p := range schedule.Periods {
		if time.Now().Weekday() == p.Open.Day {
			return p.Open.Time
		}
	}
	return ""
}
