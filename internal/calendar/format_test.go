package calendar

import (
	"strings"
	"testing"
	"time"

	"github.com/helloworlde/teslamate-calendar/internal/model"
)

func TestFormatDurationZH(t *testing.T) {
	if got := FormatDurationZH(225 * 60); got != "3小时45分钟" {
		t.Fatalf("225 min as seconds: got %q", got)
	}
	if got := FormatDurationZH(40 * 60); got != "40分钟" {
		t.Fatalf("40 min: got %q", got)
	}
}

func TestFormatTimeRangeCrossDay(t *testing.T) {
	loc := time.Local
	s := time.Date(2026, 4, 11, 21, 1, 0, 0, loc)
	e := time.Date(2026, 4, 12, 1, 49, 0, 0, loc)
	if r := FormatTimeRange(s, e, loc); r != "21:01-次日 01:49" {
		t.Fatalf("got %q", r)
	}
}

func TestNormalizePlaceName(t *testing.T) {
	if g := NormalizePlaceName("苏宁易购, 海淀区"); g != "苏宁易购（海淀区）" {
		t.Fatalf("got %q", g)
	}
}

func TestDriveDescriptionNoPolylineLeak(t *testing.T) {
	long := strings.Repeat("A", 100)
	polyline := "abcdefghijklmnopqrstuvwxyz"
	d := model.Drive{
		StartDate:    ptrTime(time.Date(2026, 4, 24, 9, 33, 0, 0, time.Local)),
		EndDate:      ptrTime(time.Date(2026, 4, 24, 10, 0, 0, 0, time.Local)),
		HasRoute:     true,
		StartAddress: "A",
		EndAddress:   "B",
	}
	desc := BuildDriveDescription("Model 3", d, time.Local)
	if strings.Contains(desc, long) || strings.Contains(desc, polyline) {
		t.Fatalf("raw polyline leaked: %s", desc)
	}
}

func TestDriveDescriptionRouteLineWhenPolylineSet(t *testing.T) {
	d := model.Drive{
		StartDate:    ptrTime(time.Date(2026, 4, 24, 9, 33, 0, 0, time.Local)),
		EndDate:      ptrTime(time.Date(2026, 4, 24, 10, 0, 0, 0, time.Local)),
		HasRoute:     true,
		StartAddress: "A",
		EndAddress:   "B",
	}
	desc := BuildDriveDescription("M3", d, time.Local)
	if !strings.Contains(desc, "轨迹：已记录，可在 TeslaMate 看板查看") {
		t.Fatalf("missing trajectory line: %s", desc)
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
