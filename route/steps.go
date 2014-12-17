package main

import (
	"fmt"
	"math"

	"private/routenplaner/src/src/common"
)

const (
	START = iota
	GOAL
	DIRECTION_CROSS        // überqueren Sie eine straße
	DIRECTION_HOLD         // weiter auf
	DIRECTION_LEFT         // links abbiegen
	DIRECTION_RIGHT        // rechts abbiegen
	DIRECTION_SLIGHT_LEFT  // leicht links
	DIRECTION_SLIGHT_RIGHT // leicht rechts
	DIRECTION_SHARP_LEFT   // leicht links, um auf derselben Straße zu bleiben
	DIRECTION_SHARP_RIGHT  // leicht rechts, um auf derselben Straße zu bleiben

)

type Step struct {
	Maxspeed int
	Tunnel   string // "in", "out", ""
	Motorway string // "in", "out", ""

	instruction    int
	nodes          []*common.Node // alle Nodes (>= 2!), die zu diesem Schritt gehören
	ways           StepWayList    // die Straßen, auf denen man gerade fährt (können mehr als eine sein, z. B. wenn dieselbe Straße aufgeteilt ist in mehrere Abschnitte)
	cross          *common.Way    // wird eine Straße überquert?
	hold           bool           // (Links/rechts/...) abbiegen, "um auf derselben Straße [wie der Step zuvor] zu bleiben"
	hints          *HintMgr
	goal_from_node *common.Node

	prev *Step
	next *Step
}

func newStep() *Step {
	return &Step{
		nodes: make([]*common.Node, 0),
		hints: newHintMgr(nil),
	}
}

type StepWayList []*common.Way

// mögliche rückgabewerte: 0 = nein, 1 = ja, 2 = teilweise
func (swl StepWayList) HasBicycleSupport() int {
	every_way := swl[0].HasBicycleSupport()
	for _, way := range swl[1:] {
		if way.HasBicycleSupport() {
			if !every_way {
				return 2
			}
		}
	}
	if every_way {
		return 1
	}
	return 0
}

func direction(from, via, to *common.Node) int {
	d := degree(from, via, to)
	switch {
	case d > 300:
		return DIRECTION_SHARP_LEFT
	case (d <= 300) && (d > 210):
		return DIRECTION_LEFT
	case (d >= 190) && (d <= 210):
		return DIRECTION_SLIGHT_LEFT
	case (d > 170) && (d < 190):
		return DIRECTION_HOLD
	case (d <= 170) && (d >= 150):
		return DIRECTION_SLIGHT_RIGHT
	case (d < 150) && (d >= 60):
		return DIRECTION_RIGHT
	case d < 60:
		return DIRECTION_SHARP_RIGHT
	case math.IsNaN(d):
		fmt.Printf("degree = NaN (from = %d, via = %d, to = %d)\n", from.ID, via.ID, to.ID)
		return DIRECTION_HOLD
	}
	panic(fmt.Sprintf("unreachable: d = %f", d))
}

func (response *RouteResponse) steps() []*Step {
	steps := make([]*Step, 0, 100)

	// Calculate the steps
	var current_step = newStep()
	current_step.instruction = START
	current_step.nodes = append(current_step.nodes, response.Nodes[0].Node)
	steps = append(steps, current_step)

	if response.Nodes[0].Hint != nil {
		current_step.hints.Merge(response.Nodes[0].Hint)
		// current_step.cost_hints = append(current_step.cost_hints, newHint(response.Nodes[0].Node, response.Nodes[0].Hint))
	}

	var prev *common.Node
	var next *common.Node
	var current_way *common.Way = response.Nodes[0].Node.CommonWay(nil, response.Nodes[1].Node, response.osm)
	current_step.ways = append(current_step.ways, current_way)

	current_step.Maxspeed = current_way.Maxspeed

	for idx, current_route_item := range response.Nodes[1 : len(response.Nodes)-1] {
		// maximale erlaubte Geschwindigkeit ermitteln
		// TODO: Nicht ganz korrekt, da auch nur ein Teil der Straße beschränkt sein kann (auch andere Geschwindigkeiten möglich)
		if current_step.Maxspeed <= 0 {
			current_step.Maxspeed = current_way.Maxspeed
		} else {
			if current_way.Maxspeed > 0 {
				current_step.Maxspeed = min(current_step.Maxspeed, current_way.Maxspeed)
			}
		}

		current := current_route_item.Node

		current_index := idx + 1
		current_step.nodes = append(current_step.nodes, current)
		if current_route_item.Hint != nil {
			current_step.hints.Merge(current_route_item.Hint)
			//current_step.cost_hints = append(current_step.cost_hints, newHint(current, current_route_item.Hint))
		}

		prev = response.Nodes[current_index-1].Node
		next = response.Nodes[current_index+1].Node

		// Ist man weiterhin auf demselben Weg oder steht ein Straßenwechsle bevor?
		next_way := current.CommonWay(prev, next, response.osm)

		if current_way != next_way {
			// Bisherige Straße und neue Straße sind nicht mehr identisch
			// Prev --[current_way]--> Current --[next_way]--> Next --[next2_way]--> Next2

			// Wechsel der Straße ignorieren, wenn
			// - der Straßenname identisch ist und
			// - direction(prev, current, next) == HOLD (d.h. nicht abgebogen werden muss)
			// - und man nicht nun durch einen Tunnel fährt (d.h. Tunnelwerte identisch sind)
			if current_way.Streetname() == next_way.Streetname() &&
				direction(prev, current, next) == DIRECTION_HOLD &&
				current_way.Tunnel == next_way.Tunnel {
				current_way = next_way
				current_step.ways = append(current_step.ways, next_way)
				continue
			}

			// - Wenn der Straßenname nicht identisch ist ODER
			// - sich die Richtung stärker als 'HOLD' verändert ODER
			// - man nun durch einen Tunnel fährt,
			// wird ein neuer Step angelegt mit der jeweiligen Richtung

			next_step := newStep()
			next_step.prev = current_step
			current_step.next = next_step
			current_step = next_step
			steps = append(steps, current_step)
			current_step.nodes = append(current_step.nodes, current) // Startknoten
			current_step.ways = append(current_step.ways, next_way)
			/*
				!! Dies hier nicht machen, da diese Informationen bereits im Step zuvor enthalten sind und sich auch darauf beziehen.
				if current_route_item.Hint != "" {
					current_step.cost_hints = append(current_step.cost_hints, newHint(current, current_route_item.Hint))
				}
			*/

			if current_way.Tunnel != next_way.Tunnel {
				// Fährt man IN einen Tunnel?
				if next_way.Tunnel != "no" && next_way.Tunnel != "" { // dann muss current_way.Tunnel == "no" sein
					current_step.Tunnel = "in"
				} else {
					current_step.Tunnel = "out"
				}
			}

			if current_way.Highway != "motorway" && current_way.Highway != "motorway_link" &&
				(next_way.Highway == "motorway_link" || next_way.Highway == "motorway") {
				// Nun auf Autobahnauffahrt
				current_step.Motorway = "in"
				current_step.hints.AddWayHint(next, "Sie fahren nun auf die Autobahn.")
			}

			if (current_way.Highway == "motorway_link" || current_way.Highway == "motorway") &&
				(next_way.Highway != "motorway_link" && next_way.Highway != "motorway") &&
				current_step.prev != nil {
				// Nun auf Autobahnabfahrt
				current_step.prev.Motorway = "out"
				current_step.prev.hints.AddWayHint(next, "Sie verlassen jetzt die Autobahn.")
			}

			if name, has := current.Tags["motorway_junction"]; has && next_way.Highway == "motorway_link" {
				current_step.prev.hints.AddWayHint(next, fmt.Sprintf("Nehmen Sie die Abfahrt %s.", name))
			}

			// Wegname ist derselbe, also auf dem Weg weiterfahren
			// Es wird also nur die Richtung ggf. verändert
			if current_way.Streetname() == next_way.Streetname() {
				// (Instruction mit HOLD)
				// Frage ist nur, ob links/rechts/leicht links/leicht rechts
				current_step.hold = true
			}

			// Wird nur ein großer Damm überquert?
			if current_index+2 < len(response.Nodes) { // Dafür benötigen wir Next2 (s. o.)
				next2 := response.Nodes[current_index+2].Node
				next2_way := next.CommonWay(current, next2, response.osm)

				// Überquerung eines Damms von Current auf Next, wenn:
				// - current und next sind ampeln
				// - distanz(current, next) <= 50 meter
				// - direction(current, next, next2) == hold
				// - next_way != next_way2

				if distanceNodes(current, next) < 50 &&
					direction(current, next, next2) == DIRECTION_HOLD &&
					current.Tags["highway"] == "traffic_signals" &&
					next.Tags["highway"] == "traffic_signals" &&
					next2_way.Streetname() != next_way.Streetname() {

					// Wir überqueren nur einen Damm
					current_step.instruction = DIRECTION_CROSS
					current_way = next_way
					continue
				}
			}

			// Neue Richtung ermitteln
			current_step.instruction = direction(prev, current, next)

			if next_way == nil {
				panic("Darf nicht passieren")
			}
			current_way = next_way
			continue
		}
	}

	finish_step := newStep()
	finish_step.ways = append(finish_step.ways, current_way)
	finish_step.nodes = append(finish_step.nodes, response.Nodes[len(response.Nodes)-1].Node)
	finish_step.nodes = append(finish_step.nodes, response.request.Destination)
	finish_step.instruction = GOAL
	finish_step.goal_from_node = response.Nodes[len(response.Nodes)-2].Node
	steps = append(steps, finish_step)

	return steps
}

func (step *Step) Text(lang string) string {
	text := ""

	switch lang {
	default: // Deutsch
		switch step.instruction {
		case START:
			return fmt.Sprintf("Beginnen Sie auf %s",
				step.ways[0].Fullname())

		case GOAL:
			return fmt.Sprintf("Sie haben Ihr Ziel %s erreicht.", step.nodes[1].WayText(osm))

		case DIRECTION_LEFT:
			text = "Links abbiegen"
		case DIRECTION_SLIGHT_LEFT:
			text = "Leicht links abbiegen"
		case DIRECTION_SHARP_LEFT:
			text = "Scharf links abbiegen"

		case DIRECTION_RIGHT:
			text = "Rechts abbiegen"
		case DIRECTION_SLIGHT_RIGHT:
			text = "Leicht rechts abbiegen"
		case DIRECTION_SHARP_RIGHT:
			text = "Scharf rechts abbiegen"

		case DIRECTION_HOLD:
			text = "Weiter"

		case DIRECTION_CROSS:
			text = fmt.Sprintf("Überqueren Sie %s.", step.ways[0].Fullname())

		default:
			panic(fmt.Sprintf("Unbekannte Anweisung: %d", step.instruction))
		}
	}

	// Tunnelmeldung berücksichtigen
	// (Links/Rechts/Weiter)... im Tunnel ... auf Weg...
	switch step.Tunnel {
	case "in":
		text = fmt.Sprintf("%s nun in Tunnel", text)
	case "out":
		text = fmt.Sprintf("%s nun aus Tunnel heraus", text)
	}

	if step.instruction >= DIRECTION_HOLD {
		text = fmt.Sprintf("%s auf %s", text, step.ways[0].Fullname())

		if step.hold {
			text = fmt.Sprintf("%s, um auf %s zu bleiben", text, step.ways[0].Fullname())
		}
	}

	return text
}

func (step *Step) Hints(lang string) []*Hint {
	profile_hints := step.hints.GetHints()
	hints := make([]*Hint, 0, len(profile_hints))
	hints = append(hints, profile_hints...)

	if step.next == nil {
		// Finish note hat keine hints
		return hints
	}

	switch lang {
	default: // Deutsch
		// Step nur sehr kurz? Prüfen und Hinweis geben, auf welcher Fahrspur man sich
		// einsortieren sollte (links/rechts)

		// TODO: Exakte Fahrspur als eigenes Attribut (maxlanes und lane#) angeben!

		if step.Distance() < 300 && step.next != nil {
			// Step unter 300 Metern, prüfen, ob abgebogen wird
			// Nur relevant, wenn mehr als 1 Lanes auf der Straße (bei oneway) oder mehr als 2 Lanes
			if (step.ways[0].Lanes > 1 && step.ways[0].Oneway != 0) || (step.ways[0].Lanes > 2) {
				switch step.next.instruction {
				case DIRECTION_LEFT, DIRECTION_SLIGHT_LEFT, DIRECTION_SHARP_LEFT:
					hints = append(hints,
						newHint(step.next.nodes[0],
							fmt.Sprintf("Danach links halten."),
						))
				case DIRECTION_RIGHT, DIRECTION_SLIGHT_RIGHT, DIRECTION_SHARP_RIGHT:
					hints = append(hints,
						newHint(step.next.nodes[0],
							fmt.Sprintf("Danach rechts halten."),
						))

					// TODO: Alle ergänzen
				}
			}
		}

		// "Danach direkt links/rechts/usw."
		// Nur dann, wenn zuvor nicht noch eine Straße mit derselben Richtung kommt
		// und es innerhalb der nächsten 150 Metern geschieht
		if step.Distance() < 150 && step.next != nil {
			/*d := direction(step.nodes[len(step.nodes)-1],
			step.next.nodes[0],
			step.next.nodes[1])*/

			make_hint := true
			for _, node := range step.nodes[1 : len(step.nodes)-1] {
				if len(node.Ways) > 1 {
					// Ein Knoten auf dem Weg zum nächsten Step hat mindestens eine
					// Abzweigung.

					// TODO: Prüfen, ob die Abwzweigung dieselbe ist! Es kann trotzdem
					// direkt hingewiesen werden, wenn die eine links, und die andere Straße
					// rechts abgeht

					make_hint = false
				}
			}
			if make_hint {
				switch step.next.instruction {
				case DIRECTION_LEFT, DIRECTION_SHARP_LEFT:
					hints = append(hints,
						newHint(step.next.nodes[0],
							fmt.Sprintf("Danach direkt links."),
						))
				case DIRECTION_RIGHT, DIRECTION_SHARP_RIGHT:
					hints = append(hints,
						newHint(step.next.nodes[0],
							fmt.Sprintf("Danach direkt rechts."),
						))
				}
			}
		}

	}

	return hints
}

// Wandelt die interne instruction in ein universales Stichwort zurück
// zum Beispiel DIRECTION_LEFT in "left". Genutzt z. B., um ein passendes
// Symbol anzeigen zu können
func (step *Step) Direction() string {
	switch step.instruction {
	case START:
		return "start"
	case GOAL:
		return "goal"
	case DIRECTION_CROSS:
		return "cross"
	case DIRECTION_HOLD:
		return "hold"
	case DIRECTION_LEFT:
		return "left"
	case DIRECTION_RIGHT:
		return "right"
	case DIRECTION_SHARP_LEFT:
		return "sharp_left"
	case DIRECTION_SHARP_RIGHT:
		return "sharp_right"
	case DIRECTION_SLIGHT_LEFT:
		return "slight_left"
	case DIRECTION_SLIGHT_RIGHT:
		return "slight_right"
	default:
		panic(fmt.Sprintf("%d is missing", step.instruction))
	}
}

// Gibt die Distanz dieses Schritts in Metern zurück
func (step *Step) Distance() int {
	total := 0

	var prev *common.Node = step.nodes[0]
	for _, node := range step.nodes[1:] {
		total += int(distanceNodes(prev, node) * 1000)
		prev = node
	}

	return total
}
