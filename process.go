package main

import (
	"log"
	"math"
	"time"

	geo "github.com/kellydunn/golang-geo"
	"github.com/skelterjohn/geom"
)

// ferryPathPoints are latitude and longitude of the Seattle-Bainbridge route, sourced from Google Maps
// We interpolate between these points linearly for the time being, but a spline function
// could be introduced if we need the extra accuracy (which I don't think we need)
var ferryPathPoints = []geom.Coord{
	{X: 47.622453, Y: -122.509274},
	{X: 47.620197, Y: -122.498288},
	{X: 47.620009, Y: -122.497602},
	{X: 47.619546, Y: -122.496700},
	{X: 47.619170, Y: -122.496078},
	{X: 47.618331, Y: -122.495220},
	{X: 47.617825, Y: -122.494855},
	{X: 47.617116, Y: -122.494469},
	{X: 47.608176, Y: -122.491014},
	{X: 47.607757, Y: -122.490735},
	{X: 47.607163, Y: -122.490134},
	{X: 47.606643, Y: -122.489362},
	{X: 47.606353, Y: -122.488804},
	{X: 47.605934, Y: -122.487774},
	{X: 47.605688, Y: -122.486615},
	{X: 47.605471, Y: -122.484770},
	{X: 47.605326, Y: -122.482388},
	{X: 47.604169, Y: -122.352440},
	{X: 47.603069, Y: -122.343750},
	{X: 47.602869, Y: -122.342291},
	{X: 47.602824, Y: -122.339544},
}

var ferryPathTotalLength float64

func calcTotalDistance() {
	for i, v := range ferryPathPoints {
		if i <= 0 {
			continue
		}
		ferryPathTotalLength += math.Sqrt(math.Pow(ferryPathPoints[i-1].X-v.X, 2) + math.Pow(ferryPathPoints[i-1].Y-v.Y, 2))
	}
}

func (vesselLoc *vesselLocation) process(conf *config) float64 {
	var cumulativeDistanceTravelled float64
	var closestSegment int
	var subClosestSegmentProgress float64 // The progress of the ferry along the closest segment
	var reversed bool
	if vesselLoc.ArrivingTerminalID == conf.terminal {
		reversed = true
	}

	// Interpolate forward in time using the heading and the speed
	durationAhead := time.Now().Sub(vesselLoc.TimeStamp.Time)
	distanceAhead := durationAhead.Hours() * 1.852 // A knot is 1.852 KM/h
	interpolatedCoordinate := convertGeoPoint(geo.NewPoint(vesselLoc.Latitude, vesselLoc.Longitude).PointAtDistanceAndBearing(distanceAhead, vesselLoc.Heading))

	// Find the closest point
	closestSegment = -1
	smallestDistanceToSegment := -1.0
	for i, v := range ferryPathPoints {
		if i <= 0 {
			continue
		}
		var slope geom.Coord
		// Get the negative reciprocal of the slope of the segment we're testing against
		// so that the tester segment is perpendicular if it intersects
		slope.X = (ferryPathPoints[i-1].Minus(v).Y * conf.routeWidthFactor) * -1
		slope.Y = (ferryPathPoints[i-1].Minus(v).X * conf.routeWidthFactor) * -1
		intersectionTestSegment := geom.Segment{A: interpolatedCoordinate.Plus(slope), B: interpolatedCoordinate.Minus(slope)}
		p, ok := intersectionTestSegment.Intersection(&geom.Segment{A: ferryPathPoints[i-1], B: v})
		if ok {
			distanceToSegment := p.DistanceFrom(interpolatedCoordinate)
			if distanceToSegment < smallestDistanceToSegment || smallestDistanceToSegment == -1.0 {
				smallestDistanceToSegment = distanceToSegment
				closestSegment = i
				subClosestSegmentProgress = p.DistanceFrom(ferryPathPoints[i-1])
			}
		}
	}

	if closestSegment == -1 {
		log.Println("Ferry is not on path, consider increasing the width flag's value")
		return -1.0
	}

	for i := 1; i < closestSegment; i++ {
		cumulativeDistanceTravelled += ferryPathPoints[i].DistanceFrom(ferryPathPoints[i-1])
	}
	cumulativeDistanceTravelled += subClosestSegmentProgress

	percentage := cumulativeDistanceTravelled / ferryPathTotalLength

	if reversed {
		percentage = 1 - percentage
	}

	return percentage
}

func convertGeoPoint(pnt *geo.Point) geom.Coord {
	return geom.Coord{X: pnt.Lat(), Y: pnt.Lng()}
}

// func (p0 coordinate) distanceToLine(p1 coordinate, p2 coordinate) float64 {
// 	return math.Abs((p2.Y-p1.Y)*p0.X-(p2.X-p1.X)*p0.Y-p2.X*p1.Y-p2.Y*p1.X) / math.Sqrt(math.Pow(p2.Y-p1.Y, 2)+math.Pow(p2.X-p1.X, 2))
// }
