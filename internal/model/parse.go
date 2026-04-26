package model

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func ParseCars(raw []map[string]any) []Car {
	out := make([]Car, 0, len(raw))
	for _, item := range raw {
		id, _ := asInt64(pick(item, "id", "ID", "car_id", "CarID"))
		name, _ := asString(pick(item, "name", "Name", "display_name", "displayName"))
		out = append(out, Car{ID: id, Name: name, Raw: item})
	}
	return out
}

func ParseDrives(raw []map[string]any) []Drive {
	out := make([]Drive, 0, len(raw))
	for _, item := range raw {
		distance := firstFloat(
			pick(item, "distance", "Distance"),
			pickNested(item, "odometer_details", "odometer_distance"),
		)
		durationSec := firstFloat(
			pick(item, "duration_sec", "durationSec"),
			floatFromMinutes(pick(item, "duration_min", "durationMin")),
		)
		d := Drive{
			ID:              pick(item, "id", "ID", "drive_id", "DriveID"),
			StartDate:       asTimePtr(pick(item, "start_date", "startDate", "StartDate")),
			EndDate:         asTimePtr(pick(item, "end_date", "endDate", "EndDate")),
			Distance:        distance,
			DurationSeconds: durationSec,
			StartBatteryLevel: firstFloat(
				pick(item, "start_battery_level", "startBatteryLevel", "StartBatteryLevel"),
				pickNested(item, "battery_details", "start_battery_level"),
			),
			EndBatteryLevel: firstFloat(
				pick(item, "end_battery_level", "endBatteryLevel", "EndBatteryLevel"),
				pickNested(item, "battery_details", "end_battery_level"),
			),
			Consumption: firstFloat(
				pick(item, "consumption", "Consumption", "consumption_net"),
				pickNested(item, "energy_details", "consumption"),
			),
			AvgSpeed: firstFloat(
				pick(item, "average_speed", "avg_speed", "avgSpeed", "AvgSpeed", "speed_avg"),
				pickNested(item, "speed", "avg"),
			),
			MaxSpeed: firstFloat(
				pick(item, "max_speed", "maxSpeed", "MaxSpeed", "speed_max"),
				pickNested(item, "speed", "max"),
			),
			AvgPower: firstFloat(
				pick(item, "average_power", "avg_power", "avgPower", "AvgPower"),
				pickNested(item, "power", "avg"),
			),
			MaxPower: firstFloat(
				pick(item, "max_power", "maxPower", "MaxPower", "power_max"),
				pickNested(item, "power", "max"),
			),
			StartAddress:  asStringDefault(pick(item, "start_address", "startAddress", "StartAddress")),
			EndAddress:    asStringDefault(pick(item, "end_address", "endAddress", "EndAddress")),
			StartGeofence: asStringDefault(pick(item, "start_geofence", "startGeofence", "StartGeofence")),
			EndGeofence:   asStringDefault(pick(item, "end_geofence", "endGeofence", "EndGeofence")),
			StartLat:      asFloatPtr(pick(item, "start_lat", "startLat", "StartLat", "start_latitude", "startLatitude", "latitude_start")),
			StartLng:      asFloatPtr(pick(item, "start_lng", "startLng", "StartLng", "start_lon", "startLon", "start_longitude", "startLongitude", "longitude_start")),
			EndLat:        asFloatPtr(pick(item, "end_lat", "endLat", "EndLat", "end_latitude", "endLatitude", "latitude_end")),
			EndLng:        asFloatPtr(pick(item, "end_lng", "endLng", "EndLng", "end_lon", "endLon", "end_longitude", "endLongitude", "longitude_end")),
			Raw:           item,
		}
		out = append(out, d)
	}
	return out
}

func ParseCharges(raw []map[string]any) []Charge {
	out := make([]Charge, 0, len(raw))
	for _, item := range raw {
		c := Charge{
			ID:              pick(item, "id", "ID", "charge_id", "ChargeID"),
			StartDate:       asTimePtr(pick(item, "start_date", "startDate", "StartDate")),
			EndDate:         asTimePtr(pick(item, "end_date", "endDate", "EndDate")),
			DurationMinutes: firstFloat(pick(item, "duration_min", "durationMin")),
			StartBatteryLevel: firstFloat(
				pick(item, "start_battery_level", "startBatteryLevel", "StartBatteryLevel"),
				pickNested(item, "battery_details", "start_battery_level"),
			),
			EndBatteryLevel: firstFloat(
				pick(item, "end_battery_level", "endBatteryLevel", "EndBatteryLevel"),
				pickNested(item, "battery_details", "end_battery_level"),
			),
			KwhAdded: asFloatPtr(pick(item, "charge_energy_added", "kwh_added", "kwhAdded", "KwhAdded")),
			Cost:     asFloatPtr(pick(item, "cost", "Cost")),
			MaxPower: firstFloat(pick(item, "max_power", "maxPower", "MaxPower", "charger_power_max")),
			AvgPower: firstFloat(pick(item, "average_power", "avg_power", "avgPower", "AvgPower", "charger_power_avg")),
			Address:  asStringDefault(pick(item, "address", "Address", "location")),
			Geofence: asStringDefault(pick(item, "geofence", "Geofence")),
			Lat:      asFloatPtr(pick(item, "latitude", "lat", "Lat", "charge_lat", "chargeLat")),
			Lng:      asFloatPtr(pick(item, "longitude", "lng", "Lng", "lon", "charge_lng", "chargeLng", "charge_lon", "chargeLon")),
			Raw:      item,
		}
		out = append(out, c)
	}
	return out
}

func ParseUpdates(raw []map[string]any) []Update {
	out := make([]Update, 0, len(raw))
	for _, item := range raw {
		u := Update{
			ID:           pick(item, "id", "ID", "update_id", "UpdateID"),
			Version:      asStringDefault(pick(item, "version", "Version", "car_version", "carVersion")),
			StartDate:    asTimePtr(pick(item, "start_date", "startDate", "StartDate", "installed_at", "installedAt", "created_at", "createdAt")),
			EndDate:      asTimePtr(pick(item, "end_date", "endDate", "EndDate", "finished_at", "finishedAt")),
			Status:       asStringDefault(pick(item, "status", "Status", "state")),
			ReleaseNotes: asStringDefault(pick(item, "release_notes", "releaseNotes", "ReleaseNotes", "notes")),
			Raw:          item,
		}
		out = append(out, u)
	}
	return out
}

func ToMapSlice(v any) ([]map[string]any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var out []map[string]any
	if err := json.Unmarshal(b, &out); err == nil {
		return out, nil
	}
	var wrap struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(b, &wrap); err == nil && len(wrap.Data) >= 0 {
		return wrap.Data, nil
	}
	return nil, fmt.Errorf("unsupported array payload")
}

func pick(m map[string]any, keys ...string) any {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			return v
		}
		lk := strings.ToLower(k)
		for mk, mv := range m {
			if strings.ToLower(mk) == lk {
				return mv
			}
		}
	}
	return nil
}

func asTimePtr(v any) *time.Time {
	if v == nil {
		return nil
	}
	switch x := v.(type) {
	case string:
		if t, err := time.Parse(time.RFC3339, x); err == nil {
			return &t
		}
		layouts := []string{"2006-01-02 15:04:05", "2006-01-02 15:04", "2006-01-02"}
		for _, layout := range layouts {
			if t, err := time.ParseInLocation(layout, x, time.UTC); err == nil {
				ut := t.UTC()
				return &ut
			}
		}
	case float64:
		t := time.Unix(int64(x), 0).UTC()
		return &t
	case int64:
		t := time.Unix(x, 0).UTC()
		return &t
	}
	return nil
}

func asFloatPtr(v any) *float64 {
	if v == nil {
		return nil
	}
	switch x := v.(type) {
	case float64:
		return &x
	case float32:
		y := float64(x)
		return &y
	case int:
		y := float64(x)
		return &y
	case int64:
		y := float64(x)
		return &y
	case json.Number:
		if y, err := x.Float64(); err == nil {
			return &y
		}
	case string:
		if y, err := strconv.ParseFloat(x, 64); err == nil {
			return &y
		}
	}
	return nil
}

func asInt64(v any) (int64, bool) {
	switch x := v.(type) {
	case int64:
		return x, true
	case int:
		return int64(x), true
	case float64:
		return int64(x), true
	case string:
		y, err := strconv.ParseInt(x, 10, 64)
		return y, err == nil
	default:
		return 0, false
	}
}

func asString(v any) (string, bool) {
	if v == nil {
		return "", false
	}
	switch x := v.(type) {
	case string:
		return x, true
	case json.Number:
		return x.String(), true
	default:
		return fmt.Sprintf("%v", x), true
	}
}

func asStringDefault(v any) string {
	s, _ := asString(v)
	return s
}

func pickNested(m map[string]any, keys ...string) any {
	cur := any(m)
	for _, key := range keys {
		mv, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur, ok = mv[key]
		if !ok {
			return nil
		}
	}
	return cur
}

func firstFloat(vals ...any) *float64 {
	for _, v := range vals {
		if f := asFloatPtr(v); f != nil {
			return f
		}
	}
	return nil
}

func floatFromMinutes(v any) any {
	f := asFloatPtr(v)
	if f == nil {
		return nil
	}
	seconds := *f * 60
	return seconds
}
