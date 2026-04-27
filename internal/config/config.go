package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/helloworlde/teslamate-calendar/internal/util"
)

type Config struct {
	ListenAddr                    string
	TeslaMateAPIBaseURL           string
	TeslaMateAPIToken             string
	TeslaMateAPIAuthHeader        string
	TeslaMateAPIAuthScheme        string
	CalendarFeedToken             string
	DefaultDays                   int
	MaxDays                       int
	DefaultTimezone               string
	DefaultView                   string
	CacheTTL                      time.Duration
	RequestTimeout                time.Duration
	LogLevel                      string
	TeslaMateDashboardURLTemplate string
}

// Load 从环境变量加载配置
func Load() (Config, error) {
	cfg := Config{
		ListenAddr:                    env("LISTEN_ADDR", ":8080"),
		TeslaMateAPIToken:             os.Getenv("TESLAMATE_API_TOKEN"),
		TeslaMateAPIAuthHeader:        env("TESLAMATE_API_AUTH_HEADER", "Authorization"),
		TeslaMateAPIAuthScheme:        env("TESLAMATE_API_AUTH_SCHEME", "Bearer"),
		DefaultDays:                   envInt("DEFAULT_DAYS", 90),
		MaxDays:                       envInt("MAX_DAYS", 365),
		DefaultTimezone:               env("DEFAULT_TIMEZONE", "Asia/Shanghai"),
		DefaultView:                   env("DEFAULT_VIEW", "normal"),
		CacheTTL:                      time.Duration(envInt("CACHE_TTL_SECONDS", 1800)) * time.Second,
		RequestTimeout:                time.Duration(envInt("REQUEST_TIMEOUT_SECONDS", 10)) * time.Second,
		LogLevel:                      env("LOG_LEVEL", "info"),
		TeslaMateDashboardURLTemplate: strings.TrimSpace(os.Getenv("TESLAMATE_DASHBOARD_URL_TEMPLATE")),
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
	tok := strings.TrimSpace(os.Getenv("CALENDAR_FEED_TOKEN"))
	if tok == "" {
		return cfg, errors.New("CALENDAR_FEED_TOKEN is required: set a non-empty secret in the environment (no default)")
	}
	cfg.CalendarFeedToken = tok
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
