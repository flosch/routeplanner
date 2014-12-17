package main

import (
	"math"

	"private/routenplaner/src/src/common"
)

const earth_radius = 6371 // in km

func deg2rad(deg float64) float64 {
	return deg * (math.Pi / 180)
}

func rad2deg(rad float64) float64 {
	return rad * (180 / math.Pi)
}

func distance(lat1, long1, lat2, long2 float64) float64 {
	dLat := deg2rad(lat2 - lat1)
	dLong := deg2rad(long2 - long1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(deg2rad(lat1))*math.Cos(deg2rad(lat2))*
			math.Sin(dLong/2)*math.Sin(dLong/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	d := earth_radius * c
	return d
}

// Distanz zwischen zwei Knoten in km
func distanceNodes(n1, n2 *common.Node) float64 {
	return distance(n1.Lat, n1.Lon, n2.Lat, n2.Lon)
}
