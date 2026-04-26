package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"teslamate-calendar/internal/util"
)

type Config struct {
	ListenAddr                    string
	TeslaMateAPIBaseURL           string
	TeslaMateAPIToken             string
	TeslaMateAPIAuthHeader        string
	TeslaMateAPIAuthScheme        string
	CalendarFeedToken             string
	CalendarFeedEnable            bool
	RequireTokenForCars           bool
	DefaultDays                   int
	MaxDays                       int
	DefaultTimezone               string
	DefaultView                   string
	DailyIncludeItems             bool
	CacheTTL                      time.Duration
	RequestTimeout                time.Duration
	LogLevel                      string
	PublicBaseURL                 string
	MapProvider                   string
	TeslaMateDashboardBaseURL     string
	TeslaMateDriveDashboardPath   string
	TeslaMateChargeDashboardPath  string
	TeslaMateSummaryDashboardPath string
	TeslaMateUpdateDashboardPath  string
}

func Load() (Config, error) {
	cfg := Config{
		ListenAddr:                    env("LISTEN_ADDR", ":8080"),
		TeslaMateAPIToken:             os.Getenv("TESLAMATE_API_TOKEN"),
		TeslaMateAPIAuthHeader:        env("TESLAMATE_API_AUTH_HEADER", "Authorization"),
		TeslaMateAPIAuthScheme:        env("TESLAMATE_API_AUTH_SCHEME", "Bearer"),
		CalendarFeedToken:             env("CALENDAR_FEED_TOKEN", "tesla"),
		CalendarFeedEnable:            envBool("CALENDAR_FEED_ENABLE", true),
		RequireTokenForCars:           envBool("REQUIRE_TOKEN_FOR_CARS", true),
		DefaultDays:                   envInt("DEFAULT_DAYS", 90),
		MaxDays:                       envInt("MAX_DAYS", 365),
		DefaultTimezone:               env("DEFAULT_TIMEZONE", "Asia/Shanghai"),
		DefaultView:                   env("DEFAULT_VIEW", "normal"),
		DailyIncludeItems:             envBool("DAILY_INCLUDE_ITEMS", true),
		CacheTTL:                      time.Duration(envInt("CACHE_TTL_SECONDS", 1800)) * time.Second,
		RequestTimeout:                time.Duration(envInt("REQUEST_TIMEOUT_SECONDS", 10)) * time.Second,
		LogLevel:                      env("LOG_LEVEL", "info"),
		PublicBaseURL:                 strings.TrimRight(os.Getenv("PUBLIC_BASE_URL"), "/"),
		MapProvider:                   env("MAP_PROVIDER", "google"),
		TeslaMateDashboardBaseURL:     strings.TrimRight(os.Getenv("TESLAMATE_DASHBOARD_BASE_URL"), "/"),
		TeslaMateDriveDashboardPath:   os.Getenv("TESLAMATE_DRIVE_DASHBOARD_PATH"),
		TeslaMateChargeDashboardPath:  os.Getenv("TESLAMATE_CHARGE_DASHBOARD_PATH"),
		TeslaMateSummaryDashboardPath: os.Getenv("TESLAMATE_SUMMARY_DASHBOARD_PATH"),
		TeslaMateUpdateDashboardPath:  os.Getenv("TESLAMATE_UPDATE_DASHBOARD_PATH"),
	}
	if raw := strings.TrimSpace(os.Getenv("TESLAMATE_API_BASE_URL")); raw == "" {
		return cfg, errors.New("TESLAMATE_API_BASE_URL is required")
	} else {
		n, err := util.NormalizeTeslaMateAPIBase(raw)
		if err != nil {
			return cfg, err
		}
		cfg.TeslaMateAPIBaseURL = n
	}
	if cfg.DefaultDays < 1 {
		cfg.DefaultDays = 90
	}
	if cfg.MaxDays < 1 {
		cfg.MaxDays = 365
	}
	if cfg.DefaultDays > cfg.MaxDays {
		cfg.DefaultDays = cfg.MaxDays
	}
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 10 * time.Second
	}
	if cfg.CacheTTL <= 0 {
		cfg.CacheTTL = 1800 * time.Second
	}
	if cfg.DefaultTimezone == "" {
		cfg.DefaultTimezone = "Asia/Shanghai"
	}
	if cfg.DefaultView != "compact" && cfg.DefaultView != "normal" && cfg.DefaultView != "detail" {
		cfg.DefaultView = "normal"
	}
	if cfg.MapProvider == "" {
		cfg.MapProvider = "google"
	}
	if _, err := time.LoadLocation(cfg.DefaultTimezone); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func env(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func envInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}

func envBool(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return v
}
