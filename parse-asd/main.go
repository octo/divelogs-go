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

	r := bytes.NewReader(data)
	hdr, err := smarttrak.ReadHeader(r)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Logbook name: %q\n", hdr.Name)

	dv, err := smarttrak.ReadDive(r)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("DeviceID:    %#08x\n", dv.DeviceID)
	fmt.Printf("Time:        %v\n", dv.Time)
	fmt.Printf("Sequence:    %d\n", dv.Sequence)
	fmt.Printf("Air temp:    %.1f\n", dv.AirTemperature)
	fmt.Printf("Time limit:  %v\n", dv.TimeLimit)
	fmt.Printf("Max depth:   %.1f\n", dv.MaxDepth)
	fmt.Printf("Duration:    %v\n", dv.Duration)
	fmt.Printf("Min temp:    %.1f\n", dv.MinTemperature)
	fmt.Printf("Surface Int: %v\n", dv.SurfaceInterval)
	fmt.Printf("Pres. Start: %.1f\n", dv.PressureStart)
	fmt.Printf("Pres. End:   %.1f\n", dv.PressureEnd)

	fmt.Printf("Main info: Date: %s; Sequence: %d; Duration: %s;\n",
		dv.Time, dv.Sequence, dv.Duration)
	fmt.Printf("Temperatures: Min: %.1f; Max: %.1f; Deco: %.1f; Air: %.1f;\n",
		dv.MinTemperature, dv.MaxTemperature, dv.DecoTemperature, dv.AirTemperature)
	fmt.Printf("Depths: Average: %.1f; Max: %.1f;\n",
		dv.AverageDepth, dv.MaxDepth)

	fmt.Printf("Calculated start temp: %.1f\n", dv.Profile[0].Temperature)
}
