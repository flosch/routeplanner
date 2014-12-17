package main

import (
	"private/routenplaner/src/src/common"
)

/*
 * Das allgemeine Profil wird immer aufgerufen und berücksichtigt z. B.
 * - Abbiegebeschränkungen
 *
 * Üblicherweise werden 0 Kosten zurückgegeben, > 0 sind jedoch möglich. Es wird mehr mit
 * allowed gearbeitet, um Kantengänge auszuschließen
 */

func profile_general(hints *HintMgr, prev, via, next *common.Node, prev_way, next_way *common.Way, request *RouteRequest) (costs int, allowed bool) {
	// Nur jegliche Formen von Wegen benutzen
	if next_way.Highway == "" {
		return 0, false
	}

	if next_way.InConstruction() {
		// geplante Straße ignorieren
		return 0, false
	}

	return 0, true
}
