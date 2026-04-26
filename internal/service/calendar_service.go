package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"teslamate-calendar/internal/calendar"
	"teslamate-calendar/internal/client"
	"teslamate-calendar/internal/config"
	"teslamate-calendar/internal/model"
	"teslamate-calendar/internal/util"
)

type CalendarType string

const (
	CalendarAll     CalendarType = "all"
	CalendarDrives  CalendarType = "drives"
	CalendarCharges CalendarType = "charges"
	CalendarDaily   CalendarType = "daily"
	CalendarUpdates CalendarType = "updates"
)

type QueryParams struct {
	Days        int
	StartDate   string
	EndDate     string
	Timezone    string
	MinDistance string
	MaxDistance string
	Lang        string
	Detail      bool
	View        string
	VehicleName string
	Range       string
}

type cacheItem struct {
	Value     string
	ExpiresAt time.Time
}

type CalendarService struct {
	cfg    config.Config
	client *client.Client
	mu     sync.RWMutex
	cache  map[string]cacheItem
	group  singleflight.Group
}

func NewCalendarService(cfg config.Config, c *client.Client) *CalendarService {
	return &CalendarService{
		cfg:    cfg,
		client: c,
		cache:  map[string]cacheItem{},
	}
}

func ParseQuery(v url.Values, cfg config.Config) QueryParams {
	days := cfg.DefaultDays
	if raw := strings.TrimSpace(v.Get("days")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			days = n
		}
	}
	tz := strings.TrimSpace(v.Get("timezone"))
	if tz == "" {
		tz = cfg.DefaultTimezone
	}
	detail := true
	if raw := strings.TrimSpace(v.Get("detail")); raw != "" {
		b, err := strconv.ParseBool(raw)
		if err == nil {
			detail = b
		}
	}
	lang := strings.TrimSpace(v.Get("lang"))
	if lang == "" {
		lang = "zh-CN"
	}
	vehicleName := strings.TrimSpace(v.Get("vehicleName"))
	rng := strings.TrimSpace(strings.ToLower(v.Get("range")))
	if rng == "" {
		rng = "day"
	}
	if rng != "day" && rng != "week" && rng != "month" {
		rng = "day"
	}
	view := strings.TrimSpace(strings.ToLower(v.Get("view")))
	if view == "" {
		view = cfg.DefaultView
	}
	if view != "compact" && view != "normal" && view != "detail" {
		view = "normal"
	}
	return QueryParams{
		Days:        days,
		StartDate:   strings.TrimSpace(v.Get("startDate")),
		EndDate:     strings.TrimSpace(v.Get("endDate")),
		Timezone:    tz,
		MinDistance: strings.TrimSpace(v.Get("minDistance")),
		MaxDistance: strings.TrimSpace(v.Get("maxDistance")),
		Lang:        lang,
		Detail:      detail,
		View:        view,
		VehicleName: vehicleName,
		Range:       rng,
	}
}

func (s *CalendarService) CalendarICS(ctx context.Context, key string, carID string, typ CalendarType, q QueryParams) (string, error) {
	if !s.cfg.CalendarFeedEnable {
		return "", errors.New("calendar feed disabled")
	}
	if q.Days > s.cfg.MaxDays {
		return "", fmt.Errorf("days exceeds MAX_DAYS: %d", s.cfg.MaxDays)
	}
	if _, err := strconv.ParseInt(carID, 10, 64); err != nil {
		return "", errors.New("invalid car id")
	}

	if hit, ok := s.getCache(key); ok {
		return hit, nil
	}
	v, err, _ := s.group.Do(key, func() (any, error) {
		if hit, ok := s.getCache(key); ok {
			return hit, nil
		}
		ics, buildErr := s.buildICS(ctx, carID, typ, q)
		if buildErr != nil {
			if stale, ok := s.getStale(key); ok {
				return stale, nil
			}
			return "", buildErr
		}
		s.setCache(key, ics)
		return ics, nil
	})
	if err != nil {
		return "", err
	}
	return v.(string), nil
}

func (s *CalendarService) buildICS(ctx context.Context, carID string, typ CalendarType, q QueryParams) (string, error) {
	gs, _ := s.client.GetGlobalSettings(ctx)
	dash := s.dashboardBase(gs)
	carName := "Tesla"
	if car, err := s.client.GetCar(ctx, carID); err == nil {
		carName = resolveVehicleName(car, q.VehicleName)
	} else if q.VehicleName != "" {
		carName = q.VehicleName
	}
	tr, err := util.BuildTimeRange(q.StartDate, q.EndDate, q.Days, s.cfg.DefaultDays, s.cfg.MaxDays, q.Timezone, time.Now())
	if err != nil {
		return "", err
	}
	query := client.Query{
		StartDate:   tr.StartUTC,
		EndDate:     tr.EndUTC,
		HasStartEnd: true,
		MinDistance: q.MinDistance,
		MaxDistance: q.MaxDistance,
	}
	events := make([]calendar.Event, 0)
	if typ == CalendarAll || typ == CalendarDrives || typ == CalendarDaily {
		drives, err := s.client.ListDrives(ctx, carID, query)
		if err != nil {
			return "", mapUpstreamErr(err)
		}
		if typ == CalendarAll || typ == CalendarDrives {
			events = append(events, calendar.DriveEvents(carID, carName, drives, q.View, q.Detail, dash, s.cfg.TeslaMateDriveDashboardPath)...)
		}
		if typ == CalendarDaily || typ == CalendarAll {
			charges, err := s.client.ListCharges(ctx, carID, query)
			if err != nil {
				return "", mapUpstreamErr(err)
			}
			updates, _ := s.client.ListUpdates(ctx, carID)
			rows := calendar.BuildDailySummaries(drives, charges, updates, tr.Loc)
			if typ == CalendarDaily {
				switch q.Range {
				case "week":
					events = append(events, calendar.WeeklySummaryEvents(carID, carName, rows, tr.Loc, q.View, q.Detail, dash, s.cfg.TeslaMateSummaryDashboardPath)...)
				case "month":
					events = append(events, calendar.MonthlySummaryEvents(carID, carName, rows, tr.Loc, q.View, q.Detail, dash, s.cfg.TeslaMateSummaryDashboardPath)...)
				default:
					events = append(events, calendar.DailySummaryEvents(carID, carName, rows, tr.Loc, q.View, q.Detail, s.cfg.DailyIncludeItems, dash, s.cfg.TeslaMateSummaryDashboardPath)...)
				}
			}
			if typ == CalendarAll {
				switch q.Range {
				case "week":
					events = append(events, calendar.WeeklySummaryEvents(carID, carName, rows, tr.Loc, q.View, q.Detail, dash, s.cfg.TeslaMateSummaryDashboardPath)...)
				case "month":
					events = append(events, calendar.MonthlySummaryEvents(carID, carName, rows, tr.Loc, q.View, q.Detail, dash, s.cfg.TeslaMateSummaryDashboardPath)...)
				default:
					events = append(events, calendar.DailySummaryEvents(carID, carName, rows, tr.Loc, q.View, q.Detail, s.cfg.DailyIncludeItems, dash, s.cfg.TeslaMateSummaryDashboardPath)...)
				}
			}
		}
	}
	if typ == CalendarAll || typ == CalendarCharges {
		charges, err := s.client.ListCharges(ctx, carID, query)
		if err != nil {
			return "", mapUpstreamErr(err)
		}
		events = append(events, calendar.ChargeEvents(carID, carName, charges, q.View, q.Detail, dash, s.cfg.TeslaMateChargeDashboardPath)...)
	}
	if typ == CalendarAll || typ == CalendarUpdates {
		updates, err := s.client.ListUpdates(ctx, carID)
		if err != nil {
			return "", mapUpstreamErr(err)
		}
		events = append(events, calendar.UpdateEvents(carID, carName, updates, q.Detail, dash, s.cfg.TeslaMateUpdateDashboardPath)...)
	}
	return calendar.BuildCalendar(calendarName(carName, typ, q.Range), tr.Loc.String(), events), nil
}

func (s *CalendarService) ListCars(ctx context.Context) ([]model.Car, error) {
	return s.client.ListCars(ctx)
}

func (s *CalendarService) GetCar(ctx context.Context, carID string) (map[string]any, error) {
	return s.client.GetCar(ctx, carID)
}

func (s *CalendarService) UpstreamReady(ctx context.Context) error {
	return s.client.Readyz(ctx)
}

func (s *CalendarService) UpstreamHealth(ctx context.Context) error {
	return s.client.Healthz(ctx)
}

func (s *CalendarService) getCache(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.cache[key]
	if !ok || time.Now().After(v.ExpiresAt) {
		return "", false
	}
	return v.Value, true
}

func (s *CalendarService) getStale(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.cache[key]
	if !ok {
		return "", false
	}
	return v.Value, true
}

func (s *CalendarService) setCache(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[key] = cacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(s.cfg.CacheTTL),
	}
}

func (s *CalendarService) dashboardBase(gs model.GlobalSettings) string {
	if u := strings.TrimSpace(s.cfg.TeslaMateDashboardBaseURL); u != "" {
		return strings.TrimRight(u, "/")
	}
	if u := model.GrafanaBaseURLFromGlobalSettings(gs); u != "" {
		return strings.TrimRight(strings.TrimSpace(u), "/")
	}
	return ""
}

func mapUpstreamErr(err error) error {
	msg := err.Error()
	if strings.Contains(msg, "timeout") {
		return fmt.Errorf("gateway timeout: %w", err)
	}
	return fmt.Errorf("upstream unavailable: %w", err)
}

func calendarName(carName string, typ CalendarType, rng string) string {
	switch typ {
	case CalendarDrives:
		return carName + " · 行程"
	case CalendarCharges:
		return carName + " · 充电"
	case CalendarDaily:
		if rng == "week" {
			return carName + " · 周报"
		}
		if rng == "month" {
			return carName + " · 月报"
		}
		return carName + " · 日报"
	case CalendarUpdates:
		return carName + " · 更新"
	default:
		return carName + " · TeslaMate"
	}
}

func stringAny(v any) string {
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(s)
}

func resolveVehicleName(car map[string]any, override string) string {
	if strings.TrimSpace(override) != "" {
		return strings.TrimSpace(override)
	}
	candidates := []string{
		stringAny(car["name"]),
		stringAny(car["display_name"]),
		stringAny(car["displayName"]),
		stringAny(car["vehicle_name"]),
		stringAny(car["vehicleName"]),
		stringAny(car["car_name"]),
		stringAny(car["carName"]),
		stringAny(car["model"]),
	}
	for _, name := range candidates {
		if name != "" {
			return name
		}
	}
	vin := stringAny(car["vin"])
	if len(vin) >= 6 {
		return "Tesla-" + vin[len(vin)-6:]
	}
	return "Tesla"
}
