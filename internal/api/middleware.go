package api

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/helloworlde/teslamate-calendar/internal/config"
)

// ConstantTimeEquals 使用常量时间比较两个字符串，防止时序攻击
func ConstantTimeEquals(a, b string) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// RequireCalendarPathToken 验证日历订阅 URL 中的 token
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
