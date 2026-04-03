package ddb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const ddbURL = "http://ddb.glidernet.org/download/?j=1"

// Device represents a single entry in the OGN device database.
type Device struct {
	DeviceType    string `json:"device_type"`
	DeviceID      string `json:"device_id"`
	AircraftModel string `json:"aircraft_model"`
	Registration  string `json:"registration"`
	CN            string `json:"cn"`
	Tracked       bool   `json:"tracked"`
	Identified    bool   `json:"identified"`
}

type response struct {
	Devices []struct {
		DeviceType    string `json:"device_type"`
		DeviceID      string `json:"device_id"`
		AircraftModel string `json:"aircraft_model"`
		Registration  string `json:"registration"`
		CN            string `json:"cn"`
		Tracked       string `json:"tracked"`
		Identified    string `json:"identified"`
	} `json:"devices"`
}

// GetDevices downloads the device database from OGN.
// It uses a 30-second HTTP timeout to avoid hanging on slow networks.
func GetDevices() ([]Device, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(ddbURL)
	if err != nil {
		return nil, fmt.Errorf("fetch device database: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device database returned HTTP %d", resp.StatusCode)
	}

	var r response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("decode device database: %w", err)
	}

	devices := make([]Device, 0, len(r.Devices))
	for _, d := range r.Devices {
		devices = append(devices, Device{
			DeviceType:    d.DeviceType,
			DeviceID:      d.DeviceID,
			AircraftModel: d.AircraftModel,
			Registration:  d.Registration,
			CN:            d.CN,
			Tracked:       d.Tracked == "Y",
			Identified:    d.Identified == "Y",
		})
	}
	return devices, nil
}

// LookupByID builds a map from device ID to Device for fast lookups.
func LookupByID(devices []Device) map[string]Device {
	m := make(map[string]Device, len(devices))
	for _, d := range devices {
		m[d.DeviceID] = d
	}
	return m
}
