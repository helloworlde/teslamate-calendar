package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"teslamate-calendar/internal/client"
	"teslamate-calendar/internal/config"
	"teslamate-calendar/internal/service"
)

func TestDocsEndpoints(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer upstream.Close()

	cl, err := client.New(upstream.URL, "", "Authorization", "Bearer", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		CalendarFeedEnable: true,
		DefaultDays:        90,
		MaxDays:            365,
		DefaultTimezone:    "Asia/Shanghai",
		CacheTTL:           time.Minute,
		DefaultView:        "normal",
		MapProvider:        "google",
	}
	r := NewRouter(cfg, NewHandlers(cfg, service.NewCalendarService(cfg, cl)))

	for _, p := range []string{"/openapi.json", "/swagger/index.html", "/scalar"} {
		req := httptest.NewRequest(http.MethodGet, p, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%s expected 200 got %d", p, w.Code)
		}
	}
}

func TestOpenAPISpecContent(t *testing.T) {
	var root map[string]any
	if err := json.Unmarshal([]byte(BuildOpenAPISpec(config.Config{
		CalendarFeedToken: "tesla",
		DefaultDays:       90,
		DefaultTimezone:   "Asia/Shanghai",
		DefaultView:       "normal",
	})), &root); err != nil {
		t.Fatal(err)
	}
	if root["openapi"] != "3.0.3" {
		t.Fatalf("openapi: %v", root["openapi"])
	}
	comps, _ := root["components"].(map[string]any)
	if comps["schemas"] == nil || comps["parameters"] == nil || comps["responses"] == nil {
		t.Fatalf("expected components: schemas, parameters, responses")
	}
	paths, _ := root["paths"].(map[string]any)
	if paths["/healthz"] == nil || paths["/calendar/token/{Token}/cars/{CarID}/daily.ics"] == nil {
		t.Fatalf("expected core paths in spec")
	}
}
