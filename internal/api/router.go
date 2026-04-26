package api

import (
	"github.com/gin-gonic/gin"

	"teslamate-calendar/internal/config"
	"teslamate-calendar/internal/service"
)

func NewRouter(cfg config.Config, h *Handlers) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), RequestLog(cfg))

	r.GET("/healthz", h.Healthz)
	r.GET("/readyz", h.Readyz)
	r.GET("/ping", h.Ping)
	r.GET("/openapi.json", h.OpenAPIJSON)
	r.GET("/swagger/doc.json", h.SwaggerDocJSON)
	r.GET("/swagger/index.html", h.SwaggerIndex)
	r.GET("/scalar", h.Scalar)

	cars := r.Group("/")
	if cfg.RequireTokenForCars {
		cars.Use(RequireToken(cfg, false))
	}
	cars.GET("/cars", h.Cars)
	cars.GET("/cars/:CarID", h.Car)

	calendar := r.Group("/calendar")
	{
		byURL := calendar.Group("/token/:Token")
		byURL.Use(RequireToken(cfg, true))
		byURL.GET("/cars/:CarID/all.ics", h.Calendar(service.CalendarAll))
		byURL.GET("/cars/:CarID/drives.ics", h.Calendar(service.CalendarDrives))
		byURL.GET("/cars/:CarID/charges.ics", h.Calendar(service.CalendarCharges))
		byURL.GET("/cars/:CarID/daily.ics", h.Calendar(service.CalendarDaily))
		byURL.GET("/cars/:CarID/updates.ics", h.Calendar(service.CalendarUpdates))
	}

	return r
}
