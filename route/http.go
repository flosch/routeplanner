package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/flosch/pongo"
	"github.com/gorilla/mux"
)

func requestHandler(w http.ResponseWriter, r *http.Request) {
	var tplResultPage = pongo.Must(pongo.FromFile("gui/map.html", nil))
	err := tplResultPage.ExecuteRW(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type jsonLocation struct {
	ID        int
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
}

type jsonHint struct {
	Hint     string
	Location jsonLocation
}

type jsonStepMeta struct {
	Maxspeed       int    `json:",omitempty"` // in lokaler Einheit, z. B. km/h
	Tunnel         string `json:",omitempty"` // in/out oder leer
	BicycleSupport int    // 0=nein, 1=ja, 2=teilweise, je nach dem, ob der Step Unterstützung für Räder bietet
	Motorway       string // in/out oder leer

}

type jsonStep struct {
	Direction   string
	Instruction string
	Way         string
	Distance    int
	Meta        jsonStepMeta
	Hints       []jsonHint
	Nodes       []jsonLocation
}

type jsonResponse struct {
	CopyrightNotice string
	CalcDuration    int `json:",omitempty"` // in mikrosekunden
	Success         bool
	Profile         string
	Lang            string
	Error           string `json:",omitempty"`

	DepartureNode   jsonLocation
	DestinationNode jsonLocation

	Meta struct {
		Distance int // in meters
		Time     struct {
			Duration struct {
				Minutes int // in minutes
				Note    string
			}
			EstimatedDeparture time.Time
			EstimatedArrival   time.Time
		}
		BicycleSupportPercentage int // 0 - 100% in der Strecke bzgl. Fahrradkompatibilität
	}

	HumanReadable struct {
		From  string
		To    string
		Steps []jsonStep
	}

	Nodes []jsonLocation `json:",omitempty"`
}

func routeHandler(w http.ResponseWriter, r *http.Request) {
	departure_id, err := strconv.Atoi(r.FormValue("departure"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	destination_id, err := strconv.Atoi(r.FormValue("destination"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ref := strings.TrimSpace(r.FormValue("ref"))
	if ref == "" {
		http.Error(w, "Bitte einen 'ref' liefern", http.StatusBadRequest)
		return
	}
	if len(ref) > 50 {
		http.Error(w, "'ref' darf maximal 50 Zeichen lang sein", http.StatusBadRequest)
		return
	}

	response := jsonResponse{
		CopyrightNotice: "Zugrundeliegende Geoinformationen © OpenStreetMap-Mitwirkende (lizenziert unter ODbL), Routing-Engine © Florian Schlachter",
		Lang:            "de", // TODO
	}

	defer func() {
		log.Printf("[Route, ref=%s, success=%t, error=%s, dur=%d] From=%d (%s) To=%d (%s)\n", ref,
			response.Success, response.Error,
			response.CalcDuration,
			response.DepartureNode.ID, response.HumanReadable.From,
			response.DestinationNode.ID, response.HumanReadable.To)

		buf, err := json.MarshalIndent(response, "", "    ")
		if err != nil {
			panic(err)
		}
		w.Write(buf)
	}()

	// Profilermittlung
	response.Profile = r.FormValue("profile")

	// Messe Zeit der Berechnung
	stime := time.Now()

	departure_node := osm.Nodes.Get(departure_id)
	destination_node := osm.Nodes.Get(destination_id)

	if departure_node == nil {
		response.Error = "Departure node not found"
		return
	}

	if destination_node == nil {
		response.Error = "Destination node not found"
		return
	}

	// Erstelle Anfrage
	request, err := makeRouteRequest(osm,
		departure_node,   // Von
		destination_node, // Nach
		response.Profile, // mit dem Fahrrad
	)

	// 612796806 = S Lichtenrade, 175207143 = ganz im norden berlins
	if err != nil {
		response.Error = err.Error()
		return
	}

	// Berechne Route
	route, err := request.calculate()

	response.CalcDuration = int(time.Since(stime).Nanoseconds() / 1000)

	if err != nil {
		response.Error = err.Error()
		return
	}

	if route != nil {
		response.Success = true

		response.DepartureNode = jsonLocation{
			ID:        departure_node.ID,
			Latitude:  departure_node.Lat,
			Longitude: departure_node.Lon,
		}

		response.DestinationNode = jsonLocation{
			ID:        destination_node.ID,
			Latitude:  destination_node.Lat,
			Longitude: destination_node.Lon,
		}

		response.HumanReadable.From = route.DepartureText()
		response.HumanReadable.To = route.DestinationText()
		response.Meta.Distance = route.Distance()
		travelTime, travelTimeNote := route.TravelTime()
		response.Meta.Time.Duration.Minutes = travelTime
		response.Meta.Time.Duration.Note = travelTimeNote
		response.Meta.Time.EstimatedDeparture = time.Now().Add(10 * time.Minute)
		response.Meta.Time.EstimatedArrival = response.Meta.Time.EstimatedDeparture.Add(time.Duration(travelTime) * time.Minute)

		if response.Profile == "bike" {
			response.Meta.BicycleSupportPercentage = route.BicycleSupportPercentage()
		}

		for _, step := range route.steps() {
			jstep := jsonStep{
				Direction:   step.Direction(),
				Instruction: step.Text("de"),
				Way:         step.ways[0].Fullname(),
				Distance:    step.Distance(),
				Meta: jsonStepMeta{
					Maxspeed:       step.Maxspeed,
					Tunnel:         step.Tunnel,
					BicycleSupport: step.ways.HasBicycleSupport(),
					Motorway:       step.Motorway,
				},
			}

			// Hints
			hints := step.Hints("de")
			for _, hint := range hints {
				jstep.Hints = append(jstep.Hints, jsonHint{
					Hint: hint.Text,
					Location: jsonLocation{
						ID:        hint.Node.ID,
						Latitude:  hint.Node.Lat,
						Longitude: hint.Node.Lon,
					},
				})
			}

			// Nodes
			for _, node := range step.nodes {
				jstep.Nodes = append(jstep.Nodes, jsonLocation{
					ID:        node.ID,
					Latitude:  node.Lat,
					Longitude: node.Lon,
				})
			}

			response.HumanReadable.Steps = append(
				response.HumanReadable.Steps,
				jstep)
		}

		// Nodes
		for _, node := range route.Nodes {
			response.Nodes = append(response.Nodes, jsonLocation{
				ID:        node.Node.ID,
				Latitude:  node.Node.Lat,
				Longitude: node.Node.Lon,
			})
		}
	}
}

func discoverHandler(w http.ResponseWriter, r *http.Request) {
	lat, err := strconv.ParseFloat(r.FormValue("lat"), 64)
	if err != nil {
		http.Error(w, "Latitude is no float", http.StatusBadRequest)
		return
	}
	lon, err := strconv.ParseFloat(r.FormValue("lon"), 64)
	if err != nil {
		http.Error(w, "Longitude is no float", http.StatusBadRequest)
		return
	}

	ref := strings.TrimSpace(r.FormValue("ref"))
	if ref == "" {
		http.Error(w, "Bitte einen 'ref' liefern", http.StatusBadRequest)
		return
	}
	if len(ref) > 50 {
		http.Error(w, "'ref' darf maximal 50 Zeichen lang sein", http.StatusBadRequest)
		return
	}

	var response struct {
		CopyrightNotice string
		Found           bool
		CalcDuration    int64 // in mikrosekunden
		Location        jsonLocation
		Meta            struct {
			Name string
		}
	}

	response.CopyrightNotice = "Zugrundeliegende Geoinformationen © OpenStreetMap-Mitwirkende (lizenziert unter ODbL), Routing-Engine © Florian Schlachter"

	profile := r.FormValue("profile")

	stime := time.Now()
	node := discover(lat, lon, profile)
	response.CalcDuration = time.Since(stime).Nanoseconds() / 1000

	var resid int // for logging
	if node != nil {
		resid = node.ID
		response.Found = true
		response.Location = jsonLocation{
			ID:        node.ID,
			Latitude:  node.Lat,
			Longitude: node.Lon,
		}
		name := node.WayText(osm)
		if name == "" {
			name = node.Tags["name"]
		}
		if name == "" {
			name = fmt.Sprintf("Weg (ID %d)", node.ID)
		}
		response.Meta.Name = name
	}
	buf, err := json.MarshalIndent(response, "", "    ")
	if err != nil {
		panic(err)
	}
	w.Write(buf)

	log.Printf("[Discover, ref=%s, found=%t, dur=%d] Profile=%s ReqLat=%f ReqLon=%f ResID=%d\n",
		ref,
		response.Found, response.CalcDuration,
		profile, lat, lon, resid)
}

func apiDocHandler(w http.ResponseWriter, r *http.Request) {
	var tplResultPage = pongo.Must(pongo.FromFile("gui/api.html", nil))
	err := tplResultPage.ExecuteRW(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func appHandler(w http.ResponseWriter, r *http.Request) {
	var tplResultPage = pongo.Must(pongo.FromFile("gui/app.html", nil))
	err := tplResultPage.ExecuteRW(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func init() {
	r := mux.NewRouter()
	r.HandleFunc("/", requestHandler)
	r.HandleFunc("/route", routeHandler)
	r.HandleFunc("/discover", discoverHandler)
	r.HandleFunc("/api", apiDocHandler)
	r.HandleFunc("/app", appHandler)
	http.Handle("/static/", http.FileServer(http.Dir("gui/")))
	http.Handle("/", r)
}
