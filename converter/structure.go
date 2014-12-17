package main

import (
	"encoding/gob"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"private/routenplaner/src/src/common"
)

func convert(xmlData *XMLOSM) (*common.OSMBinary, error) {
	var err error

	osm := &common.OSMBinary{
		Version: xmlData.Version,
		Bounds: common.Bounds{
			MinLat: xmlData.Bounds.MinLat,
			MinLon: xmlData.Bounds.MinLon,
			MaxLat: xmlData.Bounds.MaxLat,
			MaxLon: xmlData.Bounds.MaxLon,
		},
		Nodes:        make([]*common.Node, 0),
		Ways:         make([]*common.Way, 0),
		Relations:    make([]*common.Relation, 0),
		Restrictions: make([]int, 0),
		NodeLookup:   make(map[int][]int),
	}

	//
	// Nodes
	//
	for _, xmlNode := range xmlData.Nodes {
		node := &common.Node{
			ID:         xmlNode.ID,
			Version:    xmlNode.Version,
			Lat:        xmlNode.Lat,
			Lon:        xmlNode.Lon,
			Tags:       make(map[string]string),
			Neighbours: make([]int, 0),
			Ways:       make([]int, 0),
		}
		osm.Nodes = append(osm.Nodes, node)

		for _, tag := range xmlNode.Tags {
			node.Tags[tag.Key] = tag.Value
		}

		node.Access = node.Tags["access"]
		delete(node.Tags, "access")

		// No need for this node anymore, can be cleaned up by garbage collection
		// This 'hack' is not very clean, but might work
		// xmlData.Nodes[idx] = nil
	}

	osm.Nodes.Sort() // do a sort

	//
	// Ways
	//
	for _, xmlWay := range xmlData.Ways {
		way := &common.Way{
			ID:      xmlWay.ID,
			Version: xmlWay.Version,
			Nodes:   make([]int, 0, len(xmlWay.Nodes)),
			Tags:    make(map[string]string),
		}

		// tags := make(map[string]string)
		// Read tags
		for _, tag := range xmlWay.Tags {
			way.Tags[tag.Key] = tag.Value
		}

		// Extract some important tags
		way.Name = way.Tags["name"]
		delete(way.Tags, "name")
		way.Ref = way.Tags["ref"]
		delete(way.Tags, "ref")
		way.Highway = way.Tags["highway"]
		delete(way.Tags, "highway")

		way.Cycleway = way.Tags["cycleway"]
		delete(way.Tags, "cycleway")
		way.CyclewayLeft = way.Tags["cycleway:left"]
		delete(way.Tags, "cycleway:left")
		way.CyclewayRight = way.Tags["cycleway:right"]
		delete(way.Tags, "cycleway:right")

		way.Bicycle = way.Tags["bicycle"]
		delete(way.Tags, "bicycle")
		way.BicycleRoad = way.Tags["bicycle_road"]
		delete(way.Tags, "bicycle_road")

		way.Access = way.Tags["access"]
		delete(way.Tags, "access")
		way.Tunnel = way.Tags["tunnel"]
		delete(way.Tags, "tunnel")

		if s_lanes, has_lanes := way.Tags["lanes"]; has_lanes {
			i, err := strconv.Atoi(s_lanes)
			if err != nil {
				fmt.Printf("Could not interpret way tag 'lanes': %s\n", err)
			} else {
				way.Lanes = i
			}
		}
		if s_maxspeed, has_maxspeed := way.Tags["maxspeed"]; has_maxspeed {
			maxspeed := 0
			if s_maxspeed == "walk" {
				maxspeed = 5 // 5 km/h Schrittgeschwindigkeit
			} else {
				maxspeed, err = strconv.Atoi(s_maxspeed)
				if err != nil {
					fmt.Printf("Could not interpret way tag 'maxspeed' (%s): %s\n", s_maxspeed, err)
					maxspeed = 0
				}
			}
			way.Maxspeed = maxspeed
		}

		osm.Ways = append(osm.Ways, way)

		way.Oneway = 0
		if value, is_oneway := way.Tags["oneway"]; is_oneway {
			if value == "yes" || value == "true" || value == "1" {
				way.Oneway = 1
			} else if value == "-1" || value == "reverse" {
				way.Oneway = -1
			}
		}

		// Nur dann Verbindungen zwischen Straßenknoten herstellen, wenn es
		// sich auch um eine Straße handelt
		if way.Highway != "" {
			// Read nodes and create connections between nodes
			for idx, nodeRef := range xmlWay.Nodes {
				node := osm.Nodes.Get(nodeRef.Ref)

				// Das darf eigentlich nicht passieren, sonst ist das OSM-file unvollständig
				if node == nil {
					return nil, errors.New(fmt.Sprintf("node %d not found in node db", nodeRef.Ref))
				}

				// Way alle Nodes zuweisen
				way.Nodes = append(way.Nodes, nodeRef.Ref)

				// Dem jeweiligen Node eine Referenz auf den Way geben
				node.Ways = append(node.Ways, xmlWay.ID)

				// Nodes verknüpfen
				if idx+1 < len(xmlWay.Nodes) {
					// Es braucht zwei für eine Verknüpfung:
					nextRef := xmlWay.Nodes[idx+1]
					next := osm.Nodes.Get(nextRef.Ref)

					// Darf nicht passieren, sonst OSM-file unvollständig
					if next == nil {
						return nil, errors.New("Next ID not found")
					}

					node.UnfilteredNeighbours = append(node.UnfilteredNeighbours, nextRef.Ref)
					next.UnfilteredNeighbours = append(next.UnfilteredNeighbours, nodeRef.Ref)

					// Verknüpfung abhängig von Fahrtrichtung
					switch way.Oneway {
					case 1:
						// Oneway
						node.Neighbours = append(node.Neighbours, nextRef.Ref)
					case -1:
						// Oneway reverse
						next.Neighbours = append(next.Neighbours, nodeRef.Ref)
					case 0:
						// Kein Oneway
						node.Neighbours = append(node.Neighbours, nextRef.Ref)
						next.Neighbours = append(next.Neighbours, nodeRef.Ref)
					}

				}
			}
		} else {
			// Bei dem Way handelt es sich um keine Straße (sondern vllt. ein Gebäude), aber trotzdem
			// werden wenigstens dem WAY(!) alle seine Knoten zugefügt
			for _, nodeRef := range xmlWay.Nodes {
				node := osm.Nodes.Get(nodeRef.Ref)

				// Das darf eigentlich nicht passieren, sonst ist das OSM-file unvollständig
				if node == nil {
					return nil, errors.New(fmt.Sprintf("node %d not found in node db", nodeRef.Ref))
				}

				// Way alle Nodes zuweisen
				way.Nodes = append(way.Nodes, nodeRef.Ref)
			}
		}
	}

	osm.Ways.Sort()

	//
	// Relations
	//
	for _, xmlRelation := range xmlData.Relations {
		relation := &common.Relation{
			ID:      xmlRelation.ID,
			Members: make([]*common.Member, 0, len(xmlRelation.Members)),
			Tags:    make(map[string]string),
		}

		for _, xmlMember := range xmlRelation.Members {
			member := &common.Member{
				Type: xmlMember.Type,
				Ref:  xmlMember.Ref,
				Role: xmlMember.Role,
			}
			relation.Members = append(relation.Members, member)
		}

		for _, tag := range xmlRelation.Tags {
			relation.Tags[tag.Key] = tag.Value
		}

		// If this relation is a restriction, put it in the restrictions-index
		if relation.Tags["type"] == "restriction" {
			osm.Restrictions = append(osm.Restrictions, relation.ID)

			// Restrictions den einzelnen Ways direkt zuordnen
			// Felder von den Members extrahieren

			from := -1 // way
			via := -1  // node OR way
			to := -1   // way
			location_hint := 0

			for _, member := range relation.Members {
				if member.Role == "from" {
					from = member.Ref
				} else if member.Role == "via" {
					via = member.Ref
				} else if member.Role == "to" {
					to = member.Ref
				} else if member.Role == "location_hint" {
					location_hint = member.Ref
				} else {
					fmt.Printf("Unknown role: %s in %d\n", member.Role, relation.ID)
				}
			}

			if from < 0 || to < 0 || via < 0 {
				fmt.Printf("Restriction malformed: from/to/via is < 0 (relation %d)\n", relation.ID)

				// Fehlerhafte Restriction, alle drei Werte MÜSSEN gegeben sein
				goto skip
			}

			// Wenn der Weg passt, Restriction anlegen
			way := osm.Ways.Get(from)
			if way == nil {
				fmt.Printf("Way %d for relation %d not found.\n", from, relation.ID)
				goto skip
			}

			restriction := &common.Restriction{
				Via:           via,
				To:            to,
				Location_hint: location_hint,
			}

			res := relation.Tags["restriction"]
			if strings.HasPrefix(res, "only_") {
				way.OnlyRestrictions = append(way.OnlyRestrictions, restriction)
			} else if strings.HasPrefix(res, "no_") {
				way.NoRestrictions = append(way.NoRestrictions, restriction)
			} else {
				fmt.Printf("Restriction not handled: %s\n", relation.Tags["restriction"])
			}
		}

	skip:
		osm.Relations = append(osm.Relations, relation)
	}

	//
	// Lookup-Datenbank
	//
	const steps = 3
	osm.NodeLookupCenter = osm.Nodes[0]
	for _, node := range osm.Nodes {
		// Nur interessante Nodes berücksichtigen (Straßen, Einrichtungen, Shops)
		_, is_amenity := node.Tags["amenity"]
		_, is_shop := node.Tags["shop"]
		_, has_highway := node.Tags["highway"]
		is_part_of_way := len(node.Ways) > 0
		if !is_amenity && !is_shop && !is_part_of_way && !has_highway {
			continue
		}
		d := int(math.Floor(distanceNodes(osm.NodeLookupCenter, node)*10+0.5)) * 100
		osm.NodeLookup[d] = append(osm.NodeLookup[d], node.ID)
	}

	/*for k, v := range osm.NodeLookup {
		fmt.Printf("Items distance = %d: %d\n", k, len(v))
	}*/

	return osm, nil
}

func saveBinary(filename string, osm *common.OSMBinary) error {
	fout, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fout.Close()

	// Save it to disk
	err = gob.NewEncoder(fout).Encode(osm)
	if err != nil {
		panic(err)
		return err
	}

	return nil
}
