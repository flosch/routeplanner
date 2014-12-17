package main

import (
	"private/routenplaner/src/src/common"
)

func profile_bike(hints *HintMgr, prev, via, next *common.Node, prev_way, next_way *common.Way, request *RouteRequest) (int, bool) {
	// Prev ----[prev_way]---- Via ----[next_way]---- Next

	// Crossing als Nodes berücksichtigen (Übergänge für Fahrräder/Fußgänger), haben kein highway-Tag!

	speed := float64(0) // in km/h
	penalties := 0      // in seconds
	hasBicycleSupport := next_way.HasBicycleSupport()

	// Autobahnen meiden
	if next_way.IsMotorway() {
		return 0, false
	}

	// Schnellstraßen meiden
	if next_way.IsTrunk() && !hasBicycleSupport {
		return 0, false
	}

	if next_way.Highway == "bridleway" ||
		next_way.Highway == "raceway" ||
		next_way.Highway == "bus_guideway" {
		return 0, false
	}

	// Abbiegebeschränkungen prüfen
	if prev_way != nil {
		// Es gibt einen Vorgänger, dann können wir auch Abbiegebeschränkungen testen
		// Es gibt keinen, wenn wir die Route gerade als ersten Node gestartet haben

		if !prev_way.CheckTurnRestrictions(request.osm, next_way, via) {
			// Gegen eine Abbiegebeschränkung verstoßen, nicht erlauben
			return 0, false
		}
	}

	distance := distanceNodes(via, next) // in km

	// Barrieren umfahren und ggf. darauf hinweisen
	if barrier, has_barrier := next.Tags["barrier"]; has_barrier {
		if next.Access == "private" && barrier != "bollard" {
			hints.AddNodeHint(next, "Auf dem Weg befindet sich eine Barriere, die sich leider nicht umfahren lässt. Unter Umständen lässt sie sich nicht öffnen.")
			speed = 1
			penalties += 60 // 1 minute strafe
		} else {
			if barrier != "bollard" {
				// Alle anderen Barrier bekommen eine Geschwindigkeit von 5 km/h
				hints.AddNodeHint(next, "Auf dem Weg befindet sich eine Barriere.")
				//speed = 5
				penalties += 15 // 30 s strafe für umgehen der barriere
			}
		}
	}

	// TODO: Andere Barriertypen berücksichtigen
	// ggf auch mit next_way.Bicycle != "yes" &&

	//
	// Fahrbesonderheiten, durch die eine Geschwindigkeitsermittlung der Straße nicht mehr notwendig wird
	//

	if speed == 0 {
		switch {
		// Kreisverkehre wenn möglich umfahren
		case next_way.Tags["junction"] == "roundabout":
			speed = 5

		// Fahrräder explizit verboten
		case next_way.Bicycle == "no":
			hints.AddWayHint(via, "Das Rad muss hier geschoben werden!")
			speed = 5

		// Kundenwege oder Privatwege nicht präferieren
		case next_way.Access == "private":
			speed = 1

		// access = customers hier nicht berücksichtigt; wie?

		// Treppen meiden
		case next_way.Highway == "steps":
			hints.AddWayHint(via, "Achtung! Auf diesem Weg befindet sich eine Treppe, die nicht umfahren werden konnte.")
			speed = 1
			penalties = 120

		// Kopfsteinpflaster und anderen schlechten Untergrund meiden
		case next_way.Tags["surface"] == "sett" ||
			next_way.Tags["surface"] == "cobblestone" ||
			next_way.Tags["surface"] == "cobblestone:flattened":
			speed = 5
		}
	}

	if speed == 0 {
		// Grundsätzlich: Straßen speziell für Fahrräder haben Vorrang
		if hasBicycleSupport {
			speed = STANDARD_BIKE_SPEED + 5

			// Schönen Pfad präferieren (z. B. Mauerweg)
			if next_way.Highway == "path" || next_way.Highway == "track" {
				speed += 5
			}
		} else {
			//
			// Wenn nicht speziell für Fahrräder, dann Straßengeschwindigkeit zunächst errechnen
			//
			switch next_way.Highway {

			// Servicestraßen
			case "service":
				speed = 1

			// Reine Fußgängerwege, kein Fahrrad erlaubt
			case "pedestrian", "footway":
				hints.AddWayHint(via, "Achtung Fußgängerweg! Das Rad muss hier geschoben werden!")
				speed = 5

			case "track":
				tracktype := next_way.Tags["tracktype"]
				if tracktype != "grade1" && tracktype != "grade2" {
					// Überwiegend bis völlig unbefestigt
					speed = 5
				}
			}
		}

		// Wenn jetzt noch speed == 0 ist, dann Standard nehmen
		if speed == 0 {
			// Trunk (wenn mit Fahrradunterstützung), Primary, Secondary, Tertiary, alle anderen Wege
			speed = STANDARD_BIKE_SPEED
		}

		// Bonus für beleuchtete Straßen
		if next_way.Tags["lit"] == "yes" {
			speed += 1
		}
	}

	//
	// Reisezeit ermitteln
	//

	// Bei Ampeln warten

	if next.Tags["highway"] == "traffic_signals" {
		penalties += TRAFFIC_LIGHT_WAIT_TIME
	}

	// Bahnübergang
	if next.Tags["railway"] == "level_crossing" {
		penalties += LEVEL_CROSSING_TIME
	}

	traveltime := ((distance / speed) * 60 * 60) + float64(penalties) // in sekunden
	return int(traveltime * 1000), true                               // aber in millisekunden zurückliefern
}
