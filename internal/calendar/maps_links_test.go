package calendar

import (
	"strings"
	"testing"
	"time"

	"teslamate-calendar/internal/model"
)

func TestGoogleMapsURLs(t *testing.T) {
	a, b, c, d := 39.1, 116.3, 39.2, 116.4
	route := GoogleRouteURL(&a, &b, &c, &d)
	if !strings.Contains(route, "maps/dir") {
		t.Fatalf("route url invalid: %s", route)
	}
	point := GooglePointURL(&a, &b)
	if !strings.Contains(point, "maps/search") {
		t.Fatalf("point url invalid: %s", point)
	}
	search := GoogleSearchURL("蓟门里东区")
	if !strings.Contains(search, "query=") {
		t.Fatalf("search url invalid: %s", search)
	}
}

func TestDashboardURL(t *testing.T) {
	start := time.Date(2026, 4, 24, 9, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	u := DashboardURL("https://grafana.example.com", "/d/tesla/drives", "1", start, end)
	if !strings.Contains(u, "var-car_id=1") || !strings.Contains(u, "from=") {
		t.Fatalf("dashboard url invalid: %s", u)
	}
}

func TestLinksSection(t *testing.T) {
	desc := "hello"
	with := AppendLinksSection(desc, "https://maps.google.com/a", "https://grafana.example.com")
	if !strings.Contains(with, "🔗 链接") || !strings.Contains(with, "地图：") {
		t.Fatalf("links section missing: %s", with)
	}
	without := AppendLinksSection(desc, "", "")
	if without != desc {
		t.Fatalf("unexpected links section when empty")
	}
}

func TestDriveAndChargeEventsContainLinks(t *testing.T) {
	start := time.Date(2026, 4, 24, 9, 0, 0, 0, time.UTC)
	end := start.Add(30 * time.Minute)
	lat1, lng1, lat2, lng2 := 39.1, 116.3, 39.2, 116.4
	d := []model.Drive{{
		ID:           1,
		StartDate:    &start,
		EndDate:      &end,
		StartAddress: "A",
		EndAddress:   "B",
		StartLat:     &lat1,
		StartLng:     &lng1,
		EndLat:       &lat2,
		EndLng:       &lng2,
	}}
	ev := DriveEvents("1", "Model 3", d, "normal", true, "https://grafana.example.com", "/d/drives")
	if len(ev) != 1 || !strings.Contains(ev[0].Description, "🔗 链接") {
		t.Fatalf("drive links missing")
	}

	c := []model.Charge{{
		ID:        2,
		StartDate: &start,
		EndDate:   &end,
		Address:   "C",
		Lat:       &lat1,
		Lng:       &lng1,
	}}
	ec := ChargeEvents("1", "Model 3", c, "normal", true, "https://grafana.example.com", "/d/charges")
	if len(ec) != 1 || !strings.Contains(ec[0].Description, "🔗 链接") {
		t.Fatalf("charge links missing")
	}
}
