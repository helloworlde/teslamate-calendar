package api

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/helloworlde/teslamate-calendar/internal/config"
)

func ConfigureSlog(cfg config.Config) {
	var l slog.Level
	switch strings.ToLower(strings.TrimSpace(cfg.LogLevel)) {
	case "debug":
		l = slog.LevelDebug
	case "warn", "warning":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: l})))
}

func requestLogLevel(cfg config.Config) int {
	switch strings.ToLower(strings.TrimSpace(cfg.LogLevel)) {
	case "error":
		return 1
	case "warn", "warning":
		return 2
	case "info", "":
		return 3
	case "debug":
		return 4
	default:
		return 3
	}
}

func shouldLogStatus(level int, status int) bool {
	switch {
	case level >= 4:
		return true
	case level == 3:
		return true
	case level == 2:
		return status >= 400
	case level == 1:
		return status >= 500
	default:
		return true
	}
}

func redactAuth(val string) string {
	v := strings.TrimSpace(val)
	if v == "" {
		return "none"
	}
	low := strings.ToLower(v)
	if strings.HasPrefix(low, "bearer ") {
		return "bearer(****)"
	}
	if len(v) < 2 {
		return "****"
	}
	return "header(****)"
}

func RequestLog(cfg config.Config) gin.HandlerFunc {
	lvl := requestLogLevel(cfg)
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		st := c.Writer.Status()
		if !shouldLogStatus(lvl, st) {
			return
		}
		d := time.Since(start)
		pathQ := c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			pathQ += "?" + c.Request.URL.RawQuery
		}
		auth := redactAuth(c.GetHeader("Authorization"))
		extra := make([]any, 0, 6)
		if t := c.Param("Token"); t != "" {
			extra = append(extra, "url_token", "set")
		}
		if id := c.Param("CarID"); id != "" {
			extra = append(extra, "car_id", id)
		}
		extra = append(extra,
			"method", c.Request.Method,
			"path", pathQ,
			"status", st,
			"ms", d.Milliseconds(),
			"ip", c.ClientIP(),
			"auth", auth,
		)
		if lvl >= 4 {
			extra = append(extra, "user_agent", strings.TrimSpace(c.GetHeader("User-Agent")))
		}
		if st >= 500 {
			slog.Error("http request", extra...)
		} else if st >= 400 {
			slog.Warn("http request", extra...)
		} else {
			slog.Info("http request", extra...)
		}
	}
}
