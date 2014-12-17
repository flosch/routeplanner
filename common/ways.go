package common

import "fmt"

func (way *Way) Streetname() string {
	name := way.Name
	if name == "" {
		name = way.Tags["reg_name"] // Regionaler Name
	}
	if way.Name == "" {
		// TODO: Bestimmte Typ des Wegs, z. B. Feldweg, Pfad oder nur Weg
		switch {
		case way.Highway == "motorway":
			name = way.Tags["reg_name"]
			if name == "" {
				name = "Autobahn"
			}
			dest := way.Tags["destination"]
			if dest == "" {
				return name
			}
			return fmt.Sprintf("%s Richtung %s", name, dest)
		case way.Highway == "motorway_link":
			// TODO: Hier fehlt die Information ob "Auffahrt" oder "Abfahrt"
			// (müsste man ermitteln dynamisch)
			name = way.Tags["reg_name"]
			if name == "" {
				name = "Autobahn-Anschlussrampe"
			}
			dest := way.Tags["destination"]
			if dest != "" {
				name = fmt.Sprintf("%s Richtung %s", name, dest)
			}
			ref := way.Tags["destination:ref"]
			if ref != "" {
				name = fmt.Sprintf("%s von %s", name, ref)
			}
			return name
		case way.Highway == "cycleway":
			return "Fahrradweg"
		case way.Cycleway != "":
			switch way.Cycleway {
			case "lane", "opposite", "opposite_lane":
				return "Radfahrstreifen"
			case "track", "opposite_track":
				return "Radweg"
			}
		case way.IsPrimary():
			name = "Bundesstraße"
		case way.IsSecondary():
			name = "Landesstraße"
		case way.IsTertiary():
			name = "Kreisstraße"
		case way.Highway == "path":
			name = "Pfad"
		case way.Highway == "track":
			name = "Feldweg"
		case way.Highway == "steps":
			name = "Treppe"
		case way.Highway == "road":
			name = "Straße"
		case way.Highway == "footway":
			name = "Fußgängerweg"
		}
	}
	ref := way.Tags["ref"]
	if ref != "" {
		name = fmt.Sprintf("%s (%s)", name, ref)
	}
	if name == "" {
		name = "Weg" // Letzter Ausweg
	}
	return name
}

func (way *Way) Fullname() string {
	if way.Ref != "" {
		return fmt.Sprintf("%s (%s)", way.Streetname(), way.Ref)
	}
	return way.Streetname()
}

func (way *Way) IsPrimary() bool {
	return way.Highway == "primary" ||
		way.Highway == "primary_link"
}

func (way *Way) IsSecondary() bool {
	return way.Highway == "secondary" ||
		way.Highway == "secondary_link"
}

func (way *Way) IsTertiary() bool {
	return way.Highway == "tertiary" ||
		way.Highway == "tertiary_link"
}

// Schnellstraße
func (way *Way) IsTrunk() bool {
	return way.Highway == "trunk" ||
		way.Highway == "trunk_link"
}

// Autobahn
func (way *Way) IsMotorway() bool {
	return way.Highway == "motorway" ||
		way.Highway == "motorway_link"
}

func (way *Way) IsSpeedway() bool {
	return way.IsMotorway() || way.IsTrunk()
}

func (way *Way) HasBicycleSupport() bool {
	return way.Highway == "cycleway" || way.Cycleway != "" ||
		way.CyclewayLeft != "" || way.CyclewayRight != "" ||
		way.Bicycle == "yes" || way.Bicycle == "designated" ||
		way.BicycleRoad == "yes"
}

func (way *Way) CanBeUsedByCars() bool {
	return !(way.Highway == "track" || way.Highway == "path" || way.Highway == "cycleway" ||
		way.Highway == "bridleway" || way.Highway == "footway" ||
		way.Tags["motor_vehicle"] == "no" || way.Tags["motor_vehicle"] == "forestry" ||
		way.Tags["motor_vehicle"] == "agricultural") &&
		!way.OnlyForWalkers()
}

func (way *Way) OnlyForWalkers() bool {
	return way.Highway == "steps" || way.Highway == "pedestrian"
}

func (way *Way) InConstruction() bool {
	return way.Highway == "construction" || way.Highway == "proposed"
}

func (way *Way) CheckTurnRestrictions(osm *OSMBinary, to_way *Way, via_node *Node) bool {
	// Only restrictions
	for _, restriction := range way.OnlyRestrictions {
		if via_node.ID != restriction.Via {
			continue // Nächste Restriction prüfen
		}

		if restriction.To != to_way.ID {
			return false
		}
	}

	for _, restriction := range way.NoRestrictions {
		if via_node.ID != restriction.Via {
			continue // Nächste Restriction prüfen
		}

		// NICHT zugelassen: from -> to
		if restriction.To == to_way.ID {
			return false
		}
	}

	// Wenn keine Restriction greift, ist der Move erlaubt
	return true
}
