package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Setenv("TESLAMATE_API_BASE_URL", "http://teslamateapi:8080/api")
	t.Setenv("CALENDAR_FEED_TOKEN", "test-token-not-default")
	t.Setenv("DEFAULT_TIMEZONE", "Asia/Shanghai")
	t.Setenv("DEFAULT_DAYS", "120")
	t.Setenv("MAX_DAYS", "366")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.TeslaMateAPIBaseURL != "http://teslamateapi:8080" {
		t.Fatalf("expected host-only base, got %q", cfg.TeslaMateAPIBaseURL)
	}
	if cfg.DefaultDays != 120 {
		t.Fatalf("unexpected DefaultDays: %d", cfg.DefaultDays)
	}
	if cfg.MaxDays != 366 {
		t.Fatalf("unexpected MaxDays: %d", cfg.MaxDays)
	}
	if cfg.CalendarFeedToken != "test-token-not-default" {
		t.Fatalf("token: %q", cfg.CalendarFeedToken)
	}
}

func TestLoadRequiresCalendarFeedToken(t *testing.T) {
	t.Setenv("TESLAMATE_API_BASE_URL", "http://a:1")
	t.Cleanup(func() { _ = os.Unsetenv("CALENDAR_FEED_TOKEN") })
	_ = os.Unsetenv("CALENDAR_FEED_TOKEN")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error when CALENDAR_FEED_TOKEN is empty")
	}
}

func TestGoModModulePath(t *testing.T) {
	_, f, _, _ := runtime.Caller(0)
	repo := filepath.Clean(filepath.Join(filepath.Dir(f), "../.."))
	b, err := os.ReadFile(filepath.Join(repo, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	first := strings.SplitN(string(b), "\n", 2)[0]
	if first != "module github.com/helloworlde/teslamate-calendar" {
		t.Fatalf("unexpected first line: %q", first)
	}
}
