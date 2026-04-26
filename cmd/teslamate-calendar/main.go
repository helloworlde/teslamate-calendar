package main

import (
	"log"

	"github.com/helloworlde/teslamate-calendar/internal/api"
	"github.com/helloworlde/teslamate-calendar/internal/client"
	"github.com/helloworlde/teslamate-calendar/internal/config"
	"github.com/helloworlde/teslamate-calendar/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}
	api.ConfigureSlog(cfg)
	c, err := client.New(
		cfg.TeslaMateAPIBaseURL,
		cfg.TeslaMateAPIToken,
		cfg.TeslaMateAPIAuthHeader,
		cfg.TeslaMateAPIAuthScheme,
		cfg.RequestTimeout,
	)
	if err != nil {
		log.Fatalf("create client failed: %v", err)
	}
	svc := service.NewCalendarService(cfg, c)
	h := api.NewHandlers(cfg, svc)
	r := api.NewRouter(cfg, h)
	if err := r.Run(cfg.ListenAddr); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
