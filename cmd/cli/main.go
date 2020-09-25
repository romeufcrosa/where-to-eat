package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/romeufcrosa/where-to-eat/gateways/google"

	domain "github.com/romeufcrosa/where-to-eat/domain/entities"
	"github.com/romeufcrosa/where-to-eat/domain/services"
	"googlemaps.github.io/maps"
)

const (
	osxCmd      = "airport"
	osxArgs     = "-s"
	linuxCmd    = "iwgetid"
	linuxArgs   = "--raw"
	windowsCmd  = "netsh"
	windowsArgs = "wlan show networks"
)

func main() {
	compoundRows := scanWiFiNetwork()

	ctx := context.TODO()

	gateway, err := domain.NewGoogleGeo()
	if err != nil {
		log.Fatal(err)
	}

	googleGateway := google.NewGoogleGateway(gateway.Client)
	locator := services.NewGeolocatorWith(&googleGateway)
	result, err := locator.FetchLocation(ctx, compoundRows)
	if err != nil {
		log.Fatal(err)
	}

	codeRequest := &maps.GeocodingRequest{
		LatLng: &result.Location,
	}

	loc, err := gateway.Client.Geocode(ctx, codeRequest)
	if err != nil {
		log.Fatal(err)
	}

	// loc[0] contains the address
	fmt.Println(loc[0].FormattedAddress)

	// TODO look for food places with that Lat and Lng
	findFood(loc[0], gateway)
}

func findFood(location maps.GeocodingResult, gw *domain.GoogleMapsAPI) {
	var places []maps.PlacesSearchResult
	var nextToken string
	currentPosition := &location.Geometry.Location

	searchRequest := &maps.NearbySearchRequest{
		Location: currentPosition,
		Type:     maps.PlaceTypeRestaurant,
		Radius:   3000,
	}

	for {
		if len(nextToken) > 0 {
			log.Println("Token found... modifying request!")
			time.Sleep(2 * time.Second)
			searchRequest = &maps.NearbySearchRequest{
				Location:  currentPosition,
				Type:      maps.PlaceTypeRestaurant,
				Radius:    3000,
				PageToken: nextToken,
				RankBy:    maps.RankByProminence,
				MinPrice:  maps.PriceLevelModerate,
			}
		}
		log.Println("Sending request to API")
		response, err := gw.Client.NearbySearch(context.Background(), searchRequest)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Found %d results in response", len(response.Results))
		places = append(places, response.Results...)

		if len(response.NextPageToken) == 0 {
			log.Println("No token found... exit loop!")
			break
		}

		log.Println("Token found... continue loop!")
		nextToken = response.NextPageToken
	}

	randomPlaceID, err := getRandomPlace(places)
	if err != nil {
		log.Fatal(err)
	}

	detailsRequest := &maps.PlaceDetailsRequest{
		PlaceID: randomPlaceID,
	}
	placeDetail, err := gw.Client.PlaceDetails(context.Background(), detailsRequest)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("------------------------")
	log.Printf("Nome: %s\n", placeDetail.Name)
	log.Printf("Morada: %s\n", placeDetail.FormattedAddress)
	log.Printf("Proximidade: %s\n", placeDetail.Vicinity)
	log.Printf("Preço: %d/5 \n", placeDetail.PriceLevel)
	log.Printf("Rating: %f\n", placeDetail.Rating)
	log.Printf("Está aberto agora? %s", FormatBool(placeDetail.OpeningHours.OpenNow))
}

func scanWiFiNetwork() []string {
	platform := runtime.GOOS
	if platform == "darwin" {
		return forOSX()
	} else if platform == "windows" {
		return forWindows()
	} else {
		// FIXME: need to return compounds rows as the other ones
		return forLinux()
	}
}

func forLinux() []string {
	cmd := exec.Command(linuxCmd, linuxArgs)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	// start the command after having set up the pipe
	if err := cmd.Start(); err != nil {
		panic(err)
	}

	var str string

	if b, err := ioutil.ReadAll(stdout); err == nil {
		str += (string(b) + "\n")
	}

	name := strings.Replace(str, "\n", "", -1)
	return []string{name}
}

func forOSX() []string {
	superRegex := `[a-f0-9]{2}:[a-f0-9]{2}:[a-f0-9]{2}:[a-f0-9]{2}:[a-f0-9]{2}:[a-f0-9]{2}\s(-\d{2})*\s*[0-9]+`
	re := regexp.MustCompile(superRegex)
	out, err := exec.Command(osxCmd, osxArgs).Output()
	if err != nil {
		log.Fatal(err)
	}

	airportRows := strings.Split(string(out), "\n")
	var compoundRows []string
	for i := range airportRows {
		match := re.FindStringSubmatch(
			airportRows[i],
		)
		if len(match) > 0 {
			rep := regexp.MustCompile(`\s*(-\d{2})\s*`)
			cleanRow := rep.ReplaceAllString(match[0], " ")
			compoundRows = append(compoundRows, cleanRow)
		}
	}

	return compoundRows
}

func forWindows() []string {
	var output string

	// TODO: Investigate if this trick is needed to "wake up" the interface
	// netsh interface set interface name="<NIC name>" admin=disabled
	// netsh interface set interface name="<NIC name>" admin=enabled
	stdout, err := exec.Command(windowsCmd, "wlan", "show", "networks", "mode=bssid").Output()
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
	output = string(stdout)

	// TODO: Grab all BSSID's with Channels
	macRegex := `[a-f0-9]{2}:[a-f0-9]{2}:[a-f0-9]{2}:[a-f0-9]{2}:[a-f0-9]{2}:[a-f0-9]{2}`
	macMatcher := regexp.MustCompile(macRegex)
	channelRegex := `Channel\s+:\s\d+`
	chanMatcher := regexp.MustCompile(channelRegex)

	macMatches := macMatcher.FindAllStringSubmatch(output, 2)
	// TODO: If matches is < 2 we might need to poke the network (see above)
	channelMatches := chanMatcher.FindAllStringSubmatch(output, 2)
	var channels = make([]string, 2)
	for index, match := range channelMatches {
		strippedMatch := strings.Replace(match[0], " ", "", -1)
		parts := strings.Split(strippedMatch, ":")
		channels[index] = parts[1]
	}

	if len(macMatches) < 2 {
		panic(err)
	}

	return []string{
		macMatches[0][0] + " " + channels[0],
		macMatches[1][0] + " " + channels[1],
	}
}

func getRandomPlace(places []maps.PlacesSearchResult) (string, error) {
	if len(places) == 0 {
		return "", errors.New("no suitable place found")
	}

	rand.Seed(time.Now().UnixNano())
	var randomPlace maps.PlacesSearchResult
	arrayPos := rand.Intn(len(places))

	if randomPlace = places[arrayPos]; randomPlace.Rating < 1 {
		places = append(places[:arrayPos], places[arrayPos+1:]...)
		return getRandomPlace(places)
	}

	return randomPlace.PlaceID, nil
}

// FormatBool transforms a boolean into a string
func FormatBool(b *bool) string {
	if ok := *b; !ok {
		return "Não"
	}

	if *b {
		return "Sim"
	}
	return "Não"
}
