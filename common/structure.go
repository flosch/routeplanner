package common

import (
	"sort"
)

type Tag struct {
	Key   string
	Value string
}

type Member struct {
	Type string
	Ref  int
	Role string
}

type Node struct {
	ID      int
	Version int16
	Lat     float64
	Lon     float64
	Tags    map[string]string

	// Computed values
	Access               string
	Neighbours           []int // oneway berücksichtigt
	UnfilteredNeighbours []int // oneway UNberücksichtigt (z. B. für Fußgänger usw)
	Ways                 []int
}

type Restriction struct {
	Via           int // eigentlich ein Punkt, dürfen laut Wiki aber auch mehrere Ways sein (TODO!!)
	To            int
	Location_hint int // Punkt als Hinweis für die Position des Schilds
	// TODO: Zeitliche Beschränkungen und Ausnahme von Fahrzeugen (s. Wiki)
}

type Way struct {
	ID               int
	Version          int16
	Nodes            []int
	Tags             map[string]string
	OnlyRestrictions []*Restriction
	NoRestrictions   []*Restriction

	// Computed values
	Oneway        int8
	Access        string
	Name          string
	Maxspeed      int
	Highway       string
	Cycleway      string
	CyclewayLeft  string
	CyclewayRight string
	Bicycle       string
	BicycleRoad   string
	Lanes         int
	Ref           string
	Tunnel        string // zwar nur "yes" oder "no", aber es kann ja auch unbestimmt sein
}

type Relation struct {
	ID      int
	Members []*Member // Reihenfolge wichtig
	Tags    map[string]string
}

type Bounds struct {
	MinLat float64
	MinLon float64
	MaxLat float64
	MaxLon float64
}

type OSMBinary struct {
	Version      float64
	Bounds       Bounds
	Nodes        NodeList     // []*Node
	Ways         WayList      // []*Way
	Relations    RelationList // []*Relation
	Restrictions []int        // Number of Relation-IDs which are restrictions

	// POI-Lookup

	// Geo-Lookup
	NodeLookupCenter *Node
	NodeLookup       map[int][]int // Distance (in 10-Meter-Schritten) ->
}

type RelationList []*Relation

func (rl RelationList) Len() int {
	return len(rl)
}

func (rl RelationList) Less(i, j int) bool {
	return rl[i].ID < rl[j].ID
}

func (rl RelationList) Swap(i, j int) {
	rl[i], rl[j] = rl[j], rl[i]
}

func (rl RelationList) Get(id int) *Relation {
	i := sort.Search(len(rl), func(x int) bool {
		return rl[x].ID >= id
	})
	if i < len(rl) && rl[i].ID == id {
		return rl[i]
	}
	return nil
}

func (rl RelationList) Sort() {
	sort.Sort(rl)
}

type WayList []*Way

func (wl WayList) Len() int {
	return len(wl)
}

func (wl WayList) Less(i, j int) bool {
	return wl[i].ID < wl[j].ID
}

func (wl WayList) Swap(i, j int) {
	wl[i], wl[j] = wl[j], wl[i]
}

func (wl WayList) Get(id int) *Way {
	i := sort.Search(len(wl), func(x int) bool {
		return wl[x].ID >= id
	})
	if i < len(wl) && wl[i].ID == id {
		return wl[i]
	}
	return nil
}

func (wl WayList) Sort() {
	sort.Sort(wl)
}

type NodeList []*Node

func (nl NodeList) Len() int {
	return len(nl)
}

func (nl NodeList) Less(i, j int) bool {
	return nl[i].ID < nl[j].ID
}

func (nl NodeList) Swap(i, j int) {
	nl[i], nl[j] = nl[j], nl[i]
}

func (nl NodeList) Get(id int) *Node {
	i := sort.Search(len(nl), func(x int) bool {
		return nl[x].ID >= id
	})
	if i < len(nl) && nl[i].ID == id {
		return nl[i]
	}
	return nil
}

func (nl NodeList) Sort() {
	sort.Sort(nl)
}
