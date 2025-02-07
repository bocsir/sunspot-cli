//add option to change location string
//implement ascii art somehow

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type LocationData []struct {
	PlaceID     int      `json:"place_id"`
	Licence     string   `json:"licence"`
	OsmType     string   `json:"osm_type"`
	OsmID       int      `json:"osm_id"`
	Boundingbox []string `json:"boundingbox"`
	Lat         string   `json:"lat"`
	Lon         string   `json:"lon"`
	DisplayName string   `json:"display_name"`
	Class       string   `json:"class"`
	Type        string   `json:"type"`
	Importance  float64  `json:"importance"`
}

type Location struct {
	Lat, Lon float64
}

type SunData struct {
	Results struct {
		Sunrise                   string `json:"sunrise"`
		Sunset                    string `json:"sunset"`
		SolarNoon                 string `json:"solar_noon"`
		DayLength                 string `json:"day_length"`
		CivilTwilightBegin        string `json:"civil_twilight_begin"`
		CivilTwilightEnd          string `json:"civil_twilight_end"`
		NauticalTwilightBegin     string `json:"nautical_twilight_begin"`
		NauticalTwilightEnd       string `json:"nautical_twilight_end"`
		AstronomicalTwilightBegin string `json:"astronomical_twilight_begin"`
		AstronomicalTwilightEnd   string `json:"astronomical_twilight_end"`
	} `json:"results"`
	Status string `json:"status"`
	Tzid   string `json:"tzid"`
}

type SunTimes struct {
	sunrise, sunset int
}

func main() {
	//if !location
	var Location Location

	//this is where i need to pull location from text file
	StoredLocation, err := getCoords()
	//if no location, ask user for one and store coords
	if err != nil {
		Location = stringToCoords()
		storeCoords(Location)
	}

	SunTimes := getSunriseAndSetMin(StoredLocation)
	angle := getAngle(SunTimes)
	fmt.Println(angle)
}

// get user input for lat, lon
func stringToCoords() Location {
	//get key
	err := godotenv.Load("local.env")
	if err != nil {
		fmt.Print(err.Error())
	}
	key := os.Getenv("GEOCODEKEY")

	//get and format user location input
	fmt.Printf("human readable location: ")
	reader := bufio.NewReader(os.Stdin)
	locationInput, _ := reader.ReadString('\n')
	locationInput = strings.TrimSpace(locationInput)
	locationInput = strings.Replace(locationInput, " ", "+", -1)
	locationInput = strings.Replace(locationInput, ",", "+", -1)

	//get lat, lon from user input
	url := fmt.Sprintf("https://geocode.maps.co/search?q=%s&api_key=%s", locationInput, key)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Print(err.Error())
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Print(err.Error())
	}
	defer res.Body.Close()

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		fmt.Print(err.Error())
	}

	fmt.Println(string(body))

	var LocationData LocationData
	err = json.Unmarshal(body, &LocationData)
	if err != nil {
		fmt.Print(err.Error())
	}

	//add functionality to ask which is the correct of list of locations after user input
	//currently just using the first location
	firstLocation := LocationData[0]
	latStr := firstLocation.Lat
	lonStr := firstLocation.Lon

	latFloat, _ := strconv.ParseFloat(latStr, 64)
	lonFloat, _ := strconv.ParseFloat(lonStr, 64)

	var location = Location{latFloat, lonFloat}

	return location
}

// store coords in text file
func storeCoords(Location Location) {
	//comma separated lat,lon
	locStr := fmt.Sprintf("%f,%f", Location.Lat, Location.Lon)

	//create and write to file
	f, err := os.Create("coords.txt")
	if err != nil {
		fmt.Print(err.Error())
	}
	defer f.Close()
	_, err2 := f.WriteString(locStr)
	if err2 != nil {
		fmt.Print(err.Error())
	}
}

// get coords from text file
func getCoords() (Location, error) {
	//read from file
	coords, err := os.ReadFile("coords.txt")
	if err != nil {
		fmt.Print(err.Error())
	}
	if len(coords) < 1 {
		var EmptyLoc = Location{0.0, 0.0}
		return EmptyLoc, errors.New("Coords not found")
	}
	//format
	coordsSeparated := strings.Split(string(coords), ",")
	latFloat, _ := strconv.ParseFloat(coordsSeparated[0], 64)
	lonFloat, _ := strconv.ParseFloat(coordsSeparated[1], 64)
	var Location = Location{latFloat, lonFloat}

	return Location, nil
}

// needs to take lat and lon as input
func getSunriseAndSetMin(Location Location) SunTimes {
	//test values:
	// lat := 36.7201600
	// lon := -4.4203400

	lat := Location.Lat
	lon := Location.Lon

	//call API for times
	url := fmt.Sprintf("https://api.sunrise-sunset.org/json?lat=%f&lng=%f", lat, lon)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Print(err.Error())
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Print(err.Error())
	}

	defer res.Body.Close()

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		fmt.Print(err.Error())
	}
	fmt.Println(string(body))

	var sunData SunData
	err = json.Unmarshal([]byte(body), &sunData)
	if err != nil {
		fmt.Print(err.Error())
	}

	fmt.Println(sunData)
	//format results
	sunrise := getMinutes(sunData.Results.Sunrise)
	sunset := getMinutes(sunData.Results.Sunset)
	var Times = SunTimes{sunrise, sunset}
	return Times
}

// get angle to create ASCII art with
func getAngle(Times SunTimes) float64 {
	c := time.Now().Minute() + time.Now().Hour()*60

	//test values:
	// c := 600
	// r := 450
	// s := 1065

	r := Times.sunrise
	s := Times.sunset

	var angle float64
	if r < c && c < s { //day
		angle = float64(c-r) / float64(s-r) * 180
	} else { //night
		angle = float64(c+(1440-s))/float64(1440-(s-r))*180 + 180
	}
	return angle
}

// parse to minutes
func getMinutes(timeStr string) int {
	t, err := time.Parse("3:04:05 PM", timeStr)
	if err != nil {
		fmt.Println(err.Error())
	}
	return t.Hour()*60 + t.Minute()
}
