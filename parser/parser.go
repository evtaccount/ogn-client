package parser

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// PositionMessage represents a decoded position report.
type PositionMessage struct {
	Raw       string
	Callsign  string
	Timestamp time.Time
	Latitude  float64
	Longitude float64
	Altitude  float64 // meters
}

var rePosition = regexp.MustCompile(`^(?P<callsign>[A-Z0-9-]+)>[^:]+:/?(?P<ts>\d{6})h(?P<latdeg>\d{2})(?P<latmin>\d{2}\.\d{2})(?P<lathem>[NS])/(?P<londeg>\d{3})(?P<lonmin>\d{2}\.\d{2})(?P<lonhem>[EW]).*?(?P<course>\d{3})/(?P<speed>\d{3})/A=(?P<alt>\d+)`)

// Parse attempts to parse an APRS/OGN position message. Only a subset of fields
// is extracted. An error is returned if the message does not match the expected
// pattern.
func Parse(raw string) (*PositionMessage, error) {
	m := rePosition.FindStringSubmatch(strings.TrimSpace(raw))
	if m == nil {
		return nil, errors.New("unrecognized message format")
	}

	get := func(name string) string {
		for i, n := range rePosition.SubexpNames() {
			if n == name {
				return m[i]
			}
		}
		return ""
	}

	latDeg, _ := strconv.Atoi(get("latdeg"))
	latMin, _ := strconv.ParseFloat(get("latmin"), 64)
	lonDeg, _ := strconv.Atoi(get("londeg"))
	lonMin, _ := strconv.ParseFloat(get("lonmin"), 64)
	lat := float64(latDeg) + latMin/60.0
	lon := float64(lonDeg) + lonMin/60.0
	if get("lathem") == "S" {
		lat = -lat
	}
	if get("lonhem") == "W" {
		lon = -lon
	}

	altFeet, _ := strconv.Atoi(get("alt"))
	altitude := float64(altFeet) * 0.3048

	ts := get("ts")
	t, _ := time.Parse("150405", ts)
	now := time.Now().UTC()
	timestamp := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.UTC)

	return &PositionMessage{
		Raw:       raw,
		Callsign:  get("callsign"),
		Timestamp: timestamp,
		Latitude:  lat,
		Longitude: lon,
		Altitude:  altitude,
	}, nil
}

func (p *PositionMessage) String() string {
	return fmt.Sprintf("%s %.5f %.5f %.0fm", p.Callsign, p.Latitude, p.Longitude, p.Altitude)
}
