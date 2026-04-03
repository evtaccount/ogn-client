package parser

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Unit conversion constants (matching python-ogn-client).
const (
	FeetToMeter   = 0.3048
	FpmToMs       = FeetToMeter / 60 // feet per minute → m/s
	KnotsToKph    = 1.852
	HpmToDegs     = 180.0 / 60.0 // half-turn per minute → deg/s
	InchToMm      = 25.4
)

// MessageType identifies the kind of parsed APRS message.
type MessageType string

const (
	TypePosition        MessageType = "position"
	TypePositionWeather MessageType = "position_weather"
	TypeStatus          MessageType = "status"
	TypeServer          MessageType = "server"
	TypeComment         MessageType = "comment"
	TypeUnknown         MessageType = "unknown"
)

// BeaconType identifies the sender hardware/platform.
type BeaconType string

const (
	BeaconFlarm      BeaconType = "flarm"
	BeaconTracker    BeaconType = "tracker"
	BeaconReceiver   BeaconType = "receiver"
	BeaconFanet      BeaconType = "fanet"
	BeaconSpider     BeaconType = "spider"
	BeaconPilotAware BeaconType = "pilot_aware"
	BeaconSkylines   BeaconType = "skylines"
	BeaconSpot       BeaconType = "spot"
	BeaconInreach    BeaconType = "inreach"
	BeaconCapturs    BeaconType = "capturs"
	BeaconFlymaster  BeaconType = "flymaster"
	BeaconLT24       BeaconType = "lt24"
	BeaconSafesky    BeaconType = "safesky"
	BeaconNaviter    BeaconType = "naviter"
	BeaconMicrotrak  BeaconType = "microtrak"
	BeaconUnknown    BeaconType = "unknown"
)

// dstcallBeaconType maps the APRS destination callsign to a beacon type.
var dstcallBeaconType = map[string]BeaconType{
	"OGCAPT":  BeaconCapturs,
	"OGNFNT":  BeaconFanet,
	"OGFLR":   BeaconFlarm,
	"OGFLR6":  BeaconFlarm,
	"OGFLR7":  BeaconFlarm,
	"OGFLYM":  BeaconFlymaster,
	"OGNINRE": BeaconInreach,
	"OGLT24":  BeaconLT24,
	"OGNMTK": BeaconMicrotrak,
	"OGNAVI":  BeaconNaviter,
	"OGNSDR":  BeaconReceiver,
	"OGNSKY":  BeaconSafesky,
	"OGPAW":   BeaconPilotAware,
	"OGSKYL":  BeaconSkylines,
	"OGSPID":  BeaconSpider,
	"OGSPOT":  BeaconSpot,
	"OGNTRK":  BeaconTracker,
}

// PositionMessage represents a decoded position report.
type PositionMessage struct {
	Raw            string
	Type           MessageType
	BeaconType     BeaconType
	Callsign       string
	DstCall        string
	ReceiverName   string
	Relay          string
	Timestamp      time.Time
	Latitude       float64
	Longitude      float64
	Altitude       float64 // meters
	Course         int     // degrees
	GroundSpeed    float64 // km/h
	ClimbRate      float64 // m/s
	TurnRate       float64 // deg/s
	SignalQuality  float64 // dB
	ErrorCount     int
	FreqOffset     float64 // kHz
	GPSQuality     string  // e.g. "3x5"
	FlightLevel    float64
	SignalPower    float64 // dBm
	SoftwareVer    float64
	HardwareVer    int
	Address        string // 6-hex ICAO-like address
	AddressType    int
	AircraftType   int
	RealAddress    string
	Stealth        bool
	NoTracking     bool
	UserComment    string
}

// StatusMessage represents a decoded receiver status report.
type StatusMessage struct {
	Raw                    string
	Type                   MessageType
	Callsign               string
	Timestamp              time.Time
	Version                string
	Platform               string
	CPULoad                float64
	FreeRAM                float64 // MB
	TotalRAM               float64 // MB
	NTPError               float64 // ms
	RTCrystalCorrection    float64 // ppm
	Voltage                float64 // V
	Amperage               float64 // A
	CPUTemp                float64 // °C
	SendersVisible         int
	Latency                float64 // s
	SendersTotal           int
	RecCrystalCorrection   int     // manual, ppm
	RecCrystalCorrFine     float64 // auto, ppm
	RecInputNoise          float64 // dB
	SendersSignal          float64 // dB
	SendersMessages        int
	GoodSendersSignal      float64
	GoodSenders            int
	GoodAndBadSenders      int
	UserComment            string
}

// ServerMessage represents a server comment/timestamp line.
type ServerMessage struct {
	Raw       string
	Type      MessageType
	Server    string
	IPAddress string
	Port      int
	Version   string
	Timestamp time.Time
}

// CommentMessage represents a plain comment line (starts with #).
type CommentMessage struct {
	Raw     string
	Type    MessageType
	Comment string
}

// Regex patterns.
var (
	// Position pattern: captures header + coordinates + course/speed/alt + OGN extension.
	rePosition = regexp.MustCompile(
		`^(?P<callsign>[\w-]+)>(?P<dstcall>[\w-]+)` +
			`(?:,(?P<path>[^:]+))?` +
			`:/?(?P<ts>\d{6})h` +
			`(?P<latdeg>\d{2})(?P<latmin>\d{2}\.\d{2})(?P<lathem>[NS])` +
			`(?P<sym1>.)` +
			`(?P<londeg>\d{3})(?P<lonmin>\d{2}\.\d{2})(?P<lonhem>[EW])` +
			`(?P<sym2>.)` +
			`(?:(?P<course>\d{3})/(?P<speed>\d{3}))?` +
			`(?:/A=(?P<alt>-?\d+))?` +
			`(?:\s+(?P<comment>.*))?$`)

	// Status pattern: callsign>dstcall,path:>HHMMSSh status_text
	reStatus = regexp.MustCompile(
		`^(?P<callsign>[\w-]+)>(?P<dstcall>[\w*,-]+):>` +
			`(?P<ts>\d{6})h\s*(?P<status>.*)$`)

	// Server comment: # logresp ... server ...
	reServer = regexp.MustCompile(
		`^# logresp\s+`)

	// OGN extension after the position report.
	reOGN = regexp.MustCompile(
		`!W(.)(.)!` +
			`\s+id(?P<idflags>[0-9A-Fa-f]{2})(?P<id>[0-9A-Fa-f]{6})` +
			`(?:\s+(?P<climbrate>[+-]?\d+)fpm)?` +
			`(?:\s+(?P<turnrate>[+-]?\d+(?:\.\d+)?)rot)?` +
			`(?:\s+FL(?P<fl>\d+(?:\.\d+)?))?` +
			`(?:\s+(?P<signal>[+-]?\d+(?:\.\d+)?)dB)?` +
			`(?:\s+(?P<errors>\d+)e)?` +
			`(?:\s+(?P<freqofs>[+-]?\d+(?:\.\d+)?)kHz)?` +
			`(?:\s+gps(?P<gps>\d+x\d+))?` +
			`(?:\s+s(?P<sw>\d+\.\d+))?` +
			`(?:\s+h(?P<hw>[0-9A-Fa-f]+))?` +
			`(?:\s+r(?P<addr>[0-9A-Fa-f]{6}))?` +
			`(?:\s+(?P<power>[+-]?\d+(?:\.\d+)?)dBm)?`)

	// Receiver status fields.
	reReceiverStatus = regexp.MustCompile(
		`v(?P<version>\d+\.\d+\.\d+)\.(?P<platform>[\w-]+)` +
			`(?:\s+CPU:(?P<cpu>[\d.]+))?` +
			`(?:\s+RAM:(?P<ram_free>[\d.]+)/(?P<ram_total>[\d.]+)MB)?` +
			`(?:\s+NTP:(?P<ntp_offset>[\d.]+)ms/(?P<ntp_correction>[+-]?[\d.]+)ppm)?` +
			`(?:\s+(?P<voltage>[\d.]+)V)?` +
			`(?:\s+(?P<amperage>[\d.]+)A)?` +
			`(?:\s+(?P<cpu_temp>[+-]?[\d.]+)C)?` +
			`(?:\s+(?P<visible_senders>\d+)/(?P<senders>\d+)Acfts)?` +
			`(?:\s+Lat:(?P<latency>[\d.]+)s)?` +
			`(?:\s+RF:(?P<rf_correction_manual>[+-]?\d+)(?P<rf_correction_automatic>[+-][\d.]+)ppm/(?P<noise>[+-]?[\d.]+)dB)?` +
			`(?:\s+Senders:(?P<senders_signal>[+-]?[\d.]+)dB/(?P<senders_messages>\d+))?` +
			`(?:\s+Good:(?P<good_senders_signal>[+-]?[\d.]+)dB/(?P<good_senders>\d+)/(?P<good_and_bad_senders>\d+))?`)
)

// Parse attempts to parse a raw APRS/OGN message line.
// It returns one of: *PositionMessage, *StatusMessage, *ServerMessage, *CommentMessage.
func Parse(raw string) (interface{}, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("empty message")
	}

	// Server comments.
	if strings.HasPrefix(raw, "# logresp") {
		return &ServerMessage{Raw: raw, Type: TypeServer}, nil
	}
	// Generic comments.
	if strings.HasPrefix(raw, "#") {
		return &CommentMessage{Raw: raw, Type: TypeComment, Comment: strings.TrimSpace(raw[1:])}, nil
	}

	// Try status message.
	if sm := parseStatus(raw); sm != nil {
		return sm, nil
	}

	// Try position message.
	if pm := parsePosition(raw); pm != nil {
		return pm, nil
	}

	return nil, fmt.Errorf("unrecognized message format: %.80s", raw)
}

// ParsePosition parses only position messages, returning an error for other types.
// This is a convenience wrapper for callers that only care about position data.
func ParsePosition(raw string) (*PositionMessage, error) {
	msg := parsePosition(strings.TrimSpace(raw))
	if msg == nil {
		return nil, errors.New("not a position message")
	}
	return msg, nil
}

func parsePosition(raw string) *PositionMessage {
	m := rePosition.FindStringSubmatch(raw)
	if m == nil {
		return nil
	}
	get := submatchGetter(rePosition, m)

	dstcall := get("dstcall")
	bt, ok := dstcallBeaconType[dstcall]
	if !ok {
		bt = BeaconUnknown
	}

	pm := &PositionMessage{
		Raw:        raw,
		Type:       TypePosition,
		BeaconType: bt,
		Callsign:   get("callsign"),
		DstCall:    dstcall,
	}

	// Receiver name and relay from path.
	if path := get("path"); path != "" {
		parts := strings.Split(path, ",")
		if len(parts) > 0 {
			last := parts[len(parts)-1]
			pm.ReceiverName = strings.TrimSuffix(last, "*")
		}
		if len(parts) > 0 && parts[0] != "TCPIP*" && strings.HasSuffix(parts[0], "*") {
			pm.Relay = strings.TrimSuffix(parts[0], "*")
		}
	}

	// Coordinates.
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
	pm.Latitude = lat
	pm.Longitude = lon

	// Course and speed.
	if cs := get("course"); cs != "" {
		pm.Course, _ = strconv.Atoi(cs)
	}
	if sp := get("speed"); sp != "" {
		knots, _ := strconv.ParseFloat(sp, 64)
		pm.GroundSpeed = knots * KnotsToKph
	}

	// Altitude.
	if alt := get("alt"); alt != "" {
		altFeet, _ := strconv.Atoi(alt)
		pm.Altitude = float64(altFeet) * FeetToMeter
	}

	// Timestamp with day-boundary correction.
	if ts := get("ts"); ts != "" {
		pm.Timestamp = createTimestamp(ts, time.Now().UTC())
	}

	// OGN-specific extension in the comment field.
	comment := get("comment")
	if comment != "" {
		parseOGNExtension(pm, comment)
	}

	return pm
}

func parseOGNExtension(pm *PositionMessage, comment string) {
	// Enhanced position precision: !Wab! adds to lat/lon minutes.
	if i := strings.Index(comment, "!W"); i >= 0 && i+4 <= len(comment) && comment[i+3] == '!' {
		latEnhance := float64(comment[i+2]-'0') / 1000.0 / 60.0
		lonEnhance := float64(comment[i+3]-'0') / 1000.0 / 60.0
		_ = latEnhance
		_ = lonEnhance
		// The precision digits refine the last decimal; for simplicity we note them
		// but the regex already captures 2 decimal places of minutes.
	}

	m := reOGN.FindStringSubmatch(comment)
	if m == nil {
		pm.UserComment = comment
		return
	}
	get := submatchGetter(reOGN, m)

	// ID flags byte: bits 7-6 = address type, bits 5-2 = aircraft type, bit 1 = notrack, bit 0 = stealth.
	if flags := get("idflags"); flags != "" {
		if v, err := strconv.ParseUint(flags, 16, 8); err == nil {
			pm.AddressType = int((v >> 6) & 0x03)
			pm.AircraftType = int((v >> 2) & 0x0F)
			pm.NoTracking = (v & 0x02) != 0
			pm.Stealth = (v & 0x01) != 0
		}
	}
	pm.Address = strings.ToUpper(get("id"))

	if v := get("climbrate"); v != "" {
		fpm, _ := strconv.ParseFloat(v, 64)
		pm.ClimbRate = fpm * FpmToMs
	}
	if v := get("turnrate"); v != "" {
		rot, _ := strconv.ParseFloat(v, 64)
		pm.TurnRate = rot * HpmToDegs
	}
	if v := get("fl"); v != "" {
		pm.FlightLevel, _ = strconv.ParseFloat(v, 64)
	}
	if v := get("signal"); v != "" {
		pm.SignalQuality, _ = strconv.ParseFloat(v, 64)
	}
	if v := get("errors"); v != "" {
		pm.ErrorCount, _ = strconv.Atoi(v)
	}
	if v := get("freqofs"); v != "" {
		pm.FreqOffset, _ = strconv.ParseFloat(v, 64)
	}
	if v := get("gps"); v != "" {
		pm.GPSQuality = v
	}
	if v := get("sw"); v != "" {
		pm.SoftwareVer, _ = strconv.ParseFloat(v, 64)
	}
	if v := get("hw"); v != "" {
		hw, _ := strconv.ParseUint(v, 16, 32)
		pm.HardwareVer = int(hw)
	}
	if v := get("addr"); v != "" {
		pm.RealAddress = strings.ToUpper(v)
	}
	if v := get("power"); v != "" {
		pm.SignalPower, _ = strconv.ParseFloat(v, 64)
	}
}

func parseStatus(raw string) *StatusMessage {
	m := reStatus.FindStringSubmatch(raw)
	if m == nil {
		return nil
	}
	get := submatchGetter(reStatus, m)

	sm := &StatusMessage{
		Raw:      raw,
		Type:     TypeStatus,
		Callsign: get("callsign"),
	}

	if ts := get("ts"); ts != "" {
		sm.Timestamp = createTimestamp(ts, time.Now().UTC())
	}

	status := get("status")
	rm := reReceiverStatus.FindStringSubmatch(status)
	if rm == nil {
		sm.UserComment = status
		return sm
	}
	rget := submatchGetter(reReceiverStatus, rm)

	sm.Version = rget("version")
	sm.Platform = rget("platform")
	sm.CPULoad = parseFloat(rget("cpu"))
	sm.FreeRAM = parseFloat(rget("ram_free"))
	sm.TotalRAM = parseFloat(rget("ram_total"))
	sm.NTPError = parseFloat(rget("ntp_offset"))
	sm.RTCrystalCorrection = parseFloat(rget("ntp_correction"))
	sm.Voltage = parseFloat(rget("voltage"))
	sm.Amperage = parseFloat(rget("amperage"))
	sm.CPUTemp = parseFloat(rget("cpu_temp"))
	sm.SendersVisible = parseInt(rget("visible_senders"))
	sm.SendersTotal = parseInt(rget("senders"))
	sm.Latency = parseFloat(rget("latency"))
	sm.RecCrystalCorrection = parseInt(rget("rf_correction_manual"))
	sm.RecCrystalCorrFine = parseFloat(rget("rf_correction_automatic"))
	sm.RecInputNoise = parseFloat(rget("noise"))
	sm.SendersSignal = parseFloat(rget("senders_signal"))
	sm.SendersMessages = parseInt(rget("senders_messages"))
	sm.GoodSendersSignal = parseFloat(rget("good_senders_signal"))
	sm.GoodSenders = parseInt(rget("good_senders"))
	sm.GoodAndBadSenders = parseInt(rget("good_and_bad_senders"))

	return sm
}

// createTimestamp reconstructs a full UTC time from an HHMMSS string and a
// reference timestamp, correcting for day boundaries (±12 h tolerance).
func createTimestamp(ts string, ref time.Time) time.Time {
	if len(ts) < 6 {
		return ref
	}
	hh, _ := strconv.Atoi(ts[0:2])
	mm, _ := strconv.Atoi(ts[2:4])
	ss, _ := strconv.Atoi(ts[4:6])

	result := time.Date(ref.Year(), ref.Month(), ref.Day(), hh, mm, ss, 0, time.UTC)

	// Correct for day boundary.
	if result.Sub(ref) > 12*time.Hour {
		result = result.AddDate(0, 0, -1)
	} else if ref.Sub(result) > 12*time.Hour {
		result = result.AddDate(0, 0, 1)
	}
	return result
}

// String formats a position message for display.
func (p *PositionMessage) String() string {
	return fmt.Sprintf("%s %.5f %.5f %.0fm %.0fkm/h %+.1fm/s",
		p.Callsign, p.Latitude, p.Longitude, p.Altitude, p.GroundSpeed, p.ClimbRate)
}

// --- Utilities ---

// CheapRuler performs fast distance/bearing calculations for distances under 500 km.
// Ported from python-ogn-client / cheap-ruler.
type CheapRuler struct {
	kx, ky float64
}

// NewCheapRuler creates a ruler calibrated for the given latitude (decimal degrees).
func NewCheapRuler(lat float64) *CheapRuler {
	c := math.Cos(lat * math.Pi / 180)
	c2 := 2*c*c - 1
	c3 := 2*c*c2 - c
	c4 := 2*c*c3 - c2
	c5 := 2*c*c4 - c3
	return &CheapRuler{
		kx: 1000 * (111.41513*c - 0.09455*c3 + 0.00012*c5),
		ky: 1000 * (111.13209 - 0.56605*c2 + 0.0012*c4),
	}
}

// Distance returns the distance in meters between two points (lon, lat).
func (r *CheapRuler) Distance(lonLatA, lonLatB [2]float64) float64 {
	dx := (lonLatA[0] - lonLatB[0]) * r.kx
	dy := (lonLatA[1] - lonLatB[1]) * r.ky
	return math.Sqrt(dx*dx + dy*dy)
}

// Bearing returns the bearing in degrees from point A to point B.
func (r *CheapRuler) Bearing(lonLatA, lonLatB [2]float64) float64 {
	dx := (lonLatB[0] - lonLatA[0]) * r.kx
	dy := (lonLatB[1] - lonLatA[1]) * r.ky
	if dx == 0 && dy == 0 {
		return 0
	}
	result := math.Atan2(-dy, dx)*180/math.Pi + 90
	if result < 0 {
		result += 360
	}
	return result
}

// NormalizedQuality returns signal quality normalized to 10 km distance.
func NormalizedQuality(distanceM, signalQualityDB float64) float64 {
	if distanceM <= 0 {
		return 0
	}
	return signalQualityDB + 20.0*math.Log10(distanceM/10000.0)
}

// --- helpers ---

func submatchGetter(re *regexp.Regexp, m []string) func(string) string {
	names := re.SubexpNames()
	return func(name string) string {
		for i, n := range names {
			if n == name && i < len(m) {
				return m[i]
			}
		}
		return ""
	}
}

func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func parseInt(s string) int {
	if s == "" {
		return 0
	}
	v, _ := strconv.Atoi(s)
	return v
}
