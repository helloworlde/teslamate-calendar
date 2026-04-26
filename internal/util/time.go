package util

import (
	"errors"
	"fmt"
	"net/url"
	"time"
)

type TimeRange struct {
	StartUTC time.Time
	EndUTC   time.Time
	Days     int
	Loc      *time.Location
}

func BuildTimeRange(startRaw, endRaw string, days, defaultDays, maxDays int, timezone string, now time.Time) (TimeRange, error) {
	if days <= 0 {
		days = defaultDays
	}
	if days > maxDays {
		return TimeRange{}, fmt.Errorf("days exceeds MAX_DAYS: %d", maxDays)
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return TimeRange{}, errors.New("invalid timezone")
	}
	var start, end time.Time
	if startRaw != "" {
		start, err = ParseFlexibleTime(startRaw, loc)
		if err != nil {
			return TimeRange{}, fmt.Errorf("invalid startDate: %w", err)
		}
	}
	if endRaw != "" {
		end, err = ParseFlexibleTime(endRaw, loc)
		if err != nil {
			return TimeRange{}, fmt.Errorf("invalid endDate: %w", err)
		}
	}
	if startRaw == "" && endRaw == "" {
		end = now.UTC()
		start = end.AddDate(0, 0, -days)
	}
	if startRaw == "" && endRaw != "" {
		start = end.AddDate(0, 0, -days)
	}
	if startRaw != "" && endRaw == "" {
		end = now.UTC()
	}
	if end.Before(start) {
		return TimeRange{}, errors.New("endDate before startDate")
	}
	return TimeRange{
		StartUTC: start.UTC(),
		EndUTC:   end.UTC(),
		Days:     days,
		Loc:      loc,
	}, nil
}

func ParseFlexibleTime(raw string, loc *time.Location) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t, nil
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
		"2006/01/02 15:04:05",
		"2006/01/02 15:04",
		"2006/01/02",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, raw, loc); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("unsupported time format")
}

func RFC3339UTC(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func AddQueryTime(v url.Values, key string, t time.Time) {
	v.Set(key, RFC3339UTC(t))
}
