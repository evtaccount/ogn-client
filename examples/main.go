// Example: connect to OGN and print parsed positions within 100 km of Munich.
package main

import (
	"fmt"
	"log"
	"os"

	"ogn/client"
	"ogn/parser"
)

func main() {
	// Filter: 100 km radius around Munich (48.14, 11.58).
	filter := client.RangeFilter(48.14, 11.58, 100)

	c := client.New("example", filter)
	c.Logger = log.New(os.Stderr, "[ogn] ", log.LstdFlags)

	fmt.Printf("Connecting with filter: %s\n", filter)

	err := c.Run(func(raw string) {
		msg, err := parser.Parse(raw)
		if err != nil {
			return // skip unparseable lines
		}

		switch m := msg.(type) {
		case *parser.PositionMessage:
			fmt.Printf("%-12s %8.4f %8.4f %6.0fm %5.1fkm/h %+5.1fm/s  addr=%s type=%d\n",
				m.Callsign, m.Latitude, m.Longitude, m.Altitude,
				m.GroundSpeed, m.ClimbRate, m.Address, m.AircraftType)
		case *parser.StatusMessage:
			fmt.Printf("[STATUS] %-12s v%s cpu=%.1f ram=%.0f/%.0fMB\n",
				m.Callsign, m.Version, m.CPULoad, m.FreeRAM, m.TotalRAM)
		case *parser.ServerMessage:
			fmt.Printf("[SERVER] %s\n", m.Raw)
		}
	}, true)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
