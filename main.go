package main

import (
	"fmt"
	"log"
	"time"

	flag "github.com/ogier/pflag"
)

type config struct {
	apiKey                string
	departingTerminal     int
	updateFrequency       int
	vesselEndpointBaseURL string
	routeWidthFactor      float64
}

func main() {
	var conf config
	flag.StringVarP(&conf.apiKey, "key", "k", "", "WSDOT Traveller Information API key (provisioned at http://wsdot.wa.gov/traffic/api/)") // Required flag
	flag.IntVarP(&conf.departingTerminal, "terminals", "t", 3,
		"Departing Terminal ID to get boat progress for, terminal list avaliable by making a GET to http://www.wsdot.wa.gov/ferries/api/terminals/rest/terminalbasics?apiaccesscode={CODE}") // 3 is the Bainbridge Island Ferry terminal
	flag.IntVarP(&conf.updateFrequency, "update", "u", 60, "The frequency of GETs to the /vessellocations endpoint, in seconds")
	flag.StringVarP(&conf.vesselEndpointBaseURL, "baseurl", "b", "http://www.wsdot.wa.gov/ferries/api/vessels/rest", "The URL of the WSDOT REST API vessel data endpoint")
	flag.Float64VarP(&conf.routeWidthFactor, "width", "w", 300, "The 'width' factor of the route, this determines how far away the ferry can be to still be considered on route")
	flag.Parse()

	// apiKey flag is required
	if conf.apiKey == "" {
		log.Fatal("Please specify an API using the -k flag.")
	}

	// We only have the ferryPathPoints for the Seattle-Bainbridge route
	if conf.departingTerminal != 3 {
		log.Fatal("Processing location data is only implemented for Departing Terminal ID 3")
	}

	log.Println("Flags parsed.")

	// Calculate the total distnce of the ferryPathPoints (see process.go)
	calcTotalDistance()

	// Channel for data from the endpoint
	locationChan := make(chan vesselLocation)

	// Get initial data from the endpoint
	go conf.update(locationChan)

	// Go doesn't have a monotonic time source, so we'll use this for measuring duration
	lastUpdate := time.Now()

	var locData vesselLocation
	for {
		fmt.Println("\033c")
		if time.Since(lastUpdate).Seconds() >= float64(conf.updateFrequency) {
			// We update the progess concurrently with this HTTP request to the endpoint
			go conf.update(locationChan)
			lastUpdate = time.Now()
		}
		// Select makes channel interactions non-blocking
		select {
		case locData = <-locationChan:
			if locData.TimeStamp == (Time{time.Time{}}) {
				// Mandatory field is empty, something went wrong
				log.Println("Recieved empty location data; trying again in ", conf.updateFrequency, " seconds")
				continue
			} else {
				log.Println("Location data updated sucsessfully")
			}
		default: // For this to be non-blocking the default clause is required
			break
		}
		if !locData.AtDock && locData.InService {
			// locData.process(&conf)
			fmt.Println(locData.process(&conf), "\n",
				"Last endpoint query: ", lastUpdate.String(), "\n",
				"Last endpoint change: ", locData.TimeStamp.String()) // Clear the screen and print the current progress
		} else if locData.AtDock {
			fmt.Println(locData.VesselName, "is currently docked.")
		} else {
			fmt.Println("Couldn't find a ferry departing from terminal", conf.departingTerminal)
		}
		time.Sleep(300 * time.Millisecond) // We can update very fast, but we don't need to
	}
}
