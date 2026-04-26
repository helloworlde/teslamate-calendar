package service

import (
	"testing"

	"github.com/helloworlde/teslamate-calendar/internal/model"
)

func TestResolveVehicleNamePriority(t *testing.T) {
	car := model.CarProfile{
		Name:        "Model 3",
		DisplayName: "DisplayName",
		Model:       "Model",
	}
	if got := resolveVehicleName(car, ""); got != "Model 3" {
		t.Fatalf("unexpected name: %s", got)
	}
}

func TestResolveVehicleNameFallbackAndOverride(t *testing.T) {
	car := model.CarProfile{
		DisplayName: "Display",
		VIN:         "LRWABCDEFG1234567",
	}
	if got := resolveVehicleName(car, "我的ModelY"); got != "我的ModelY" {
		t.Fatalf("override not applied: %s", got)
	}
	car2 := model.CarProfile{VIN: "LRWABCDEFG1234567"}
	if got := resolveVehicleName(car2, ""); got != "Tesla-234567" {
		t.Fatalf("vin fallback unexpected: %s", got)
	}
	if got := resolveVehicleName(model.CarProfile{}, ""); got != "Tesla" {
		t.Fatalf("final fallback unexpected: %s", got)
	}
}

func TestCalendarNameNoCarID(t *testing.T) {
	cal := calendarName("Model 3", CalendarDaily, "day")
	if cal != "Model 3 · 日报" {
		t.Fatalf("unexpected calname: %s", cal)
	}
}
