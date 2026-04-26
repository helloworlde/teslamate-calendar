package util

import (
	"net/url"
	"testing"
	"time"
)

func TestUTCConversionRFC3339(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	tt, err := ParseFlexibleTime("2026-04-01 08:00:00", loc)
	if err != nil {
		t.Fatal(err)
	}
	v := url.Values{}
	AddQueryTime(v, "startDate", tt)
	if got := v.Get("startDate"); got != "2026-04-01T00:00:00Z" {
		t.Fatalf("unexpected utc RFC3339: %s", got)
	}
}
