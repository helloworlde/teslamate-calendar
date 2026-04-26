package calendar

import (
	"testing"
	"time"

	"teslamate-calendar/internal/model"
)

func TestDailySummaryAggregation(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	d1s := time.Date(2026, 4, 1, 10, 0, 0, 0, loc)
	d1e := d1s.Add(30 * time.Minute)
	dist := 12.3
	kwh := 22.2
	c1s := time.Date(2026, 4, 1, 21, 0, 0, 0, loc)
	c1e := c1s.Add(20 * time.Minute)

	rows := BuildDailySummaries(
		[]model.Drive{{StartDate: &d1s, EndDate: &d1e, Distance: &dist}},
		[]model.Charge{{StartDate: &c1s, EndDate: &c1e, KwhAdded: &kwh}},
		nil,
		loc,
	)
	if len(rows) != 1 {
		t.Fatalf("unexpected rows: %d", len(rows))
	}
	if rows[0].DriveCount != 1 || rows[0].ChargeCount != 1 {
		t.Fatal("aggregation failed")
	}
}
