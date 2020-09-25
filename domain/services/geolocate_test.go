package services

import (
	"testing"

	"github.com/stretchr/testify/mock"

	mocks "github.com/romeufcrosa/where-to-eat/tests/mocks/domain/services"
	maps "googlemaps.github.io/maps"

	. "github.com/onsi/gomega"
)

func TestFormatBool(t *testing.T) {
	testCases := []struct {
		desc     string
		boolean  bool
		expected string
	}{
		{
			desc:     "Format false bool",
			boolean:  false,
			expected: "Não",
		},
		{
			desc:     "Format true bool",
			boolean:  true,
			expected: "Sim",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			RegisterTestingT(t)
			resp := FormatBool(tC.boolean)
			Expect(resp).To(Equal(tC.expected), "Should match expected conversion")
		})
	}
}

func TestFetchLocation(t *testing.T) {
	testCases := []struct {
		desc        string
		locatorMock func() GeoLocator
	}{
		{
			desc: "Fetch with just 1 AP",
			locatorMock: func() GeoLocator {
				locator := &mocks.GeoLocator{}

				aPoints := []maps.WiFiAccessPoint{
					{
						MACAddress: "68-EC-C5-C7-D9-F4",
						Channel:    13,
					},
				}

				locator.On("Geolocate", mock.Anything, aPoints).Return()

				return locator
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			RegisterTestingT(t)
		})
	}
}
