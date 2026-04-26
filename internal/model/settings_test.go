package model

import "testing"

func TestGrafanaBaseURLFromGlobalSettings(t *testing.T) {
	gs := GlobalSettings{GrafanaURL: "http://grafana.lan:3000"}
	if s := GrafanaBaseURLFromGlobalSettings(gs); s != "http://grafana.lan:3000" {
		t.Fatalf("got %q", s)
	}
}
