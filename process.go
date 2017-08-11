package main

type processedVesselLocation struct {
	pathIndex int
}

type coordinate struct {
	X, Y float32
}

// ferryPathPoints are latitude and longitude of the Seattle-Bainbridge route, sourced from Google Maps
// We interpolate between these points linearly for the time being, but a spline function
// could be introduced if we need the extra accuracy (which I don't think we need)
var ferryPathPoints = []coordinate{ // XXX: why does golint yell at me about this?
	{47.602824, -122.339544},
	{47.602869, -122.342291},
	{47.603069, -122.343750},
	{47.604169, -122.352440},
	{47.605326, -122.482388},
	{47.605471, -122.484770},
	{47.605688, -122.486615},
	{47.605934, -122.487774},
	{47.606353, -122.488804},
	{47.606643, -122.489362},
	{47.607163, -122.490134},
	{47.607757, -122.490735},
	{47.608176, -122.491014},
	{47.617116, -122.494469},
	{47.617825, -122.494855},
	{47.618331, -122.495220},
	{47.619170, -122.496078},
	{47.619546, -122.496700},
	{47.620009, -122.497602},
	{47.620197, -122.498288},
	{47.622453, -122.509274},
}

func (v *vesselLocation) process() processedVesselLocation {
	return processedVesselLocation{}
}

func (v *processedVesselLocation) progress() int {
	return 0
}
