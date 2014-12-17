package main

import (
	"errors"
	"fmt"
	"time"

	"private/routenplaner/src/src/common"
)

type CostProfileFunc func(hints *HintMgr, prev, via, next *common.Node, prev_way, next_way *common.Way, request *RouteRequest) (costs int, allowed bool)

type RouteRequest struct {
	Departure   *common.Node
	Via         []*common.Node
	Destination *common.Node

	// Constraints
	TimeDeparture  time.Time
	TimeArrival    time.Time
	MaxTravelCosts int // in Euros

	// Profile being used
	CostProfileText string
	costProfile     CostProfileFunc

	// Internal bookcounting and calculations
	calcuation_start time.Time
	osm              *common.OSMBinary
	g_values         map[int]int
	prio_items       map[int]*prioItem
	openlist         *prioQueue
	closedlist       map[int]bool
}

func makeRouteRequest(osm *common.OSMBinary, from, to *common.Node, profile string) (*RouteRequest, error) {
	if osm == nil {
		panic("osm == nil")
	}
	if from == nil {
		return nil, errors.New("From-node must be given")
	}
	if to == nil {
		return nil, errors.New("To-node must be given")
	}
	if profile == "" {
		return nil, errors.New("Profile not given")
	}

	var cost_func CostProfileFunc
	switch profile {
	case "bike":
		cost_func = profile_bike
	case "car":
		cost_func = profile_car
	default:
		return nil, errors.New(fmt.Sprintf("Profile '%s' not supported yet", profile))
	}

	return &RouteRequest{
		Departure:       from,
		Destination:     to,
		CostProfileText: profile,
		costProfile:     cost_func,

		osm:        osm,
		openlist:   newPrioQueue(),
		closedlist: make(map[int]bool),
		g_values:   make(map[int]int),
		prio_items: make(map[int]*prioItem),
	}, nil
}
