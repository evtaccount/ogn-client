package parser

import (
	"math"
	"testing"
	"time"
)

func TestParsePosition(t *testing.T) {
	raw := `FLRDDE626>OGFLR,qAS,LFLE:/092045h4533.70N/00558.52E'246/030/A=002854 !W79! id06DDE626 -019fpm +0.0rot FL028.52 17.8dB 0e -5.0kHz gps3x5 s6.01 h44 rDD4711`

	msg, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	pm, ok := msg.(*PositionMessage)
	if !ok {
		t.Fatalf("expected *PositionMessage, got %T", msg)
	}

	if pm.Callsign != "FLRDDE626" {
		t.Errorf("callsign = %q, want FLRDDE626", pm.Callsign)
	}
	if pm.DstCall != "OGFLR" {
		t.Errorf("dstcall = %q, want OGFLR", pm.DstCall)
	}
	if pm.BeaconType != BeaconFlarm {
		t.Errorf("beacon type = %q, want flarm", pm.BeaconType)
	}
	if pm.ReceiverName != "LFLE" {
		t.Errorf("receiver = %q, want LFLE", pm.ReceiverName)
	}

	// Lat: 45 + 33.70/60 = 45.56167
	if math.Abs(pm.Latitude-45.56167) > 0.001 {
		t.Errorf("latitude = %f, want ~45.5617", pm.Latitude)
	}
	// Lon: 5 + 58.52/60 = 5.97533
	if math.Abs(pm.Longitude-5.97533) > 0.001 {
		t.Errorf("longitude = %f, want ~5.9753", pm.Longitude)
	}
	// Alt: 2854 ft * 0.3048 = 869.87 m
	if math.Abs(pm.Altitude-869.87) > 1.0 {
		t.Errorf("altitude = %f, want ~869.87", pm.Altitude)
	}
	if pm.Course != 246 {
		t.Errorf("course = %d, want 246", pm.Course)
	}
	// Speed: 30 knots * 1.852 = 55.56 km/h
	if math.Abs(pm.GroundSpeed-55.56) > 0.1 {
		t.Errorf("ground speed = %f, want ~55.56", pm.GroundSpeed)
	}
	if pm.Address != "DDE626" {
		t.Errorf("address = %q, want DDE626", pm.Address)
	}
	if pm.AircraftType != 1 {
		// idflags 06 = 0b00000110 → address_type=0, aircraft_type=1, notrack=1, stealth=0
		t.Errorf("aircraft type = %d, want 1", pm.AircraftType)
	}
	if math.Abs(pm.SignalQuality-17.8) > 0.01 {
		t.Errorf("signal quality = %f, want 17.8", pm.SignalQuality)
	}
	if pm.ErrorCount != 0 {
		t.Errorf("error count = %d, want 0", pm.ErrorCount)
	}
	if pm.GPSQuality != "3x5" {
		t.Errorf("gps quality = %q, want 3x5", pm.GPSQuality)
	}
	if math.Abs(pm.SoftwareVer-6.01) > 0.001 {
		t.Errorf("software version = %f, want 6.01", pm.SoftwareVer)
	}
	if pm.HardwareVer != 0x44 {
		t.Errorf("hardware version = %d, want 68 (0x44)", pm.HardwareVer)
	}
	if pm.RealAddress != "DD4711" {
		t.Errorf("real address = %q, want DD4711", pm.RealAddress)
	}
}

func TestParseStatus(t *testing.T) {
	raw := `EPZR>APRS,TCPIP*,qAC,GLIDERN1:>093456h v0.2.7.RPI-GPU CPU:0.7 RAM:770.2/968.2MB NTP:1.8ms/-3.3ppm +55.7C 7/8Acfts[1h] RF:+54-1.1ppm/-0.16dB`

	msg, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	sm, ok := msg.(*StatusMessage)
	if !ok {
		t.Fatalf("expected *StatusMessage, got %T", msg)
	}

	if sm.Callsign != "EPZR" {
		t.Errorf("callsign = %q, want EPZR", sm.Callsign)
	}
	if sm.Version != "0.2.7" {
		t.Errorf("version = %q, want 0.2.7", sm.Version)
	}
	if sm.Platform != "RPI-GPU" {
		t.Errorf("platform = %q, want RPI-GPU", sm.Platform)
	}
	if math.Abs(sm.CPULoad-0.7) > 0.01 {
		t.Errorf("cpu load = %f, want 0.7", sm.CPULoad)
	}
	if math.Abs(sm.FreeRAM-770.2) > 0.1 {
		t.Errorf("free ram = %f, want 770.2", sm.FreeRAM)
	}
	if sm.SendersVisible != 7 {
		t.Errorf("visible senders = %d, want 7", sm.SendersVisible)
	}
}

func TestParseComment(t *testing.T) {
	raw := `# this is a comment`
	msg, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	cm, ok := msg.(*CommentMessage)
	if !ok {
		t.Fatalf("expected *CommentMessage, got %T", msg)
	}
	if cm.Comment != "this is a comment" {
		t.Errorf("comment = %q", cm.Comment)
	}
}

func TestParseServerComment(t *testing.T) {
	raw := `# logresp user unverified, server GLIDERN1`
	msg, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	sm, ok := msg.(*ServerMessage)
	if !ok {
		t.Fatalf("expected *ServerMessage, got %T", msg)
	}
	if sm.Type != TypeServer {
		t.Errorf("type = %q, want server", sm.Type)
	}
}

func TestCheapRulerDistance(t *testing.T) {
	ruler := NewCheapRuler(48.0)
	a := [2]float64{11.0, 48.0}
	b := [2]float64{11.1, 48.1}
	d := ruler.Distance(a, b)
	// Roughly 13 km
	if d < 12000 || d > 14000 {
		t.Errorf("distance = %f, want ~13000m", d)
	}
}

func TestCreateTimestamp(t *testing.T) {
	// 23:55:00 with ref at 00:05:00 next day → should roll back to previous day
	ref := mustTime("2024-06-15T00:05:00Z")
	ts := createTimestamp("235500", ref)
	if ts.Day() != 14 {
		t.Errorf("expected day 14, got %d (timestamp: %v)", ts.Day(), ts)
	}

	// 00:05:00 with ref at 23:55:00 → should roll forward to next day
	ref2 := mustTime("2024-06-15T23:55:00Z")
	ts2 := createTimestamp("000500", ref2)
	if ts2.Day() != 16 {
		t.Errorf("expected day 16, got %d (timestamp: %v)", ts2.Day(), ts2)
	}
}

func mustTime(s string) time.Time {
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return parsed
}
