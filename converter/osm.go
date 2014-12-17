package main

import (
	"encoding/xml"
	"os"
)

type XMLTag struct {
	Key   string `xml:"k,attr"`
	Value string `xml:"v,attr"`
}

type XMLMember struct {
	Type string `xml:"type,attr"`
	Ref  int    `xml:"ref,attr"`
	Role string `xml:"role,attr"`
}

type XMLNode struct {
	ID      int       `xml:"id,attr"`
	Version int16     `xml:"version,attr"`
	Lat     float64   `xml:"lat,attr"`
	Lon     float64   `xml:"lon,attr"`
	Tags    []*XMLTag `xml:"tag"`
}

type XMLWay struct {
	ID      int   `xml:"id,attr"`
	Version int16 `xml:"version,attr"`
	Nodes   []struct {
		Ref int `xml:"ref,attr"`
	} `xml:"nd"`
	Tags []*XMLTag `xml:"tag"`
}

type XMLRelation struct {
	ID      int          `xml:"id,attr"`
	Members []*XMLMember `xml:"member"`
	Tags    []*XMLTag    `xml:"tag"`
}

type XMLBounds struct {
	MinLat float64 `xml:"minlat,attr"`
	MinLon float64 `xml:"minlon,attr"`
	MaxLat float64 `xml:"maxlat,attr"`
	MaxLon float64 `xml:"maxlon,attr"`
}

type XMLOSM struct {
	XMLName xml.Name `xml:"osm"`

	Version   float64        `xml:"version,attr"`
	Bounds    XMLBounds      `xml:"bounds"`
	Nodes     []*XMLNode     `xml:"node"`
	Ways      []*XMLWay      `xml:"way"`
	Relations []*XMLRelation `xml:"relation"`
}

func loadXML(filename string) (*XMLOSM, error) {
	var xmlData *XMLOSM

	fin, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fin.Close()

	err = xml.NewDecoder(fin).Decode(&xmlData)
	if err != nil {
		return nil, err
	}

	return xmlData, nil
}
