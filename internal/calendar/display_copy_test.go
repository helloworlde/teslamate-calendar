package calendar

import (
	"strings"
	"testing"
	"time"

	"github.com/helloworlde/teslamate-calendar/internal/model"
)

func TestDriveDescriptionKeySections(t *testing.T) {
	s := time.Date(2026, 4, 24, 9, 33, 0, 0, time.UTC)
	e := s.Add(27 * time.Minute)
	d1 := 9.8
	c100 := 127.0
	m80 := 80.0
	p83, p81 := 83.0, 81.0
	d := model.Drive{
		StartDate:         &s,
		EndDate:           &e,
		Distance:          &d1,
		Consumption:       &c100,
		MaxSpeed:          &m80,
		StartAddress:      "蓟门里东区",
		EndAddress:        "城奥大厦附近",
		StartBatteryLevel: &p83,
		EndBatteryLevel:   &p81,
	}
	desc := BuildDriveDescription("Model 3", d, time.UTC)
	for _, sec := range []string{"🚗 路线", "📊 数据", "📍 位置"} {
		if !strings.Contains(desc, sec) {
			t.Fatalf("missing %s: %s", sec, desc)
		}
	}
	if strings.Contains(desc, "行程ID") {
		t.Fatal("id leak")
	}
}

func TestChargeDescriptionKeySections(t *testing.T) {
	s := time.Date(2026, 4, 11, 21, 1, 0, 0, time.UTC)
	e := time.Date(2026, 4, 12, 1, 49, 0, 0, time.UTC)
	kwh := 33.9
	p := 11.0
	c := model.Charge{
		StartDate:         &s,
		EndDate:           &e,
		Address:           "蓟门里东区",
		KwhAdded:          &kwh,
		StartBatteryLevel: fptr(11),
		EndBatteryLevel:   fptr(66),
		MaxPower:          &p,
	}
	desc := BuildChargeDescription("Model 3", c, time.UTC)
	if !strings.Contains(desc, "⚡ 充电") || !strings.Contains(desc, "📊 数据") {
		t.Fatalf("charge sections: %s", desc)
	}
	if strings.Contains(desc, "充电ID") {
		t.Fatal("id leak")
	}
}

func TestDailyICSRoundtrip(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	s1 := time.Date(2026, 4, 11, 10, 32, 0, 0, loc)
	e1 := time.Date(2026, 4, 11, 10, 42, 0, 0, loc)
	d0 := 1.85
	cons := 171.0
	sp := 43.0
	sec := 600.0
	rows := BuildDailySummaries(
		[]model.Drive{{
			StartDate: &s1, EndDate: &e1, Distance: &d0, DurationSeconds: &sec,
			StartAddress: "蓟门里东区", EndAddress: "苏宁易购, 海淀区",
			Consumption: &cons, StartBatteryLevel: &sp, EndBatteryLevel: &sp,
		}},
		[]model.Charge{},
		nil, loc,
	)
	if len(rows) != 1 {
		t.Fatalf("rows: %d", len(rows))
	}
	ev := DailySummaryEvents("1", "Model 3", rows, loc, "normal", true, "")
	if len(ev) < 1 {
		t.Fatalf("no events")
	}
	desc := ev[0].Description
	if strings.Contains(desc, "225 min") || strings.Contains(desc, "10min") {
		t.Errorf("forbidden time format: %s", desc)
	}
	if strings.Contains(desc, "33.92kWh") {
		t.Errorf("forbidden kwh format: %s", desc)
	}
	if !strings.Contains(desc, "今日概览") {
		t.Fatalf("overview: %s", desc)
	}
}

func TestDailyDrivesOnlyNoChargeBlock(t *testing.T) {
	day := time.Date(2026, 4, 24, 0, 0, 0, 0, time.Local)
	rows := []model.DailySummary{{
		Day:          day,
		DriveCount:   1,
		Distance:     10,
		DriveSeconds: 300,
	}}
	d := BuildDailyDescription("M3", rows[0], true, time.Local)
	if strings.Contains(d, "⚡ 充电 0") {
		t.Fatalf("no charge: %s", d)
	}
	rows2 := []model.DailySummary{{
		Day:           day,
		ChargeCount:   1,
		KwhAdded:      5,
		ChargeSeconds: 300,
	}}
	d2 := BuildDailyDescription("M3", rows2[0], true, time.Local)
	if strings.Contains(d2, "🚗 行程 0") {
		t.Fatalf("no drives: %s", d2)
	}
}

func fptr(f float64) *float64 {
	v := f
	return &v
}
