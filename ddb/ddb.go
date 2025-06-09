package ddb

import (
	"encoding/json"
	"net/http"
)

const ddbURL = "http://ddb.glidernet.org/download/?j=1"

// Device represents a single entry in the device database.
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
func GetDevices() ([]Device, error) {
	resp, err := http.Get(ddbURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var r response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
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
