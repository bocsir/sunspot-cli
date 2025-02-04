package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type SunTimes struct {
	sunrise, sunset int
}

type Location struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

type SunData struct {
	Results struct {
		sunrise string `json:"sunrise"`
		sunset  string `json:"sunset"`
	} `json:"results"`
}

func main() {
	Location := getLatLon()
	fmt.Println(Location)
	// SunTimes := getSunriseAndSetMin(Location)
	// angle := getAngle(SunTimes)
	// fmt.Println(angle)
}

// how to store these in client???!!
// get user input for lat, lon
func getLatLon() Location {
	//get key
	err := godotenv.Load("local.env")
	if err != nil {
		fmt.Print(err.Error())
	}
	key := os.Getenv("GEOCODEKEY")

	//get user location input
	fmt.Printf("location: ")
	var locationInput string
	fmt.Scan(&locationInput)

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

	var location Location
	err = json.Unmarshal(body, &location)
	if err != nil {
		fmt.Print(err.Error())
	}
	return location
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

	var data SunData
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Print(err.Error())
	}
	//format results
	sunrise := getMinutes(data.Results.sunrise)
	sunset := getMinutes(data.Results.sunset)
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
	t, err := time.Parse("1:02:03 PM", timeStr)
	if err != nil {
		fmt.Println(err.Error())
	}
	return t.Hour()*60 + t.Minute()
}
