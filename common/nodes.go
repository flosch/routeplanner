package common

import (
	"sort"
	"strings"
)

func (node *Node) WayText(osm *OSMBinary) string {
	way_names := make(sort.StringSlice, 0)
	for _, way_id := range node.Ways {
		way := osm.Ways.Get(way_id)
		if way.Name != "" { // Ignoriere Straßen ohne Namen, z. B. "Weg" oder "Feldweg"
			way_names = append(way_names, way.Fullname())
		}
	}
	sort.Sort(way_names)
	return strings.Join(way_names, "/")
}

func (node *Node) CommonWay(verifier, other *Node, osm *OSMBinary) *Way {
	var candidates []*Way
	for _, way1_id := range node.Ways {
		for _, way2_id := range other.Ways {
			if way1_id == way2_id {
				// Common way
				return osm.Ways.Get(way1_id)
				// candidates = append(candidates, osm.Ways.Get(way1_id))
			}
		}
	}

	return nil

	if len(candidates) > 1 {
		// Mehr als ein Kandidat als gemeinsame Straße, dann nehmen wir den verifier
		// zu hilfe
		if verifier == nil {
			return candidates[0] // FIXME DIRTY HACK, muss in der nächsten Routenplaner-Version korrigiert werden
			//panic("CommonWay kann nicht verifiziert werden, da verifier-Node fehlt")
		} else {
			var verified_candidates []*Way
			for _, way_v := range verifier.Ways {
				for _, candidate := range candidates {
					if way_v == candidate.ID {
						verified_candidates = append(verified_candidates, candidate)
					}
				}
			}
			// FIXME: Eigentlich korrekt, kann aber jetzt nicht korrigiert werden:
			lencan := len(verified_candidates)
			if lencan == 0 {
				return nil
				// panic("verified_candidates == 0")
			} else {
				return verified_candidates[0]
			}
		}
	}

	return candidates[0]
}
