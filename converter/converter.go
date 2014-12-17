package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	filename_in := flag.String("in", "", "OSM-XML input file")
	filename_out := flag.String("out", "", "binary output file")
	flag.Parse()

	if *filename_in == "" || *filename_out == "" {
		fmt.Printf("In/out filenames must be given.\n")
		os.Exit(1)
		return
	}

	fmt.Printf("Converter\n%s -> %s\n\n", *filename_in, *filename_out)

	stime := time.Now()
	defer func() {
		etime := time.Since(stime)
		fmt.Printf("\n\nCompleted in %s.\n", etime)
	}()

	// Load OSM file
	// TODO: Later add a lazy XML-reader (to not have all the data at once in the memory)
	xmlData, err := loadXML(*filename_in)
	if err != nil {
		log.Fatal(err)
	}

	// Convert into own representation
	osmBinary, err := convert(xmlData)
	if err != nil {
		log.Fatal(err)
	}

	// Save as binary to disk
	err = saveBinary(*filename_out, osmBinary)
	if err != nil {
		log.Fatal(err)
	}
}
