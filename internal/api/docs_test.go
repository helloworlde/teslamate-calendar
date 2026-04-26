package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/helloworlde/teslamate-calendar/internal/client"
	"github.com/helloworlde/teslamate-calendar/internal/config"
	"github.com/helloworlde/teslamate-calendar/internal/service"
)

func testAPIConfig() config.Config {
	return config.Config{
		CalendarFeedToken:   "s3cret-test",
		TeslaMateAPIBaseURL: "http://a",
		DefaultDays:         90,
		MaxDays:             365,
		DefaultTimezone:     "Asia/Shanghai",
		CacheTTL:            time.Minute,
		DefaultView:         "normal",
	}
}

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
	cfg := testAPIConfig()
	cfg.TeslaMateAPIBaseURL = upstream.URL
	r := NewRouter(cfg, NewHandlers(cfg, service.NewCalendarService(cfg, cl)))

	for _, p := range []string{"/openapi.json", "/swagger/index.html", "/scalar"} {
		req := httptest.NewRequest(http.MethodGet, p, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%s expected 200 got %d", p, w.Code)
		}
		if p == "/swagger/index.html" && strings.Contains(w.Body.String(), "preauthorize") {
			t.Fatal("swagger must not preauthorize")
		}
	}
}

func TestOpenAPISpecNoLiveToken(t *testing.T) {
	raw := BuildOpenAPISpec()
	if strings.Contains(raw, "s3cret") {
		t.Fatalf("spec must not echo deployment tokens")
	}
	if !strings.Contains(raw, openAPIPlaceholderToken) {
		t.Fatal("expected placeholder in spec")
	}
	var root map[string]any
	if err := json.Unmarshal([]byte(raw), &root); err != nil {
		t.Fatal(err)
	}
}

func TestOpenAPIPaths(t *testing.T) {
	var root map[string]any
	_ = json.Unmarshal([]byte(BuildOpenAPISpec()), &root)
	paths, _ := root["paths"].(map[string]any)
	if paths["/healthz"] == nil || paths["/calendar/token/{Token}/cars/{CarID}/daily.ics"] == nil {
		t.Fatalf("expected core paths")
	}
}
