package calendar

import (
	"fmt"
	"net/url"
)

func GoogleRouteURL(startLat, startLng, endLat, endLng *float64) string {
	if startLat == nil || startLng == nil || endLat == nil || endLng == nil {
		return ""
	}
	return fmt.Sprintf("https://www.google.com/maps/dir/?api=1&origin=%.6f,%.6f&destination=%.6f,%.6f", *startLat, *startLng, *endLat, *endLng)
}

func GooglePointURL(lat, lng *float64) string {
	if lat == nil || lng == nil {
		return ""
	}
	return fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%.6f,%.6f", *lat, *lng)
}

func GoogleSearchURL(location string) string {
	if location == "" {
		return ""
	}
	return "https://www.google.com/maps/search/?api=1&query=" + url.QueryEscape(location)
}
