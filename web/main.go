package main

import (
	"encoding/xml"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/octo/divelogs-go/divelogs"
	"github.com/octo/divelogs-go/smarttrak"
)

func main() {
	srv := newServer()

	http.HandleFunc("/", srv.Index)
	http.HandleFunc("/asd", srv.ASD)
	http.HandleFunc("/divelogs", srv.Divelogs)

	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

type server struct {
	templates *template.Template
}

func newServer() *server {
	s := &server{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}

	return s
}

func (s server) Index(w http.ResponseWriter, _ *http.Request) {
	s.templates.ExecuteTemplate(w, "index.html", nil)
}

func (s server) ASD(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.ASDCreate(w, r)
		return
	}

	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (s server) ASDCreate(w http.ResponseWriter, r *http.Request) {
	const maxFileSize = 1 << 20
	if err := r.ParseMultipartForm(maxFileSize); err != nil {
		log.Println("ParseMultipartForm:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("data")
	if err != nil {
		log.Println("FormFile:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := smarttrak.ReadHeader(file); err != nil {
		log.Println("smarttrack.ReadHeader:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dive, err := smarttrak.ReadDive(file)
	if err != nil {
		log.Println("smarttrack.ReadDive:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.templates.ExecuteTemplate(w, "dive.html", dive); err != nil {
		log.Println("ExecuteTemplate:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s server) Divelogs(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.DivelogsPost(w, r)
		return
	}

	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (s server) DivelogsPost(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if _, err := smarttrak.ReadHeader(r.Body); err != nil {
		log.Println("smarttrack.ReadHeader:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dive, err := smarttrak.ReadDive(r.Body)
	if err != nil {
		log.Println("smarttrack.ReadDive:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	d := divelogs.Data{
		ID:                  0,
		DiveNumber:          0,
		Time:                dive.Time,
		DiveDuration:        dive.Duration,
		SurfaceDuration:     dive.SurfaceInterval,
		MaxDepth:            dive.MaxDepth,
		MeanDepth:           dive.AverageDepth,
		Location:            "",
		Site:                "",
		Weather:             "",
		Visibility:          "",
		AirTemperature:      dive.AirTemperature,
		MaxDepthTemperature: dive.MinTemperature,
		DiveEndTemperature:  dive.DecoTemperature,
		Partner:             "",
		Boat:                "",
		Cylinder:            divelogs.Cylinder{},
		Weight:              0,
		O2Percent:           float64(dive.PercentO2),
		HEPercent:           float64(dive.PercentHE),
		LogNotes:            "",
		Latitude:            0,
		Longitude:           0,
		ZoomLevel:           0,
		SampleInterval:      4 * time.Second,
	}

	for _, p := range dive.Profile {
		d.Samples = append(d.Samples, divelogs.Sample{
			Depth: p.Depth,
		})
	}

	w.Header().Set("Content-Type", "text/xml")
	if err := xml.NewEncoder(w).Encode(d); err != nil {
		log.Println("xml.Encoder.Encode:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
