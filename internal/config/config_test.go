package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Setenv("TESLAMATE_API_BASE_URL", "http://teslamateapi:8080/api")
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
	if cfg.CalendarFeedToken != "tesla" {
		t.Fatalf("default token: %q", cfg.CalendarFeedToken)
	}
	_ = os.Unsetenv("TESLAMATE_API_BASE_URL")
}
