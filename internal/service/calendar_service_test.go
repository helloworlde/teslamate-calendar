package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"teslamate-calendar/internal/client"
	"teslamate-calendar/internal/config"
)

func newTestService(t *testing.T) *CalendarService {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/globalsettings":
			_, _ = w.Write([]byte(`{"distance_unit":"km"}`))
		case "/api/v1/cars/1/drives":
			_, _ = w.Write([]byte(`[]`))
		case "/api/v1/cars/1/charges":
			_, _ = w.Write([]byte(`[]`))
		case "/api/v1/cars/1/updates":
			_, _ = w.Write([]byte(`[]`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(s.Close)
	c, err := client.New(s.URL, "", "Authorization", "Bearer", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		TeslaMateAPIBaseURL: s.URL,
		CalendarFeedEnable:  true,
		DefaultDays:         90,
		MaxDays:             365,
		DefaultTimezone:     "Asia/Shanghai",
		CacheTTL:            30 * time.Minute,
	}
	return NewCalendarService(cfg, c)
}

func TestCacheKeySeparatedByURL(t *testing.T) {
	svc := newTestService(t)
	q := QueryParams{Days: 90, Timezone: "Asia/Shanghai", Detail: true}
	_, err := svc.CalendarICS(context.Background(), "/calendar/cars/1/drives.ics?days=90", "1", CalendarDrives, q)
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.CalendarICS(context.Background(), "/calendar/cars/1/drives.ics?days=30", "1", CalendarDrives, QueryParams{Days: 30, Timezone: "Asia/Shanghai", Detail: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(svc.cache) != 2 {
		t.Fatalf("expected 2 cache keys, got %d", len(svc.cache))
	}
}

func TestDaysExceedMaxReturnsError(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.CalendarICS(context.Background(), "k", "1", CalendarDrives, QueryParams{
		Days:     400,
		Timezone: "Asia/Shanghai",
		Detail:   true,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
