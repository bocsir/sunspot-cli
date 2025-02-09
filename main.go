// add option to change location string
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
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
		Date       string `json:"date"`
		Sunrise    string `json:"sunrise"`
		Sunset     string `json:"sunset"`
		FirstLight string `json:"first_light"`
		LastLight  string `json:"last_light"`
		Dawn       string `json:"dawn"`
		Dusk       string `json:"dusk"`
		SolarNoon  string `json:"solar_noon"`
		GoldenHour string `json:"golden_hour"`
		DayLength  string `json:"day_length"`
		Timezone   string `json:"timezone"`
		UTCOffset  int    `json:"utc_offset"`
	} `json:"results"`
	Status string `json:"status"`
}

type SunTimes struct {
	sunrise, sunset, tzOffset int
}

func main() {
	var wg sync.WaitGroup
	var location Location
	var sunTimes SunTimes
	var apiErr error

	// Get location
	storedLocation, err := getCoords()
	if err != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			loc, err := stringToCoords()
			if err != nil {
				apiErr = err
				return
			}
			location = loc
		}()

		wg.Wait()
		if apiErr != nil {
			fmt.Printf("Error getting coordinates:", apiErr)
		}
		storeCoords(location)
	} else {
		location = storedLocation
	}

	// Get sun times
	wg.Add(1)
	go func() {
		defer wg.Done()
		times, err := getSunriseAndSetMin(location)
		if err != nil {
			apiErr = err
			return
		}
		sunTimes = times
	}()

	wg.Wait()
	if apiErr != nil {
		log.Fatal("Error getting sun times:", apiErr)
	}

	angle := getAngle(sunTimes)
	fileName := getFileName(angle)
	printFileStream("./ascii-art/" + fileName)
	fmt.Println()
}

func getFileName(angle float64) string {
	fileNames := [16]string{"0.txt", "22.txt", "45.txt", "67.txt", "90.txt", "112.txt", "135.txt", "156.txt", "180.txt", "202.txt", "225.txt", "247.txt", "270.txt", "292.txt", "315.txt", "337.txt"}
	closestIncrement := int(math.Round(float64(angle) / 22.5))
	return fileNames[closestIncrement]
}

func printFileStream(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(os.Stdout, file)
	return err
}

// get user input for lat, lon
func stringToCoords() (Location, error) {
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

	return location, nil
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
	//check format
	if !isValidCoordinates(string(coords)) {
		var EmptyLoc = Location{0.0, 0.0}
		return EmptyLoc, errors.New("coords not found")
	}
	//format
	coordsSeparated := strings.Split(string(coords), ",")
	latFloat, _ := strconv.ParseFloat(coordsSeparated[0], 64)
	lonFloat, _ := strconv.ParseFloat(coordsSeparated[1], 64)
	var Location = Location{latFloat, lonFloat}

	return Location, nil
}

// regex for coordinates
func isValidCoordinates(coords string) bool {
	var coordinateRegex = regexp.MustCompile(
		`^` + // Start of string
			`([-+]?(?:[0-8][0-9](?:\.[0-9]{1,8})?|90(?:\.0{1,8})?))` + // Latitude: -90 to 90 with optional decimals
			`,` + // Comma separator
			`([-+]?(?:(?:1[0-7][0-9]|[0-9]{1,2})(?:\.[0-9]{1,8})?|180(?:\.0{1,8})?))` + // Longitude: -180 to 180 with optional decimals
			`$`, // End of string
	)

	return coordinateRegex.MatchString(coords)
}

// needs to take lat and lon as input
func getSunriseAndSetMin(Location Location) (SunTimes, error) {
	lat := Location.Lat
	lon := Location.Lon

	//call API for times
	url := fmt.Sprintf("https://api.sunrisesunset.io/json?lat=%f&lng=%f&time_format=24", lat, lon)
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

	var sunData SunData
	err = json.Unmarshal([]byte(body), &sunData)
	if err != nil {
		fmt.Print(err.Error())
	}

	riseStr, err2 := addAMPM(sunData.Results.Sunrise)
	if err2 != nil {
		fmt.Println("Error: ", err)
	}
	setStr, err := addAMPM(sunData.Results.Sunset)
	if err != nil {
		fmt.Println("Error: ", err)
	}

	fmt.Println("Sun Rise: ", riseStr)
	fmt.Println("Sun Set:  ", setStr)

	//format results
	sunrise := getMinutes(sunData.Results.Sunrise)
	sunset := getMinutes(sunData.Results.Sunset)
	var Times = SunTimes{sunrise, sunset, sunData.Results.UTCOffset}
	return Times, nil
}

// get angle to create ASCII art with
func getAngle(Times SunTimes) float64 {
	loc := time.FixedZone("Custom", Times.tzOffset*60)
	currentTime := time.Now().In(loc)
	c := currentTime.Minute() + currentTime.Hour()*60
	r := Times.sunrise
	s := Times.sunset

	var angle float64

	if r <= c && c <= s { // day
		if s == r {
			return 90 // or whatever default angle makes sense for this edge case
		}
		angle = float64(c-r) / float64(s-r) * 180
	} else { // night
		totalMinutes := 24 * 60
		if c < r { // before sunrise
			nightProgress := float64(c+(totalMinutes-s)) / float64(totalMinutes-(s-r))
			angle = 180 + (nightProgress * 180)
		} else { // after sunset
			nightProgress := float64(c-s) / float64(totalMinutes-(s-r))
			angle = 180 + (nightProgress * 180)
		}
	}

	// Ensure angle stays within 0-360 range
	return math.Mod(angle+360, 360)
}

// parse to minutes
func getMinutes(timeStr string) int {
	t, err := time.Parse("15:04:05", timeStr)
	if err != nil {
		fmt.Println(err.Error())
	}
	return t.Hour()*60 + t.Minute()
}

func addAMPM(militaryTime string) (string, error) {
	parts := strings.Split(militaryTime, ":")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid format: expected HH:MM:SS, got %s", militaryTime)
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil || hours < 0 || hours > 23 {
		return "", fmt.Errorf("invalid hours")
	}

	period := "AM"
	if hours >= 12 {
		period = "PM"
		if hours > 12 {
			hours -= 12
		}
	}
	if hours == 0 {
		hours = 12
	}

	return fmt.Sprintf("%02d:%s %s", hours, parts[1], period), nil
}
