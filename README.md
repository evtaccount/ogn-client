# ogn-client

Go library for the [Open Glider Network](http://glidernet.org/) (OGN) — a global network for tracking gliders and light aircraft via APRS.

Inspired by [python-ogn-client](https://github.com/glidernet/python-ogn-client), rewritten in Go with zero external dependencies.

## Features

- **APRS client** — connect to OGN servers, read packets, auto-reconnect, keep-alive
- **Message parser** — decode position reports (30+ fields), receiver status, server comments
- **Device database** — download and query the OGN device registry
- **Geographic filtering** — helper functions for APRS server-side range/area filters
- **Utilities** — CheapRuler (fast distance/bearing), unit conversions, signal normalization

## Installation

```bash
go get ogn
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    "os"

    "ogn/client"
    "ogn/parser"
)

func main() {
    // Track aircraft within 100 km of Munich
    filter := client.RangeFilter(48.14, 11.58, 100)
    c := client.New("myapp", filter)
    c.Logger = log.New(os.Stderr, "[ogn] ", log.LstdFlags)

    c.Run(func(raw string) {
        msg, err := parser.Parse(raw)
        if err != nil {
            return
        }
        if pm, ok := msg.(*parser.PositionMessage); ok {
            fmt.Printf("%s  lat=%.4f lon=%.4f alt=%.0fm climb=%+.1fm/s\n",
                pm.Callsign, pm.Latitude, pm.Longitude, pm.Altitude, pm.ClimbRate)
        }
    }, true) // autoreconnect=true
}
```

## Packages

### `client` — APRS Connection

```go
// Create a client with geographic filter
c := client.New("mycallsign", client.RangeFilter(48.0, 11.0, 100))

// Or without filter (full feed — high traffic!)
c := client.New("mycallsign", "")

// Optional: set custom logger
c.Logger = log.New(os.Stderr, "[ogn] ", log.LstdFlags)

// Connect and run
c.Connect()
c.Run(callback, autoreconnect)
c.Disconnect()
```

**Constants:**
| Name | Value | Description |
|------|-------|-------------|
| `DefaultHost` | `aprs.glidernet.org` | OGN APRS server |
| `PortFullFeed` | `10152` | Unfiltered stream |
| `PortFiltered` | `14580` | Server-side filtered stream |
| `DefaultKeepAlive` | `240s` | Keep-alive interval |

The client automatically selects the correct port based on whether a filter is set.

### `client` — APRS Filters

APRS servers support server-side filtering to reduce traffic. Helper functions:

```go
// Circular area: 100 km radius around a point
client.RangeFilter(48.14, 11.58, 100)    // "r/48.140000/11.580000/100"

// Rectangular area: NW corner to SE corner
client.AreaFilter(50.0, 5.0, 45.0, 15.0) // "a/50.000000/5.000000/45.000000/15.000000"

// Callsign prefix filter
client.PrefixFilter("OGN", "FLR")        // "p/OGN/FLR"

// Specific callsigns
client.BudlistFilter("FLR123456")         // "b/FLR123456"

// Packet type filter (p=position, o=object, w=weather, etc.)
client.TypeFilter("po")                   // "t/po"

// Combine multiple filters (OR logic)
client.CombineFilters(
    client.RangeFilter(48.14, 11.58, 100),
    client.PrefixFilter("FLR"),
)
```

**Available filter types on APRS-IS:**

| Filter | Syntax | Description |
|--------|--------|-------------|
| Range | `r/lat/lon/dist` | Circular area (dist in km) |
| Area | `a/latN/lonW/latS/lonE` | Rectangular bounding box |
| Prefix | `p/aa/bb` | Callsign prefix match |
| Budlist | `b/call1/call2` | Specific callsigns |
| Type | `t/types` | Packet type filter |
| Symbol | `s/sym` | APRS symbol filter |
| Digipeater | `d/call` | Via digipeater |
| Entry station | `e/call/dist` | Via entry station |
| Friend range | `f/call/dist` | Range around a station |
| My range | `m/dist` | Range around your position |
| Object | `o/obj` | Object name filter |
| Negation | `-p/aa` | Exclude (prefix with `-`) |

Multiple filters are space-separated and use OR logic.

### `parser` — Message Parsing

```go
msg, err := parser.Parse(rawLine)
switch m := msg.(type) {
case *parser.PositionMessage:
    // Aircraft/glider position report
case *parser.StatusMessage:
    // Receiver diagnostic status
case *parser.ServerMessage:
    // Server login response
case *parser.CommentMessage:
    // Comment line (starts with #)
}
```

#### PositionMessage Fields

| Field | Type | Description |
|-------|------|-------------|
| `Callsign` | `string` | Sender callsign (e.g., `FLRDDE626`) |
| `DstCall` | `string` | Destination call — indicates beacon type |
| `BeaconType` | `BeaconType` | Detected beacon type (`flarm`, `tracker`, `receiver`, etc.) |
| `ReceiverName` | `string` | Receiving station name |
| `Relay` | `string` | Relay station (if relayed) |
| `Timestamp` | `time.Time` | Position time (UTC) |
| `Latitude` | `float64` | Decimal degrees |
| `Longitude` | `float64` | Decimal degrees |
| `Altitude` | `float64` | Meters (converted from feet) |
| `Course` | `int` | Track in degrees |
| `GroundSpeed` | `float64` | km/h (converted from knots) |
| `ClimbRate` | `float64` | m/s (converted from fpm) |
| `TurnRate` | `float64` | deg/s |
| `SignalQuality` | `float64` | Signal strength in dB |
| `ErrorCount` | `int` | Transmission error count |
| `FreqOffset` | `float64` | Frequency offset in kHz |
| `GPSQuality` | `string` | GPS fix info (e.g., `"3x5"`) |
| `FlightLevel` | `float64` | Pressure altitude in FL |
| `SignalPower` | `float64` | Signal power in dBm |
| `SoftwareVer` | `float64` | Firmware version |
| `HardwareVer` | `int` | Hardware version |
| `Address` | `string` | 6-hex device address |
| `AddressType` | `int` | 0=random, 1=ICAO, 2=FLARM, 3=OGN |
| `AircraftType` | `int` | See table below |
| `RealAddress` | `string` | Original address (if randomized) |
| `Stealth` | `bool` | Stealth mode flag |
| `NoTracking` | `bool` | No-tracking flag |

**Aircraft Types:**

| Code | Type |
|------|------|
| 0 | Unknown |
| 1 | Glider / Motor glider |
| 2 | Tow plane |
| 3 | Helicopter / Rotorcraft |
| 4 | Parachute / Skydiver |
| 5 | Drop plane |
| 6 | Hang glider |
| 7 | Paraglider |
| 8 | Powered aircraft |
| 9 | Jet aircraft |
| 10 | UFO |
| 11 | Balloon |
| 12 | Airship |
| 13 | UAV / Drone |
| 15 | Static object |

**Beacon Types** (detected from destination callsign):

`flarm`, `tracker`, `receiver`, `fanet`, `spider`, `pilot_aware`, `skylines`, `spot`, `inreach`, `capturs`, `flymaster`, `lt24`, `safesky`, `naviter`, `microtrak`

#### StatusMessage Fields

Receiver diagnostic data:

| Field | Type | Description |
|-------|------|-------------|
| `Version` | `string` | Receiver software version |
| `Platform` | `string` | Hardware platform (e.g., `ARM`, `RPI-GPU`) |
| `CPULoad` | `float64` | CPU load |
| `FreeRAM` | `float64` | Free RAM in MB |
| `TotalRAM` | `float64` | Total RAM in MB |
| `NTPError` | `float64` | NTP offset in ms |
| `RTCrystalCorrection` | `float64` | NTP crystal correction in ppm |
| `Voltage` | `float64` | Supply voltage in V |
| `Amperage` | `float64` | Current draw in A |
| `CPUTemp` | `float64` | CPU temperature in C |
| `SendersVisible` | `int` | Aircraft visible now |
| `SendersTotal` | `int` | Total aircraft count |
| `RecInputNoise` | `float64` | RF noise floor in dB |
| `SendersSignal` | `float64` | Average signal in dB |

### `parser` — Utilities

```go
// Fast distance/bearing calculator (accurate under 500 km)
ruler := parser.NewCheapRuler(48.0) // calibrate for latitude
dist := ruler.Distance(
    [2]float64{11.0, 48.0}, // lon, lat point A
    [2]float64{11.1, 48.1}, // lon, lat point B
) // returns meters

bearing := ruler.Bearing(pointA, pointB) // returns degrees

// Signal quality normalized to 10 km
nq := parser.NormalizedQuality(distanceMeters, signalDB)
```

**Unit conversion constants:**
| Constant | Value | Description |
|----------|-------|-------------|
| `FeetToMeter` | 0.3048 | Feet to meters |
| `FpmToMs` | 0.00508 | Feet/min to m/s |
| `KnotsToKph` | 1.852 | Knots to km/h |
| `HpmToDegs` | 3.0 | Half-turn/min to deg/s |

### `ddb` — Device Database

```go
// Download all devices
devices, err := ddb.GetDevices()

// Build a lookup map
lookup := ddb.LookupByID(devices)
if dev, ok := lookup["DDE626"]; ok {
    fmt.Printf("%s (%s)\n", dev.Registration, dev.AircraftModel)
}
```

**Device fields:** `DeviceType`, `DeviceID`, `AircraftModel`, `Registration`, `CN` (competition number), `Tracked`, `Identified`.

## Example

A complete working example is in [`examples/main.go`](examples/main.go) — connects to OGN with a geographic filter and prints parsed positions.

```bash
cd examples && go run main.go
```

## Architecture

```
ogn-client/
├── client/
│   ├── client.go      # APRS TCP client with auto-reconnect
│   └── filter.go      # APRS filter string builders
├── parser/
│   ├── parser.go      # Message parsing, types, utilities
│   └── parser_test.go # Tests
├── ddb/
│   └── ddb.go         # OGN device database client
├── examples/
│   └── main.go        # Usage example
├── go.mod
└── README.md
```

## Differences from python-ogn-client

| Feature | Python | Go |
|---------|--------|----|
| Parser engine | Rust (ogn-parser-rs) | Native regex |
| Weather messages | Full parsing | Parsed as position (raw in UserComment) |
| Telnet client | Yes | No |
| External deps | ogn-parser-rs | None |
| Position fields | 40+ | 30+ |
| Status fields | 25+ | 20+ |

## License

See [python-ogn-client](https://github.com/glidernet/python-ogn-client) for the original license.
