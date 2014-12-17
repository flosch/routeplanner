package main

import (
	"math"

	"private/routenplaner/src/src/common"
)

// https://de.wikipedia.org/wiki/WGS84
// https://en.wikipedia.org/wiki/Haversine_formula

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

type vector struct {
	x, y float64
}

func (v vector) mul(v2 vector) float64 {
	return v.x*v2.x + v.y*v2.y
}

func (v vector) sub(v2 vector) vector {
	return vector{x: v.x - v2.x, y: v.y - v2.y}
}

func (v vector) div(d float64) vector {
	return vector{v.x / d, v.y / d}
}

func (v vector) norm() float64 {
	return math.Sqrt(v.x*v.x + v.y*v.y)
}

func degree(from, via, to *common.Node) float64 {
	p1 := vector{y: from.Lat, x: from.Lon}
	p2 := vector{y: via.Lat, x: via.Lon}
	p3 := vector{y: to.Lat, x: to.Lon}

	v1 := p1.sub(p2)
	v1 = v1.div(v1.norm())

	v2 := p3.sub(p2)
	v2 = v2.div(v2.norm())

	d := rad2deg(math.Atan2(v2.y, v2.x) - math.Atan2(v1.y, v1.x))
	if d < 0 {
		d += 360
	}
	return d
}

func degree2text(d float64) string {
	switch {
	case d > 300:
		return "scharf links"
	case (d <= 300) && (d > 210):
		return "links"
	case (d >= 150) && (d <= 210):
		return "gerade aus"
	case (d < 150) && (d >= 60):
		return "rechts"
	case d < 60:
		return "scharf rechts"
	}
	panic("unreachable")
}

func lanehelper(d float64) string {
	switch {
	case d > 210:
		return "links"
	case (d >= 150) && (d <= 210):
		return ""
	case d < 150:
		return "rechts"
	}
	panic("unreachable")
}
