package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/octo/divelogs-go/smarttrak"
)

var flagInput = flag.String("input", "", "path to input file")

func main() {
	flag.Parse()

	if *flagInput == "" {
		flag.Usage()
	}

	data, err := ioutil.ReadFile(*flagInput)
	if err != nil {
		log.Fatal(err)
	}

	delim := []byte{0x61, 0x2c, 0x69, 0x04}
	var offsets []int

	for i := 0; i < len(data)-len(delim); i++ {
		if !bytes.Equal(data[i:i+len(delim)], delim) {
			continue
		}
		offsets = append(offsets, i-8)
	}

	var diveData [][]byte
	for i := 1; i < len(offsets); i++ {
		firstByte := offsets[i]
		lastByte := len(data)
		if i < len(offsets)-1 {
			lastByte = offsets[i+1]
		}

		diveData = append(diveData, data[firstByte:lastByte])
	}

	for i, data := range diveData {
		fmt.Printf("=== Dive #%d, starting at %#04x ===\n", i+1, offsets[i+1])

		dive, err := smarttrak.ParseDive(data)
		if err != nil {
			log.Println("error parsing dive:", err)
			continue
		}

		fmt.Printf("Main info: Date: %s; Sequence: %d; Duration: %s;\n",
			dive.Time, dive.Sequence, dive.Duration)
		fmt.Printf("Temperatures: Min: %.1f; Max: %.1f; Deco: %.1f; Air: %.1f;\n",
			dive.MinTemperature, dive.MaxTemperature, dive.DecoTemperature, dive.AirTemperature)
		fmt.Printf("Depths: Average: %.1f; Max: %.1f;\n",
			dive.AverageDepth, dive.MaxDepth)
	}
}
