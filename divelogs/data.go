package divelogs

import (
	"encoding/xml"
	"math"
	"time"
)

// Data implements the data structure exported by https://divelogs.de/
type Data struct {
	ID                  int
	DiveNumber          int
	Time                time.Time
	DiveDuration        time.Duration
	SurfaceDuration     time.Duration
	MaxDepth            float64
	MeanDepth           float64
	Location            string
	Site                string
	Weather             string
	Visibility          string
	AirTemperature      float64
	MaxDepthTemperature float64
	DiveEndTemperature  float64
	Partner             string
	Boat                string
	Cylinder            Cylinder
	Weight              float64
	O2Percent           float64
	HEPercent           float64
	LogNotes            string
	Latitude            float64
	Longitude           float64
	ZoomLevel           int
	SampleInterval      time.Duration
	Samples             []Sample
}

// Sample is a single datapoint in the dive's timeseries data.
type Sample struct {
	Depth float64 `xml:"DEPTH"`
}

// Cylinder contains information about the cylinders used.
type Cylinder struct {
	Name            string
	Description     string
	Doubles         bool
	Size            float64
	StartPressure   float64
	EndPressure     float64
	WorkingPressure float64
}

// MarshalXML implements the xml.Marshaler interface.
func (d Data) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{
		Local: "DIVELOGSDATA",
	}

	ephemeral := data{
		ID:                  d.ID,
		DiveNumber:          d.DiveNumber,
		Date:                d.Time.Format("02.01.2006"),
		Time:                d.Time.Format("15:04:05"),
		DiveTimeSec:         int(math.Round(d.DiveDuration.Seconds())),
		SurfaceDuration:     int(math.Round(float64(d.SurfaceDuration.Seconds()))),
		MaxDepth:            d.MaxDepth,
		MeanDepth:           d.MeanDepth,
		Location:            cdataString(d.Location),
		Site:                cdataString(d.Site),
		Weather:             cdataString(d.Weather),
		Visibility:          cdataString(d.Visibility),
		AirTemperature:      d.AirTemperature,
		MaxDepthTemperature: d.MaxDepthTemperature,
		DiveEndTemperature:  d.DiveEndTemperature,
		Partner:             cdataString(d.Partner),
		Boat:                cdataString(d.Boat),
		CylinderName:        cdataString(d.Cylinder.Name),
		CylinderDescription: cdataString(d.Cylinder.Description),
		CylinderDoubles: func(b bool) int {
			if b {
				return 1
			}
			return 0
		}(d.Cylinder.Doubles),
		CylinderSize:            d.Cylinder.Size,
		CylinderStartPressure:   d.Cylinder.StartPressure,
		CylinderEndPressure:     d.Cylinder.EndPressure,
		CylinderWorkingPressure: d.Cylinder.WorkingPressure,
		Weight:                  d.Weight,
		O2Percent:               d.O2Percent,
		HEPercent:               d.HEPercent,
		LogNotes:                cdataString(d.LogNotes),
		Latitude:                d.Latitude,
		Longitude:               d.Longitude,
		ZoomLevel:               d.ZoomLevel,
		SampleIntervalSec:       int(math.Round(d.SampleInterval.Seconds())),
		Samples:                 d.Samples,
	}

	return enc.EncodeElement(ephemeral, start)
}

// UnmarshalXML implements the xml.Unmarshaler interface.
func (d *Data) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	var ephemeral data
	if err := dec.DecodeElement(&ephemeral, &start); err != nil {
		return err
	}

	*d = Data{
		ID:                  ephemeral.ID,
		DiveNumber:          ephemeral.DiveNumber,
		DiveDuration:        time.Duration(ephemeral.DiveTimeSec) * time.Second,
		SurfaceDuration:     time.Duration(ephemeral.SurfaceDuration) * time.Second,
		MaxDepth:            ephemeral.MaxDepth,
		MeanDepth:           ephemeral.MeanDepth,
		Location:            string(ephemeral.Location),
		Site:                string(ephemeral.Site),
		Weather:             string(ephemeral.Weather),
		Visibility:          string(ephemeral.Visibility),
		AirTemperature:      ephemeral.AirTemperature,
		MaxDepthTemperature: ephemeral.MaxDepthTemperature,
		DiveEndTemperature:  ephemeral.DiveEndTemperature,
		Partner:             string(ephemeral.Partner),
		Boat:                string(ephemeral.Boat),
		Cylinder: Cylinder{
			Name:            string(ephemeral.CylinderName),
			Description:     string(ephemeral.CylinderDescription),
			Size:            ephemeral.CylinderSize,
			StartPressure:   ephemeral.CylinderStartPressure,
			EndPressure:     ephemeral.CylinderEndPressure,
			WorkingPressure: ephemeral.CylinderWorkingPressure,
		},
		Weight:         ephemeral.Weight,
		O2Percent:      ephemeral.O2Percent,
		HEPercent:      ephemeral.HEPercent,
		LogNotes:       string(ephemeral.LogNotes),
		Latitude:       ephemeral.Latitude,
		Longitude:      ephemeral.Longitude,
		ZoomLevel:      ephemeral.ZoomLevel,
		SampleInterval: time.Duration(ephemeral.SampleIntervalSec) * time.Second,
		Samples:        ephemeral.Samples,
	}

	t, err := time.ParseInLocation("02.01.2006 15:04:05", ephemeral.Date+" "+ephemeral.Time, time.Local)
	if err != nil {
		return err
	}
	d.Time = t

	d.Cylinder.Doubles = ephemeral.CylinderDoubles != 0

	return nil
}

// data is an internal version of Data used for XML [un]marshalling
type data struct {
	XMLName                 struct{}    `xml:"DIVELOGSDATA"`
	ID                      int         `xml:"DIVELOGSID"`
	DiveNumber              int         `xml:"DIVELOGSDIVENUMBER"`
	Date                    string      `xml:"DATE"`
	Time                    string      `xml:"TIME"`
	DiveTimeSec             int         `xml:"DIVETIMESEC"`
	SurfaceDuration         int         `xml:"SURFACETIME"`
	MaxDepth                float64     `xml:"MAXDEPTH"`
	MeanDepth               float64     `xml:"MEANDEPTH"`
	Location                cdataString `xml:"LOCATION,omitempty"`
	Site                    cdataString `xml:"SITE,omitempty"`
	Weather                 cdataString `xml:"WEATHER,omitempty"`
	Visibility              cdataString `xml:"WATERVIZIBILITY,omitempty"` // sic
	AirTemperature          float64     `xml:"AIRTEMP"`
	MaxDepthTemperature     float64     `xml:"WATERTEMPMAXDEPTH"`
	DiveEndTemperature      float64     `xml:"WATERTEMPATEND"`
	Partner                 cdataString `xml:"PARTNER,omitempty"`
	Boat                    cdataString `xml:"BOATNAME,omitempty"`
	CylinderName            cdataString `xml:"CYLINDERNAME,omitempty"`
	CylinderDescription     cdataString `xml:"CYLINDERDESCRIPTION,omitempty"`
	CylinderDoubles         int         `xml:"DBLTANK"`
	CylinderSize            float64     `xml:"CYLINDERSIZE"`
	CylinderStartPressure   float64     `xml:"CYLINDERSTARTPRESSURE"`
	CylinderEndPressure     float64     `xml:"CYLINDERENDPRESSURE"`
	CylinderWorkingPressure float64     `xml:"WORKINGPRESSURE"`
	Weight                  float64     `xml:"WEIGHT"`
	O2Percent               float64     `xml:"O2PCT"`
	HEPercent               float64     `xml:"HEPCT"`
	LogNotes                cdataString `xml:"LOGNOTES,omitempty"`
	Latitude                float64     `xml:"LAT,omitempty"`
	Longitude               float64     `xml:"LNG,omitempty"`
	ZoomLevel               int         `xml:"GOOGLEMAPSZOOMLEVEL"`
	SampleIntervalSec       int         `xml:"SAMPLEINTERVAL"`
	Samples                 []Sample    `xml:"SAMPLE"`
}

type cdataString string

func (s cdataString) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	wrapped := struct {
		CData string `xml:",cdata"`
	}{string(s)}

	return enc.EncodeElement(wrapped, start)
}
