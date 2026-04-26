package service

import "testing"

func TestResolveVehicleNamePriority(t *testing.T) {
	car := map[string]any{
		"name":         "Model 3",
		"display_name": "DisplayName",
		"car_name":     "CarName",
	}
	if got := resolveVehicleName(car, ""); got != "Model 3" {
		t.Fatalf("unexpected name: %s", got)
	}
}

func TestResolveVehicleNameFallbackAndOverride(t *testing.T) {
	car := map[string]any{
		"display_name": "Display",
		"vin":          "LRWABCDEFG1234567",
	}
	if got := resolveVehicleName(car, "我的ModelY"); got != "我的ModelY" {
		t.Fatalf("override not applied: %s", got)
	}
	car2 := map[string]any{"vin": "LRWABCDEFG1234567"}
	if got := resolveVehicleName(car2, ""); got != "Tesla-234567" {
		t.Fatalf("vin fallback unexpected: %s", got)
	}
	if got := resolveVehicleName(map[string]any{}, ""); got != "Tesla" {
		t.Fatalf("final fallback unexpected: %s", got)
	}
}

func TestCalendarNameNoCarID(t *testing.T) {
	cal := calendarName("Model 3", CalendarDaily, "day")
	if cal != "Model 3 · 日报" {
		t.Fatalf("unexpected calname: %s", cal)
	}
}
