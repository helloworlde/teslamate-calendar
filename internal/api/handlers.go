package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/helloworlde/teslamate-calendar/internal/config"
	"github.com/helloworlde/teslamate-calendar/internal/service"
)

type Handlers struct {
	cfg config.Config
	svc *service.CalendarService
}

func NewHandlers(cfg config.Config, svc *service.CalendarService) *Handlers {
	return &Handlers{cfg: cfg, svc: svc}
}

func (h *Handlers) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handlers) Readyz(c *gin.Context) {
	if err := h.svc.UpstreamReady(c.Request.Context()); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"status": "not ready", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

func (h *Handlers) Ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func (h *Handlers) Cars(c *gin.Context) {
	rows, err := h.svc.ListCars(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

func (h *Handlers) Car(c *gin.Context) {
	carID := c.Param("CarID")
	obj, err := h.svc.GetCar(c.Request.Context(), carID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, obj)
}

func (h *Handlers) Calendar(typ service.CalendarType) gin.HandlerFunc {
	return func(c *gin.Context) {
		carID := c.Param("CarID")
		params := service.ParseQuery(c.Request.URL.Query(), h.cfg)
		key := c.Request.URL.RequestURI()
		ics, err := h.svc.CalendarICS(c.Request.Context(), key, carID, typ, params)
		if err != nil {
			msg := err.Error()
			switch {
			case strings.Contains(msg, "invalid car id"):
				c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			case strings.Contains(msg, "days exceeds"), strings.Contains(msg, "invalid timezone"), strings.Contains(msg, "invalid startDate"), strings.Contains(msg, "invalid endDate"):
				c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			case strings.Contains(msg, "gateway timeout"):
				c.JSON(http.StatusGatewayTimeout, gin.H{"error": msg})
			default:
				c.JSON(http.StatusBadGateway, gin.H{"error": msg})
			}
			return
		}
		c.Header("Content-Type", "text/calendar; charset=utf-8")
		c.Header("Content-Disposition", `inline; filename="teslamate-calendar.ics"`)
		c.Header("Cache-Control", "public, max-age="+strconv.Itoa(int(h.cfg.CacheTTL.Seconds())))
		c.String(http.StatusOK, ics)
	}
}
