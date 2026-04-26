package api

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/helloworlde/teslamate-calendar/internal/config"
)

func ConstantTimeEquals(a, b string) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func RequireCalendarPathToken(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		tok := strings.TrimSpace(c.Param("Token"))
		if !ConstantTimeEquals(tok, cfg.CalendarFeedToken) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Next()
	}
}
