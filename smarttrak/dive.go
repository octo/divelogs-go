// Package smarttrak implements parsing of SmartTrak's .asd files.
package smarttrak

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"time"
)

const timeOffset = 946684800 // 2000-01-01 01:00:00 +0100 CET

// WaterType denotes whether a dive happened in fresh or salt water.
type WaterType int

const (
	WaterType_Sweet WaterType = 1000
	WaterType_Salt  WaterType = 1025
)

// Density returns the water density in grams per Liter.
func (wt WaterType) Density() float64 {
	return float64(wt)
}

// Dive contains information about a single dive.
type Dive struct {
	DeviceID        uint32
	Sequence        int
	Time            time.Time
	Duration        time.Duration
	SurfaceInterval time.Duration
	TimeLimit       time.Duration
	WaterType       WaterType
	MaxDepth        float64
	AverageDepth    float64
	DepthLimit      float64
	AirTemperature  float64
	DecoTemperature float64
	MinTemperature  float64
	MaxTemperature  float64
	PressureStart   float64
	PressureEnd     float64
	TankWarning     float64
	TankReserve     float64
	Profile         []DataPoint
	PercentO2       int
	PercentHE       int

	// Unparsed
	WorkSensitivity uint16
	DesatBefore     uint16
	featureSet      uint32
	settings1       settings1
	settings2       uint32

	// Not certain
	maxPO2 float64

	// internal
	timeseriesSize uint16
}

// DataPoint holds timeseries data points.
type DataPoint struct {
	Time         time.Time
	Depth        float64
	Temperature  float64
	Alert        bool
	Warning      bool
	HighWorkload bool
	Bookmark     bool
}

// State returns a string representation of the DataPoint's state.
// Specifically alerts, warnings, and bookmarks.
func (d DataPoint) State() string {
	var ret string
	switch {
	case d.Alert:
		ret = "Alert"
	case d.Warning:
		ret = "Warning"
	case d.HighWorkload:
		ret = "High workload"
	default:
		ret = "No alert/warning"
	}
	if d.Bookmark {
		ret += ", Bookmark"
	}
	return ret
}

func ReadDive(r io.Reader) (*Dive, error) {
	data, err := readExact(r, 195)
	if err != nil {
		return nil, err
	}
	dive, err := ParseDive(data)
	if err != nil {
		return nil, err
	}

	data, err = readExact(r, int(dive.timeseriesSize))
	if err != nil {
		return nil, err
	}

	if err := dive.parseTimeseriesBlock(data); err != nil {
		return nil, err
	}

	_, err = readExact(r, 8)
	if err != nil {
		return nil, err
	}

	return dive, nil
}

// ParseDive parses the
func ParseDive(data []byte) (*Dive, error) {
	if got, want := len(data), 195; got != want {
		return nil, fmt.Errorf("unexpected size: got %d, want %d", got, want)
	}

	le := binary.LittleEndian

	s1 := settings1(le.Uint32(data[82:]))

	dive := &Dive{
		DeviceID:        le.Uint32(data[8:]),
		Sequence:        int(le.Uint16(data[28:])),
		Time:            parseTime(data[16:]),
		Duration:        parseDurationMin(le.Uint16(data[44:])),
		SurfaceInterval: parseDurationMin(le.Uint16(data[50:])),
		TimeLimit:       parseDurationMin(le.Uint16(data[33:])),
		featureSet:      le.Uint32(data[35:]),
		AirTemperature:  parseTemperature(le.Uint16(data[30:])),
		DecoTemperature: parseTemperature(le.Uint16(data[70:])),
		MinTemperature:  parseTemperature(le.Uint16(data[46:])),
		MaxTemperature:  parseTemperature(le.Uint16(data[160:])),
		PressureStart:   parsePressure(le.Uint16(data[54:])),
		PressureEnd:     parsePressure(le.Uint16(data[56:])),
		TankWarning:     parsePressure(le.Uint16(data[64:])),
		TankReserve:     parsePressure(le.Uint16(data[66:])),
		WaterType:       s1.WaterType(),
		MaxDepth:        parseDepth(le.Uint16(data[42:]), s1.WaterType()),
		AverageDepth:    parseDepth(le.Uint16(data[158:]), s1.WaterType()),
		DepthLimit:      parseDepth(le.Uint16(data[62:]), s1.WaterType()),
		// unparsed
		WorkSensitivity: le.Uint16(data[68:]),
		DesatBefore:     le.Uint16(data[72:]),
		settings1:       s1,
		settings2:       le.Uint32(data[167:]),
		// not certain
		maxPO2:         float64(le.Uint16(data[60:])) / 1000.0,
		timeseriesSize: le.Uint16(data[191:]),
	}

	return dive, nil
}

func (d *Dive) parseTimeseriesBlock(data []byte) error {
	for i := 0; i < len(data); i++ {
		b := data[i]

		switch {
		case b == 0xf0:
			log.Printf("0xF0 => %d", data[i+1])
			i += 1
		case b == 0xf2:
			log.Printf("0xF2 => %d", data[i+1])
			i += 1
		case b == 0xf3:
			log.Printf("0xF3 => %d", binary.LittleEndian.Uint16(data[i+1:]))
			i += 2
		case b == 0xf4:
			log.Printf("0xF4 => %d", binary.LittleEndian.Uint16(data[i+1:]))
			i += 2
		case b == 0xf8:
			log.Printf("0xF8 => %d", binary.LittleEndian.Uint32(data[i+1:]))
			i += 4
		case b == 0xf9:
			log.Printf("0xF9 => %d", data[i+1])
			i += 1
		case b == 0xfa:
			len, err := d.parseTimeseries(data[i+1:])
			if err != nil {
				return err
			}
			i += len
		case b == 0xfb:
			len := int(data[i+1])
			switch typ := int(data[i+2]); typ {
			case 32:
				pctO2 := binary.LittleEndian.Uint16(data[i+3:])
				pctHE := binary.LittleEndian.Uint16(data[i+5:])
				d.PercentO2 = int(pctO2)
				d.PercentHE = int(pctHE)
				log.Printf("Mixture: %d%% O₂, %d%% He", pctO2, pctHE)
				log.Printf("Maybe max pO₂: %d", binary.LittleEndian.Uint16(data[i+11:]))
			case 26:
				log.Printf("NOSTOPMIN:   %d", data[i+3])
				log.Printf("NOSTOPMBMIN: %d", data[i+4])
			default:
				log.Printf("Unknown block: type %d, payload %d bytes", typ, len-2)
			}
			i += len
		default:
			return fmt.Errorf("unexpected byte %#x at location %d", b, i)
		}
	}
	return nil
}

func (d *Dive) parseTimeseries(data []byte) (n int, err error) {
	const interval = 4 * time.Second

	state := DataPoint{
		Time: d.Time,
	}

	// Temperature is recorded relatively, i.e. in changes to the previous temperature.
	// Unfortunately, I have been unable to identify a "start temperature" in the binary data.
	// We record the up and down steps in the data, and later use the
	// MinTemperature and MaxTemperature fields to scale the fields correctly.
	var minTemp, maxTemp, currTemp int

	// ret contains the bytes that were read. Set when a 0xFB byte is read.
	var ret int

BYTE:
	for i := 0; i < len(data); i++ {
		b := data[i]

		switch {
		case b&0xf0 == 0xb0:
			raw := uint8(b & 0x0f)
			if raw&0x08 != 0 {
				raw |= 0xf0
			}
			currTemp += int(int8(raw))
			if minTemp > currTemp {
				minTemp = currTemp
			}
			if maxTemp < currTemp {
				maxTemp = currTemp
			}

			state.Temperature = float64(currTemp)
		case b&0xf0 == 0xc0:
			n := int(b & 0x0f)
			for i := 0; i < n; i++ {
				d.Profile = append(d.Profile, state)
				state.Time = state.Time.Add(interval)
			}
		case b&0xf0 == 0xe0:
			if b&0x02 != 0 {
				state.Alert = true
			}
			if b&0x01 != 0 {
				state.Warning = true
			}
			if b&0x04 != 0 {
				state.HighWorkload = true
			}
			if b&0x08 != 0 {
				state.Bookmark = true
			}
			if b&0x0f == 0 {
				state.Alert = false
				state.Warning = false
				state.HighWorkload = false
				state.Bookmark = false
			}
		case b == 0xfb:
			ret = i
			break BYTE
		case b&0x80 != 0:
			fmt.Printf("Unknown control byte: %#02X\n", b)
		default:
			diff := parseDepthDiff(b)

			state.Depth += diff
			d.Profile = append(d.Profile, state)
			state.Time = state.Time.Add(interval)
		}
	}

	// scale temperatures correctly
	if minTemp != maxTemp {
		fact := (d.MaxTemperature - d.MinTemperature) / float64(maxTemp-minTemp)
		offset := float64(minTemp)
		for i := range d.Profile {
			d.Profile[i].Temperature = d.MinTemperature + (fact * (d.Profile[i].Temperature - offset))
		}
	}

	return ret, nil
}

func parseTime(data []byte) time.Time {
	// t is half-seconds since 2000-01-01
	t := int64(binary.LittleEndian.Uint64(data[0:8]))
	t = timeOffset + t/2

	// offset is the time-zone offset in 15m increments.
	offset := int(binary.LittleEndian.Uint16(data[8:10]))
	offset *= 900 // 15 min

	return time.Unix(t, 0).In(time.FixedZone("Device/Local", offset))
}

func parseDurationMin(d uint16) time.Duration {
	return time.Duration(d) * time.Minute
}

func parseTemperature(t uint16) float64 {
	return float64(int16(t)) / 10.0
}

func parsePressure(p uint16) float64 {
	return float64(p) / 128.0
}

func parseDepth(depth uint16, wt WaterType) float64 {
	return 10.0 * float64(depth) / wt.Density()
}

func parseDepthDiff(d byte) float64 {
	// copy bit 7 to bit 8 so that when we cast to a signed int,
	// the signedness is interpreted correctly.
	d |= ((d & 0x40) << 1)
	return float64(int8(d)) / 50.0
}

type settings1 uint32

func (s settings1) WaterType() WaterType {
	const isSaltWater = 0x00100000
	if s&isSaltWater != 0 {
		return WaterType_Salt
	}
	return WaterType_Sweet
}
