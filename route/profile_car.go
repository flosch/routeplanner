package main

import (
	"fmt"

	"private/routenplaner/src/src/common"
)

func profile_car(hints *HintMgr, prev, via, next *common.Node, prev_way, next_way *common.Way, request *RouteRequest) (int, bool) {
	// Prev ----[prev_way]---- Via ----[next_way]---- Next

	// TODO: Nicht auf einer zweifahrbahnigen Straße wenden!

	// Abbiegebeschränkungen
	if prev_way != nil {
		if !prev_way.CheckTurnRestrictions(request.osm, next_way, via) {
			// Gegen eine Abbiegebeschränkung verstoßen, nicht erlauben
			return 0, false
		}
	}

	speed := float64(next_way.Maxspeed) // in km/h
	var penalties int

	// Wege, die für Autos nicht vorgesehen sind, nicht nutzen
	if !next_way.CanBeUsedByCars() {
		hints.AddWayHint(next, "Diese Straße ist nicht von Autos befahrbar, konnte jedoch nicht umfahren werden.")
		speed = 0.01
		//return 0, false
	}

	// Barrieren umfahren und ggf. darauf hinweisen
	if _, has_barrier := next.Tags["barrier"]; has_barrier {
		hints.AddNodeHint(next, "Auf dem Weg befindet sich eine Barriere, die sich leider nicht umfahren lässt.")
		speed = 1
	}

	// Privatstraßen nur mit 5 km/h
	if next_way.Access == "private" {
		hints.AddWayHint(via, "Achtung Privatstraße! Fahren Sie vorsichtig.")
		speed = 5
	}

	// Wenn noch keine Geschwindigkeit gegeben ist, Standardgeschwindigkeit ermitteln
	if speed == 0 {
		// Gesetzliche maximale Geschwindigkeit ermitteln
		switch {
		case next_way.IsMotorway():
			speed = 120
		case next_way.IsTrunk():
			speed = 100
		case next_way.IsPrimary():
			speed = 70
		case next_way.IsSecondary():
			speed = 60
		case next_way.IsTertiary():
			speed = 50
		case next_way.Highway == "service":
			speed = 10
			return 0, false
		case next_way.Highway == "living_street":
			hints.AddWayHint(via, "Achtung Wohnstraße! Fahren Sie vorsichtig.")
			speed = 10
		default:
			speed = 30
		}
	}

	// Fahren nur 95% der Geschwindigkeit
	speed *= 0.95

	// Bei Ampeln warten
	// TODO: Problem hier ist an großen Kreuzungen, dass Ampeln mehrfach gezählt werden (statt nur einmal)
	/*
		if next.Tags["highway"] == "traffic_signals" {
			penalties += TRAFFIC_LIGHT_WAIT_TIME
		}
	*/

	/*
		if prev != nil && !prev_way.IsSpeedway() && !next_way.IsSpeedway() && direction(prev, via, next) != DIRECTION_HOLD {
			penalties += 3
		}
	*/

	//fmt.Printf("%s %.2f (+%d)\n", next_way.Fullname(), speed, penalties)

	// TODO: Vorfahrtsstraße bevorzugen
	// TODO: Vorfahrt gewähren-Straßen benachteiligen
	// TODO: Ampeln an großen Kreuzungen nur einmal zählen

	distance := distanceNodes(via, next)                      // in km
	traveltime := (distance/speed)*60*60 + float64(penalties) // in sekunden

	if via.ID == 2149039562 {
		fmt.Printf(" [ auf %d, von = %d, distance=%.3f, traveltime=%.3f, speed=%.3f, penalties=%d, way_name=%s ] ", next.ID, prev.ID, distance, traveltime, speed, penalties, next_way.Fullname())
	}

	return int(traveltime * 1000), true // in millisekunden zurückgeben
}
