package client

import "fmt"

// RangeFilter returns an APRS filter string for a circular area.
// lat/lon are the center coordinates in decimal degrees; dist is the radius in km.
// Example: RangeFilter(48.0, 11.0, 100) => "r/48.000000/11.000000/100"
func RangeFilter(lat, lon float64, distKm int) string {
	return fmt.Sprintf("r/%f/%f/%d", lat, lon, distKm)
}

// AreaFilter returns an APRS filter string for a rectangular bounding box.
// latN/lonW define the north-west corner; latS/lonE define the south-east corner.
// Example: AreaFilter(50.0, 5.0, 45.0, 15.0) => "a/50.000000/5.000000/45.000000/15.000000"
func AreaFilter(latN, lonW, latS, lonE float64) string {
	return fmt.Sprintf("a/%f/%f/%f/%f", latN, lonW, latS, lonE)
}

// PrefixFilter returns an APRS filter for callsign prefixes.
// Example: PrefixFilter("OGN", "FLR") => "p/OGN/FLR"
func PrefixFilter(prefixes ...string) string {
	s := "p"
	for _, p := range prefixes {
		s += "/" + p
	}
	return s
}

// BudlistFilter returns a filter that passes traffic from/to specific callsigns.
// Example: BudlistFilter("FLR123456", "OGN123456") => "b/FLR123456/OGN123456"
func BudlistFilter(callsigns ...string) string {
	s := "b"
	for _, c := range callsigns {
		s += "/" + c
	}
	return s
}

// TypeFilter returns an APRS type filter.
// Characters: p=position, o=object, i=item, m=message, w=weather, n=NWS, t=telemetry.
// Example: TypeFilter("po") => "t/po"
func TypeFilter(types string) string {
	return "t/" + types
}

// CombineFilters joins multiple filter expressions with spaces (OR logic on the server).
// Example: CombineFilters(RangeFilter(48, 11, 50), PrefixFilter("FLR")) => "r/48.../50 p/FLR"
func CombineFilters(filters ...string) string {
	result := ""
	for i, f := range filters {
		if i > 0 {
			result += " "
		}
		result += f
	}
	return result
}
