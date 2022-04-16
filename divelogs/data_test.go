package divelogs

import (
	"encoding/xml"
	"io/ioutil"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestUnmarshal(t *testing.T) {
	testdata, err := ioutil.ReadFile("testdata/data.xml")
	if err != nil {
		t.Fatal(err)
	}

	var got Data
	if err := xml.Unmarshal(testdata, &got); err != nil {
		t.Fatal(err)
	}

	want := Data{
		ID:                  3355222,
		DiveNumber:          12,
		Time:                time.Date(2021, time.October, 17, 11, 15, 15, 0, time.Local),
		DiveDuration:        30*time.Minute + 32*time.Second,
		SurfaceDuration:     0,
		MaxDepth:            20.9,
		MeanDepth:           7.5,
		Location:            "Murner See",
		Site:                "Turm",
		Weather:             "-",
		Visibility:          "4/4",
		AirTemperature:      11.4,
		MaxDepthTemperature: 8.8,
		DiveEndTemperature:  0,
		Partner:             "",
		Boat:                "",
		Cylinder: Cylinder{
			StartPressure: 200.0,
			EndPressure:   50.0,
		},
		Weight:         0,
		O2Percent:      21,
		HEPercent:      0,
		LogNotes:       "",
		Latitude:       49.353699,
		Longitude:      12.201113,
		ZoomLevel:      12,
		SampleInterval: 4 * time.Second,
		Samples: []Sample{
			{0},
			{1.19},
			{1.37},
			{0.02},
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("xml.Unmarshal: results differ (-want/+got):\n%s", diff)
	}
}

func TestMarshal(t *testing.T) {
	testdata, err := ioutil.ReadFile("testdata/data.xml")
	if err != nil {
		t.Fatal(err)
	}

	var want Data
	if err := xml.Unmarshal(testdata, &want); err != nil {
		t.Fatal(err)
	}

	data, err := xml.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}

	var got Data
	if err := xml.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("xml.Unmarshal: results differ (-want/+got):\n%s", diff)
	}
}
