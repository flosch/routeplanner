package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"net/http"
	"time"

	"private/routenplaner/src/src/common"
)

var osm *common.OSMBinary

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	var port int
	flag.IntVar(&port, "port", 80, "http port")
	file_logging := flag.Bool("log", false, "log to file")
	flag.Parse()

	var stats runtime.MemStats

	fmt.Printf("Routing-Engine (running on port %d)\n(C)opyright 2014 Florian Schlachter\n\n", port)

	if *file_logging {
		// Start logging
		flog, err := os.Create(fmt.Sprintf("requests.%d.log", time.Now().Unix()))
		if err != nil {
			panic(err)
		}
		defer flog.Close()
		log.SetOutput(flog)
	}

	f, err := os.Open(flag.Arg(0))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	stime := time.Now()

	err = gob.NewDecoder(f).Decode(&osm)
	if err != nil {
		panic(err)
	}

	/*
		osm.NodeMap = make(map[int]*common.Node)

		for _, n := range osm.Nodes {
			osm.NodeMap[n.ID] = n
		}
	*/

	fmt.Printf("Load time: %s\n", time.Since(stime))
	runtime.ReadMemStats(&stats)
	fmt.Printf("Mem usage: %d MiB\n\n", stats.Alloc/1024/1024)

	/*
		fprof, err := os.Create("cpu.profile")
		if err != nil {
			panic(err)
		}
		defer fprof.Close()
		err = pprof.StartCPUProfile(fprof)
		if err != nil {
			panic(err)
		}
		stime = time.Now()
	*/

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
