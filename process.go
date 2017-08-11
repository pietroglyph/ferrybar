package main

import (
	"math"
	"time"

	geo "github.com/kellydunn/golang-geo"
)

type processedVesselLocation struct {
	pathIndex        int
	smallestDistance float64
}

type coordinate struct {
	X, Y float64
}

// ferryPathPoints are latitude and longitude of the Seattle-Bainbridge route, sourced from Google Maps
// We interpolate between these points linearly for the time being, but a spline function
// could be introduced if we need the extra accuracy (which I don't think we need)
var ferryPathPoints = []coordinate{ // XXX: why does golint yell at me about this?
	{47.622453, -122.509274},
	{47.620197, -122.498288},
	{47.620009, -122.497602},
	{47.619546, -122.496700},
	{47.619170, -122.496078},
	{47.618331, -122.495220},
	{47.617825, -122.494855},
	{47.617116, -122.494469},
	{47.608176, -122.491014},
	{47.607757, -122.490735},
	{47.607163, -122.490134},
	{47.606643, -122.489362},
	{47.606353, -122.488804},
	{47.605934, -122.487774},
	{47.605688, -122.486615},
	{47.605471, -122.484770},
	{47.605326, -122.482388},
	{47.604169, -122.352440},
	{47.603069, -122.343750},
	{47.602869, -122.342291},
	{47.602824, -122.339544},
}

var ferryPathTotalDistance float64

func calcTotalDistance() {
	for i, v := range ferryPathPoints {
		if i <= 0 {
			continue
		}
		ferryPathTotalDistance += math.Sqrt(math.Pow(ferryPathPoints[i-1].X-v.X, 2) + math.Pow(ferryPathPoints[i-1].Y-v.Y, 2))
	}
}

func (vesselLoc *vesselLocation) process() float64 {
	prossVesselLoc := processedVesselLocation{}

	// Set up variables
	prossVesselLoc.smallestDistance = -1
	var cumulativeLineDistance float64
	var currentLineDistance float64

	// Compute the distance from the latitude and longitude to each path point, and find the closest one
	for i, v := range ferryPathPoints {
		if i <= 0 { // Don't compute this for the first point, because a single point can't form a line
			continue
		}

		// Interpolate forward in time using the heading and the speed
		durationAhead := time.Now().Sub(vesselLoc.TimeStamp.Time)
		distanceAhead := durationAhead.Hours() * 1.852 // A knot is 1.852 KM/h
		transposedPoint := geo.NewPoint(vesselLoc.Latitude, vesselLoc.Longitude).PointAtDistanceAndBearing(distanceAhead, vesselLoc.Heading)

		// Set up variables for calculating the distance from the current lat and long to this ferry path line
		P1 := v
		P2 := ferryPathPoints[i-1]
		X0 := transposedPoint.Lat()
		Y0 := transposedPoint.Lng()
		d := math.Abs((P2.Y-P1.Y)*X0-(P2.X-P1.X)*Y0+P2.X*P1.Y-P2.Y*P1.X) / math.Sqrt(math.Pow(P2.Y-P1.Y, 2)+math.Pow(P2.X-P1.X, 2))

		// Calculate distances
		if d < prossVesselLoc.smallestDistance || prossVesselLoc.smallestDistance == -1 {
			// Get how far along the current line we are (this equation is totally wrong unless the current line for i is actually the closest to the ferry)
			currentLineDistance = math.Sqrt(math.Pow(prossVesselLoc.smallestDistance, 2) - math.Pow(d, 2))

			// Add the length of the last line to the cumulativeLineDistance, if there were any lines before us
			if i >= 2 {
				cumulativeLineDistance += math.Sqrt(math.Pow(ferryPathPoints[i-2].X-ferryPathPoints[i-1].X, 2) + math.Pow(ferryPathPoints[i-2].Y-ferryPathPoints[i-1].Y, 2))
			}

			prossVesselLoc.smallestDistance = d
			prossVesselLoc.pathIndex = i
		}
	}
	cumulativeLineDistance += currentLineDistance
	percentage := cumulativeLineDistance / ferryPathTotalDistance

	return percentage
}
