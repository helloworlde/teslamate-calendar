package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"teslamate-calendar/internal/client"
	"teslamate-calendar/internal/config"
)

func TestAllICSContainsCombinedSections(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/cars/1":
			_, _ = w.Write([]byte(`{"name":"Model 3"}`))
		case "/api/v1/globalsettings":
			_, _ = w.Write([]byte(`{"ok":true}`))
		case "/api/v1/cars/1/drives":
			_, _ = w.Write([]byte(`{"data":{"drives":[{"drive_id":1,"start_date":"2026-04-24T09:33:00+08:00","end_date":"2026-04-24T10:00:00+08:00","start_address":"A","end_address":"B","distance":9.8}]}}`))
		case "/api/v1/cars/1/charges":
			_, _ = w.Write([]byte(`{"data":{"charges":[{"charge_id":2,"start_date":"2026-04-24T21:11:00+08:00","end_date":"2026-04-24T22:11:00+08:00","address":"C","charge_energy_added":12.5}]}}`))
		case "/api/v1/cars/1/updates":
			_, _ = w.Write([]byte(`{"data":{"updates":[{"update_id":3,"start_date":"2026-04-24T11:00:00+08:00","version":"2026.2.11"}]}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer s.Close()

	cl, err := client.New(s.URL, "", "Authorization", "Bearer", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		CalendarFeedEnable:        true,
		DefaultDays:               90,
		MaxDays:                   365,
		DefaultTimezone:           "Asia/Shanghai",
		DefaultView:               "normal",
		DailyIncludeItems:         true,
		CacheTTL:                  time.Minute,
		TeslaMateDashboardBaseURL: "",
	}
	svc := NewCalendarService(cfg, cl)
	ics, err := svc.CalendarICS(context.Background(), "k", "1", CalendarAll, QueryParams{
		Days:     90,
		Timezone: "Asia/Shanghai",
		View:     "normal",
		Detail:   true,
		Range:    "day",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, needle := range []string{"🚗 Model 3", "⚡ Model 3", "软件更新", "📊 Model 3"} {
		if !strings.Contains(ics, needle) {
			t.Fatalf("all.ics missing %s", needle)
		}
	}
}
