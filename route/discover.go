package main

import (
	"math"

	"private/routenplaner/src/src/common"
)

func discover(lat, lon float64, profile string) *common.Node {
	d := int(math.Floor(distance(
		osm.NodeLookupCenter.Lat,
		osm.NodeLookupCenter.Lon,
		lat,
		lon)*10+0.5)) * 100

	// fmt.Printf("Asking for (%s)... ", profile)

	// Direct hit?
	nearest_diff := 99999999999

	if _, has_nodes := osm.NodeLookup[d]; !has_nodes {
		for distance, _ := range osm.NodeLookup {
			if nearest_diff > abs(distance-d) {
				nearest_diff = distance
			}
		}
	} else {
		nearest_diff = d
	}

	// Search through all nodes from the group nearest_distance
	var result *common.Node
	var current_distance float64
	for _, node_id := range osm.NodeLookup[nearest_diff] {
		node := osm.Nodes.Get(node_id)

		// Ignore nodes without ways
		if len(node.Ways) <= 0 {
			continue
		}

		// Ggf. Profilbeschränkungen enforcen
		switch profile {
		case "car":
			allowed := false
			for _, way_id := range node.Ways {
				way := osm.Ways.Get(way_id)
				if way.CanBeUsedByCars() {
					allowed = true
					break
				}
			}
			if !allowed {
				continue
			}
		}

		// Gehört der Node zu einer Straße, die auch befahrbar ist?
		// Gültig für alle (Autos, Fahrräder, Fußgänger)
		drivable := false
		for _, way_id := range node.Ways {
			way := osm.Ways.Get(way_id)
			if !way.InConstruction() {
				drivable = true
				break
			}
		}
		if !drivable {
			continue
		}

		if result == nil {
			result = node
			current_distance = distance(node.Lat, node.Lon, lat, lon)
		} else {
			check_distance := distance(node.Lat, node.Lon, lat, lon)
			if current_distance > check_distance {
				result = node
				current_distance = check_distance
			}
		}
	}

	return result
}
