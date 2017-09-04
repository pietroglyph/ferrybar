package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	flag "github.com/ogier/pflag"
)

type config struct {
	apiKey                string
	terminal              int
	updateFrequency       int
	vesselEndpointBaseURL string
	routeWidthFactor      float64
	loopDelay             float64
}

func main() {
	var conf config
	// Get command line flags
	flag.StringVarP(&conf.apiKey, "key", "k", "", "WSDOT Traveller Information API key (provisioned at http://wsdot.wa.gov/traffic/api/)") // REquired
	flag.IntVarP(&conf.terminal, "terminal", "t", 3,                                                                                       // 3 is the Bainbridge Island Ferry terminal
		"Terminal ID to get arriving and departing boat progress for. A terminal list is available by making a GET to http://www.wsdot.wa.gov/ferries/api/terminals/rest/terminalbasics?apiaccesscode={CODE}")
	flag.IntVarP(&conf.updateFrequency, "update", "u", 60, "The frequency of GETs to the /vessellocations endpoint, in seconds")
	flag.StringVarP(&conf.vesselEndpointBaseURL, "baseurl", "b", "http://www.wsdot.wa.gov/ferries/api/vessels/rest", "The URL of the WSDOT REST API vessel data endpoint")
	flag.Float64VarP(&conf.routeWidthFactor, "width", "w", 300, "The 'width' factor of the route, this determines how far away the ferry can be to still be considered on route")
	flag.Float64VarP(&conf.loopDelay, "delay", "d", 300, "The delay after every iteration of the loop, in milliseconds")
	flag.Parse()

	// apiKey flag is required
	if conf.apiKey == "" {
		log.Fatal("Please specify an API using the -k flag")
	}

	// We only have the ferryPathPoints for the Seattle-Bainbridge route
	if conf.terminal != 3 {
		fmt.Print("Processing location data is only implemented for Terminal ID 3, continue (y/N)? ")
		stdin := bufio.NewScanner(os.Stdin)
		stdin.Scan()
		if strings.ToLower(stdin.Text()) != "y" {
			return
		}
	}

	log.Println("Flags parsed")

	// Calculate the total distnce of the ferryPathPoints (see process.go)
	calcTotalDistance()

	// Channel for data from the endpoint
	locationChan := make(chan vesselLocations)

	// Get initial data from the endpoint
	go conf.update(locationChan)

	// Go doesn't have a monotonic time source, so we'll use this for measuring duration
	lastUpdate := time.Now()

	var relaventLocationData vesselLocations
	for {
		// Clear the screen
		// We have this up here because it lets log messages from process() show up
		fmt.Println("\033c")

		// Make a request for new data to the endpoint concurrently
		if time.Since(lastUpdate).Seconds() >= float64(conf.updateFrequency) {
			// We update the progess concurrently with this HTTP request to the endpoint
			go conf.update(locationChan)
			lastUpdate = time.Now()
		}

		// Check if new data is avaliable
		// Select makes channel interactions non-blocking
		select {
		case locations := <-locationChan:
			// Check for an invalid size, which signals that *config.update had an issue
			if len(locations) == 0 {
				log.Println("Error updating location data; trying again in", conf.updateFrequency, "seconds")
				continue
			}

			// Delete all location data for vessels not departing or arriving from the target terminal
			log.Println("Location data updated sucsessfully")
			for i := 0; i < len(locations); i++ {
				if locations[i].DepartingTerminalID != conf.terminal && locations[i].ArrivingTerminalID != conf.terminal {
					locations = append(locations[:i], locations[i+1:]...)
					i--
				}
			}
			relaventLocationData = locations
		default: // For this to be non-blocking the default clause is required
			break
		}

		// Print out info about all relevent vessels
		for _, v := range relaventLocationData {
			if v.AtDock || !v.InService {
				fmt.Print(v.VesselName, " is currently docked\n")
			} else {
				fmt.Print(v.VesselName+": ", v.process(&conf), " percent to ", v.ArrivingTerminalName, "\n")
			}
			fmt.Println("Data was last changed at", v.TimeStamp.String())
		}
		fmt.Println("Last endpoint query: ", lastUpdate.String())
		time.Sleep(300 * time.Millisecond) // We can update very fast, but we don't need to
	}
}
