package service

import (
	"testing"

	"teslamate-calendar/internal/config"
	"teslamate-calendar/internal/model"
)

func TestDashboardBaseEnvOverride(t *testing.T) {
	s := &CalendarService{cfg: config.Config{TeslaMateDashboardBaseURL: "https://a.example"}}
	if s.dashboardBase(model.GlobalSettings{}) != "https://a.example" {
		t.Fatalf("env override: %q", s.dashboardBase(model.GlobalSettings{}))
	}
}

func TestDashboardBaseFromGrafanaInSettings(t *testing.T) {
	s := &CalendarService{cfg: config.Config{}}
	gs := model.GlobalSettings{Raw: map[string]any{
		"teslamate_urls": map[string]any{
			"grafana_url": "http://g.example/",
		},
	}}
	if got := s.dashboardBase(gs); got != "http://g.example" {
		t.Fatalf("got %q", got)
	}
}
