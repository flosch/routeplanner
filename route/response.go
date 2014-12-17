package main

import (
	"fmt"

	"private/routenplaner/src/src/common"
)

type RouteNode struct {
	Prev *common.Node
	Node *common.Node
	Hint *HintMgr // Hint für den Weg von Prev -> Node
}

type RouteResponse struct {
	osm     *common.OSMBinary
	request *RouteRequest
	Nodes   []*RouteNode
}

func makeRouteResponse(request *RouteRequest, dest *prioItem) *RouteResponse {
	response := &RouteResponse{
		osm:     request.osm,
		request: request,
		Nodes:   make([]*RouteNode, 0, 250),
	}

	item := dest
	for item != nil {
		rn := &RouteNode{
			Node: item.node,
			Hint: item.hint,
		}
		if item.prev != nil {
			rn.Prev = item.prev.node
		}
		response.Nodes = append(response.Nodes, rn)
		item = item.prev
	}

	// Reverse list
	n := len(response.Nodes)
	for i := 0; i < n/2; i++ {
		response.Nodes[i], response.Nodes[n-1-i] = response.Nodes[n-1-i], response.Nodes[i]
	}

	if len(response.Nodes) < 2 {
		// Wenn weniger als 2 Nodes, dann gibt es nicht mal ein Start UND ein Ziel
		return nil
	}

	return response
}

// Gesamtdistanz in Metern
func (response *RouteResponse) Distance() int {
	total := 0
	for _, step := range response.steps() {
		total += step.Distance()
	}
	return total
}

// Gibt auf Basis der Geschwindigkeitsbegrenzungen und des Profils eine Schätzung darüber ab,
// wie lange man für die Tour benötig
// Gibt (Minuten, Note) zurück
func (response *RouteResponse) TravelTime() (int, string) {
	total_time := float64(0.0) // in Minuten

	// Um nicht mehrere Ampeln über Steps hinweg doppelt zu zählen, hier Buch über alle Ampeln führen
	traffic_signals_accouting := make(map[int]bool)

	for _, step := range response.steps() {
		// Allgemein gültige Annahme: Je Ampel +40s Verzögerung
		for _, node := range step.nodes {
			if node.Tags["highway"] == "traffic_signals" {
				if _, had_already_considered := traffic_signals_accouting[node.ID]; !had_already_considered {
					traffic_signals_accouting[node.ID] = true
					total_time += float64(TRAFFIC_LIGHT_WAIT_TIME) / 60
				}
			}
		}

		switch response.request.CostProfileText {
		case "bike":
			// Annahme füre in Fahrrad: 15km/h

			var speed float64
			if step.Maxspeed > 0 {
				speed = float64(min(15, step.Maxspeed))
			} else {
				speed = 15.0
			}
			total_time += ((float64(step.Distance()) / 1000.0) / speed) * 60
		case "car":
			// Wenn keine Geschwindigkeitsbegrenzung, dann Annahme von 40 km/h für ein Auto

			var speed float64
			if step.Maxspeed > 0 {
				speed = float64(step.Maxspeed) * 0.90 // es wird etwas pessimistisch mit nur 90% der max. Geschwindigkeit gerechnet
			} else {
				// TODO: Geschwindigkeitsberechnung abhängig vom Profil (bereits in profile_car gemacht)
				// Könnte man herausfaktorisierne und an beiden Stellen nutzen
				speed = 30 // TODO: Pauschal, hier wird aber nicht der Straßentyp berücksichtigt (was ich jedoch in profile_car() gemacht habe)
			}
			total_time += ((float64(step.Distance()) / 1000.0) / speed) * 60
		default:
			panic("Profile not supported yet")
		}
	}
	note := ""
	switch response.request.CostProfileText {
	case "bike":
		// Annahme: 15km/h
		note = "bei durchschnittlich 15 km/h und unter Berücksichtigung von Geschwindigkeitsbegrenzungen und Haltephasen an Ampeln"
	case "car":
		note = "unter Berücksichtigung von Geschwindigkeitsbegrenzungen und Haltephasen an Ampeln"
	default:
		panic("Profile not supported yet")
	}
	return int(total_time), note
}

func (response *RouteResponse) DepartureText() string {
	name := response.request.Departure.WayText(response.osm)
	if name == "" {
		name = fmt.Sprintf("Weg (ID %d)", response.request.Departure.ID)
	}
	return name
}

func (response *RouteResponse) DestinationText() string {
	name := response.request.Destination.WayText(response.osm)
	if name == "" {
		name = fmt.Sprintf("Weg (ID %d)", response.request.Destination.ID)
	}
	return name
}

// Liefert in % (0 - 100) zurück, inwiefern die gesamte Route fahrradkompatibel ist
func (response *RouteResponse) BicycleSupportPercentage() int {
	total_route_bicycle_length := 0
	total_distance := 0
	for _, step := range response.steps() {
		total_distance += step.Distance()
		if step.ways.HasBicycleSupport() > 0 {
			total_route_bicycle_length += step.Distance()
		}
	}

	return int((float64(total_route_bicycle_length) / float64(total_distance)) * 100)
}
