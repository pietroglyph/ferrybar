package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type vesselLocations []vesselLocation

type vesselLocation struct {
	VesselID                int
	VesselName              string
	Mmsi                    int `json:",omitempty"`
	DepartingTerminalID     int
	DepartingTerminalName   string
	DepartingTerminalAbbrev string
	ArrivingTerminalID      int     `json:",omitempty"`
	ArrivingTerminalName    string  `json:",omitempty"`
	ArrivingTerminalAbbrev  string  `json:",omitempty"`
	Latitude                float64 // float64 is our double analouge
	Longitude               float64
	Speed                   float64
	Heading                 float64
	InService               bool
	AtDock                  bool
	LeftDock                Time   `json:"LeftDock, string, omitempty"`
	Eta                     Time   `json:"Eta, string, omitempty"`
	EtaBasis                string `json:",omitempty"`
	ScheduledDeparture      Time   `json:"ScheduledDeparture, string, omitempty"`
	OpRouteAbbrev           []string
	VesselPositionNum       int `json:",omitempty"`
	SortSeq                 int
	ManagedBy               int  // Enum, 1 for WSF, and 2 for KCM
	TimeStamp               Time `json:"TimeStamp, string"`
}

func (conf *config) update(c chan vesselLocation) {
	var err error

	locationData := vesselLocation{}
	client := &http.Client{}

	// Set up the request
	req, err := http.NewRequest("GET", conf.vesselEndpointBaseURL+"/vessellocations?apiaccesscode="+conf.apiKey, nil)
	if err != nil {
		log.Println("Couldn't assemble a http.Request type: ", err.Error())
		c <- locationData
		return
	}
	req.Header.Set("Accept", "application/json")

	// Actually make the request
	resp, err := client.Do(req)
	if err != nil {
		log.Println("HTTP request failed: ", err.Error())
		c <- locationData
		return
	}
	defer resp.Body.Close()

	// Make sure that the response is OK
	if resp.StatusCode != 200 {
		log.Println("Endpoint returned a non-OK status code of ", resp.StatusCode, ". Your API key could be invalid.")
		bod, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(bod), "\n", req.URL.String())
		c <- locationData
		return
	}

	// Read the response body into a byte slice
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Couldn't read response body: ", err.Error())
		c <- locationData
		return
	}

	var locationArray vesselLocations
	err = json.Unmarshal(body, &locationArray)
	if err != nil {
		log.Println("Couldn't unmarshal response JSON: ", err.Error())
		c <- vesselLocation{}
		return
	}

	for _, v := range locationArray {
		fmt.Println("Bar")
		if v.InService && v.DepartingTerminalID == conf.targetTerminal {
			c <- v
		}
	}
}

// The WSF endpoint returns non RFC 3339 formatted time, so we'll have to deal with it ourselves

// Time impliments a coustom unmarshaller
type Time struct {
	time.Time
}

// UnmarshalJSON unmarshalls specially formatted time
func (t *Time) UnmarshalJSON(b []byte) error {
	// Return on "null" data like the standard library doesn
	if string(b) == "null" {
		return nil
	}

	// First get rid of the \/Date() portion
	truncated := strings.TrimSuffix(strings.TrimPrefix(string(b), "\"\\/Date("), ")\\/\"") // This is because we have a capture group, and we need to find a submatch

	// Then separate the epoch time from the time zone
	timeSplit := strings.Split(truncated, "-")

	// Make sure that there are the correct number of dashes
	if len(timeSplit) > 2 {
		return errors.New("ASP.NET time submatch had too many dashes")
	}

	// Parse the unix time into an int
	i, err := strconv.ParseInt(timeSplit[0], 10, 64)
	if err != nil {
		return err
	}

	parsedTime := time.Unix(0, i*1000000) // i is in milleseconds so we need to convert to nano, hence the multiplication

	*t = Time{parsedTime}
	return nil
}
