package main

import (
	"errors"
	"fmt"
	"time"

	"private/routenplaner/src/src/common"
)

func (request *RouteRequest) expand(item *prioItem) {
	var prev *common.Node

	if item.node.ID == 2149039562 {
		fmt.Printf("y: %d", len(item.node.Neighbours))
	}

	for _, next_id := range item.node.Neighbours {
		if item.node.ID == 2149039562 {
			fmt.Printf("...auf %d...", next_id)
		}
		next := request.osm.Nodes.Get(next_id)

		if next == nil {
			fmt.Printf("Coulnd't find node %d\n", next_id)
			continue
		}

		// Schon besucht? Dann nächsten Knoten
		if _, has_visited := request.closedlist[next.ID]; has_visited {
			continue
		}
		request.closedlist[next.ID] = true

		// Vorgänger holen
		if item.prev != nil {
			prev = item.prev.node
		} else {
			prev = nil
		}

		// Weg von dem man kommt
		var prev_way *common.Way
		if prev != nil {
			// TODO: Hier ist verifier-node aktuell NIL, man müsste den prev vom
			// prev noch ermitteln
			prev_way = prev.CommonWay(nil, item.node, request.osm)
		}

		// Weg, den man nun einschlägt
		next_way := item.node.CommonWay(prev, next, request.osm)

		if prev != nil {
			//	fmt.Printf("%s -> %s [via: %d]\n", prev_way.Fullname(), next_way.Fullname(), item.node.ID)
		}

		// Hint-Manager erstellen für diesen Weg next_way
		hintMgr := newHintMgr(next_way)

		// Generelle Kosten für den Weg berechnen
		general_costs, allowed := profile_general(hintMgr, prev, item.node, next, prev_way, next_way, request)
		if !allowed {
			// Grundsätzlich nicht erlaubt, diese Kante zu nehmen (z. B. Abbiegebeschränkung)
			// WICHTIG!!!!!
			// Wenn das passiert, muss aber der Zielknoten (also next) nochmal freigegeben
			// werden. Es könnte ja sein, dass er Teil einer anderen Lösungsstrecke ist.
			fmt.Printf("not allowed from %d to %d\n", item.node.ID, next_id)
			delete(request.closedlist, next.ID)
			delete(request.g_values, next.ID)
			delete(request.prio_items, next.ID)
			continue
		}

		// Hole profilabhängige Kosten (z. B. fürs Fahrrad, Auto, usw.)
		profile_costs, allowed := request.costProfile(hintMgr, prev, item.node, next, prev_way, next_way, request)
		if !allowed {
			// Grundsätzlich nicht erlaubt, diese Kante zu nehmen (z. B. Abbiegebeschränkung)
			// WICHTIG!!!!!
			// Wenn das passiert, muss aber der Zielknoten (also next) nochmal freigegeben
			// werden. Es könnte ja sein, dass er Teil einer anderen Lösungsstrecke ist.
			fmt.Printf("not allowed\n")
			delete(request.closedlist, next.ID)
			delete(request.g_values, next.ID)
			delete(request.prio_items, next.ID)
			continue
		}
		// Vorläufiges g(next) und f(next) berechnen
		tentative_g := request.g_values[item.node.ID] + general_costs + profile_costs

		// Wir berechnen die Kosten in Fahrtzeit in Minuten
		// !ABHÄNGIG VOM PROFIL!
		var traveltime float64
		switch request.CostProfileText {
		case "car":
			traveltime = (distanceNodes(next, request.Destination) / STANDARD_CAR_SPEED) * 60 * 60 * 1000 // in millisekunden
		case "bike":
			traveltime = (distanceNodes(next, request.Destination) / STANDARD_BIKE_SPEED) * 60 * 60 * 1000 // in millisekunden
		default:
			panic("unspported profile")
		}
		f := tentative_g + int(traveltime+0.5) // 0.5, um es korrekt zu runden

		if item.node.ID == 2149039562 {
			fmt.Printf(" (%d / %.2f) ", f, float64(profile_costs)/1000.0)
		}

		// Kann man den aktuellen Knoten günstiger erreichen über item?
		if old_g, has_g := request.g_values[next.ID]; has_g {
			if old_g > tentative_g {
				// Besseren Weg zu diesem Knoten gefunden, g und prev+hint updaten
				request.g_values[next.ID] = tentative_g
				request.prio_items[next.ID].prev = item
				request.prio_items[next.ID].hint = hintMgr
				request.prio_items[next.ID].priority = f
				request.openlist.Notifiy(request.prio_items[next.ID]) // informiere OpenList über neuen Weert
			} else {
				// Kein besserer? Dann nächsten Knoten untersuchen
				continue
			}
		} else {
			// Knoten noch nicht bekannt? Einqueuen und eintragen
			request.prio_items[next.ID] = request.openlist.add(item, next, f, hintMgr)
			request.g_values[next.ID] = tentative_g
		}
	}
	if item.node.ID == 2149039562 {
		fmt.Printf("\n")
	}
}

const MAX_CALCULATION_TIME = time.Second * 3

func (request *RouteRequest) calculate() (*RouteResponse, error) {
	// TODO: Nicht 2 mal calculate() ausführen lassen

	request.calcuation_start = time.Now()
	request.openlist.add(nil, request.Departure, 0, nil)

	var next *prioItem
	for request.openlist.Len() > 0 {
		next = request.openlist.next()

		// TODO: An dieser Stelle auch Via berücksichtigen
		// 2 Optionen für Vias: Automatische Suche nach der besten Reihenfolge der Vias
		// oder Reihenfolge

		if next.node == request.Destination {
			// Ziel gefunden
			return makeRouteResponse(request, next), nil
		}

		request.expand(next)

		if time.Since(request.calcuation_start) > MAX_CALCULATION_TIME {
			return nil, errors.New("Calculation aborted, maximum calculation time exceeded")
		}
	}

	return nil, errors.New(fmt.Sprintf("No route found from %d to %d",
		request.Departure.ID, request.Destination.ID))
}
