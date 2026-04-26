package api

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"teslamate-calendar/internal/config"
)

func ConstantTimeEquals(a, b string) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func RequireToken(cfg config.Config, fromURL bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.CalendarFeedToken == "" {
			c.Next()
			return
		}
		var token string
		if fromURL {
			token = c.Param("Token")
		} else {
			auth := c.GetHeader("Authorization")
			if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
				token = strings.TrimSpace(auth[7:])
			} else {
				token = strings.TrimSpace(auth)
			}
		}
		if !ConstantTimeEquals(token, cfg.CalendarFeedToken) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Next()
	}
}
