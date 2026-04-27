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

	"github.com/helloworlde/teslamate-calendar/internal/calendar"
	"github.com/helloworlde/teslamate-calendar/internal/client"
	"github.com/helloworlde/teslamate-calendar/internal/config"
	"github.com/helloworlde/teslamate-calendar/internal/model"
	"github.com/helloworlde/teslamate-calendar/internal/util"
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
	Detail      bool
	View        string
	VehicleName string
	Range       string
}

type cacheItem struct {
	Value     string
	ExpiresAt time.Time
}

// CalendarService 日历服务，负责生成 iCalendar 格式的日历数据
type CalendarService struct {
	cfg    config.Config
	client *client.Client
	mu     sync.RWMutex
	cache  map[string]cacheItem
	group  singleflight.Group // 防止缓存击穿
}

func NewCalendarService(cfg config.Config, c *client.Client) *CalendarService {
	return &CalendarService{
		cfg:    cfg,
		client: c,
		cache:  map[string]cacheItem{},
	}
}

// ParseQuery 解析 URL 查询参数并返回标准化的查询参数
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
		Detail:      detail,
		View:        view,
		VehicleName: vehicleName,
		Range:       rng,
	}
}

// CalendarICS 生成指定类型的 iCalendar 数据，支持缓存和防缓存击穿
func (s *CalendarService) CalendarICS(ctx context.Context, key string, carID string, typ CalendarType, q QueryParams) (string, error) {
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
	tmpl := s.cfg.TeslaMateDashboardURLTemplate
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
			events = append(events, calendar.DriveEvents(carID, carName, drives, q.View, q.Detail, tr.Loc, tmpl)...)
		}
		if typ == CalendarDaily || typ == CalendarAll {
			charges, err := s.client.ListCharges(ctx, carID, query)
			if err != nil {
				return "", mapUpstreamErr(err)
			}
			updates, _ := s.client.ListUpdates(ctx, carID)
			rows := calendar.BuildDailySummaries(drives, charges, updates, tr.Loc)
			events = append(events, s.buildSummaryEvents(carID, carName, rows, tr.Loc, q, tmpl)...)
		}
	}
	if typ == CalendarAll || typ == CalendarCharges {
		charges, err := s.client.ListCharges(ctx, carID, query)
		if err != nil {
			return "", mapUpstreamErr(err)
		}
		events = append(events, calendar.ChargeEvents(carID, carName, charges, q.View, q.Detail, tr.Loc, tmpl)...)
	}
	if typ == CalendarAll || typ == CalendarUpdates {
		updates, err := s.client.ListUpdates(ctx, carID)
		if err != nil {
			return "", mapUpstreamErr(err)
		}
		events = append(events, calendar.UpdateEvents(carID, carName, updates, q.Detail, tr.Loc, tmpl)...)
	}
	return calendar.BuildCalendar(calendarName(carName, typ, q.Range), tr.Loc.String(), events), nil
}

func (s *CalendarService) ListCars(ctx context.Context) ([]model.Car, error) {
	return s.client.ListCars(ctx)
}

func (s *CalendarService) GetCar(ctx context.Context, carID string) (model.CarProfile, error) {
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

// buildSummaryEvents 根据范围类型构建汇总事件（日报/周报/月报）
func (s *CalendarService) buildSummaryEvents(carID, carName string, rows []model.DailySummary, loc *time.Location, q QueryParams, tmpl string) []calendar.Event {
	switch q.Range {
	case "week":
		return calendar.WeeklySummaryEvents(carID, carName, rows, loc, q.View, q.Detail, tmpl)
	case "month":
		return calendar.MonthlySummaryEvents(carID, carName, rows, loc, q.View, q.Detail, tmpl)
	default:
		return calendar.DailySummaryEvents(carID, carName, rows, loc, q.View, q.Detail, tmpl)
	}
}

func resolveVehicleName(car model.CarProfile, override string) string {
	if strings.TrimSpace(override) != "" {
		return strings.TrimSpace(override)
	}
	candidates := []string{
		strings.TrimSpace(car.Name),
		strings.TrimSpace(car.DisplayName),
		strings.TrimSpace(car.Model),
	}
	for _, name := range candidates {
		if name != "" {
			return name
		}
	}
	vin := strings.TrimSpace(car.VIN)
	if len(vin) >= 6 {
		return "Tesla-" + vin[len(vin)-6:]
	}
	return "Tesla"
}
