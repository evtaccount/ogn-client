# ogn-client

Pure Go library for OGN (Open Glider Network) APRS protocol. Zero external dependencies. Inspired by python-ogn-client.

## Structure
```
client/
  client.go      — TCP APRS client with auto-reconnect & keep-alive
  filter.go      — APRS filter string builders
parser/
  parser.go      — Position/status/server/comment message parsing, CheapRuler, utilities
  parser_test.go — Tests
ddb/
  ddb.go         — OGN device database (HTTP JSON download, lookup by ID)
examples/
  main.go        — Working example
```

## Public API

### client
- `New(username, filter) *Client` — create client
- `Client.Run(callback, autoreconnect)` — blocking read loop, callback per APRS line
- `Client.RunWithTimedCallback(cb, timedCb, autoreconnect)` — + periodic callback
- `Client.Connect()` / `Client.Disconnect()`
- `Client.Filter` — settable APRS filter string
- Filter builders: `RangeFilter`, `AreaFilter`, `PrefixFilter`, `BudlistFilter`, `TypeFilter`, `CombineFilters`
- Auto port: 10152 (full feed) vs 14580 (filtered)
- Keep-alive: 240s default, concurrent goroutine
- Auto-reconnect: 5s backoff

### parser
- `Parse(raw) (interface{}, error)` → `*PositionMessage`, `*StatusMessage`, `*ServerMessage`, or `*CommentMessage`
- `ParsePosition(raw) (*PositionMessage, error)` — convenience wrapper
- **PositionMessage** (30+ fields): Callsign, Lat/Lon, Altitude(m), GroundSpeed(km/h), ClimbRate(m/s), TurnRate, SignalQuality, Address, AircraftType, BeaconType, FlightLevel, etc.
- **StatusMessage** (20+ fields): receiver diagnostics — CPU, RAM, temp, RF noise, senders
- **ServerMessage**: server info after login
- **CommentMessage**: plain comments
- **CheapRuler**: fast Distance/Bearing for <500km
- **NormalizedQuality**: signal normalization to 10km
- Unit constants: FeetToMeter, FpmToMs, KnotsToKph, HpmToDegs

### ddb
- `GetDevices() ([]Device, error)` — download from ddb.glidernet.org
- `LookupByID(devices) map[string]Device` — index by DeviceID
- **Device**: DeviceType, DeviceID, AircraftModel, Registration, CN, Tracked, Identified

## AircraftType values
7=Paraglider, 6=Hang glider, 1=Glider, 2=Tow plane, 4=Parachute, 8=Powered aircraft

## BeaconType detection
Maps APRS destination → type: OGFLR=Flarm, OGTRK=Tracker, OGNFNT=Fanet, OGNSDR=Receiver, etc.

## Testing
```
go test ./parser -v
```
