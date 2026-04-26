package model

import "testing"

func TestGrafanaBaseURLFromGlobalSettings(t *testing.T) {
	gs := GlobalSettings{Raw: map[string]any{
		"teslamate_urls": map[string]any{
			"grafana_url": "http://grafana.lan:3000",
		},
	}}
	if s := GrafanaBaseURLFromGlobalSettings(gs); s != "http://grafana.lan:3000" {
		t.Fatalf("got %q", s)
	}
}
