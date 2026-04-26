package api

import (
	"github.com/gin-gonic/gin"

	"github.com/helloworlde/teslamate-calendar/internal/config"
	"github.com/helloworlde/teslamate-calendar/internal/service"
)

func NewRouter(cfg config.Config, h *Handlers) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), RequestLog(cfg))

	r.GET("/healthz", h.Healthz)
	r.GET("/readyz", h.Readyz)
	r.GET("/ping", h.Ping)
	r.GET("/openapi.json", h.OpenAPIJSON)
	r.GET("/swagger/index.html", h.SwaggerIndex)
	r.GET("/scalar", h.Scalar)

	r.GET("/cars", h.Cars)
	r.GET("/cars/:CarID", h.Car)

	cal := r.Group("/calendar/token/:Token")
	cal.Use(RequireCalendarPathToken(cfg))
	cal.GET("/cars/:CarID/all.ics", h.Calendar(service.CalendarAll))
	cal.GET("/cars/:CarID/drives.ics", h.Calendar(service.CalendarDrives))
	cal.GET("/cars/:CarID/charges.ics", h.Calendar(service.CalendarCharges))
	cal.GET("/cars/:CarID/daily.ics", h.Calendar(service.CalendarDaily))
	cal.GET("/cars/:CarID/updates.ics", h.Calendar(service.CalendarUpdates))

	return r
}
