package calendar

import (
	"strings"
	"testing"
	"time"
)

func TestAllDayAsiaShanghaiNoPreviousDay(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	day := time.Date(2026, 4, 11, 0, 0, 0, 0, loc)
	ics := BuildCalendar("n", "Asia/Shanghai", []Event{{
		AllDay:      true,
		AllDayDate:  day,
		AllDayEnd:   day.AddDate(0, 0, 1),
		Summary:     "S",
		Description: "D",
	}})
	if !strings.Contains(ics, "DTSTART;VALUE=DATE:20260411") {
		t.Fatalf("got %q", ics)
	}
	if !strings.Contains(ics, "DTEND;VALUE=DATE:20260412") {
		t.Fatalf("got %q", ics)
	}
	if strings.Contains(ics, "X-ALT-DESC") {
		t.Fatal("no default html alt")
	}
}

func TestXAltDescWhenSet(t *testing.T) {
	day := time.Date(2026, 4, 11, 0, 0, 0, 0, time.UTC)
	ics := BuildCalendar("n", "UTC", []Event{{
		AllDay:          true,
		AllDayDate:      day,
		AllDayEnd:       day.AddDate(0, 0, 1),
		Summary:         "S",
		Description:     "plain",
		HTMLDescription: "<p>hi</p>",
	}})
	if !strings.Contains(ics, "X-ALT-DESC;FMTTYPE=text/html:") {
		t.Fatalf("expected alt desc, got: %q", ics)
	}
}

func TestAllDayUTCNudge(t *testing.T) {
	day := time.Date(2026, 4, 11, 0, 0, 0, 0, time.UTC)
	ics := BuildCalendar("n", "UTC", []Event{{
		AllDay:      true,
		AllDayDate:  day,
		AllDayEnd:   day.AddDate(0, 0, 1),
		Summary:     "S",
		Description: "D",
	}})
	if !strings.Contains(ics, "DTSTART;VALUE=DATE:20260411") || !strings.Contains(ics, "DTEND;VALUE=DATE:20260412") {
		t.Fatalf("%q", ics)
	}
}

func TestAllDayAmericaLosAngeles(t *testing.T) {
	loc, _ := time.LoadLocation("America/Los_Angeles")
	day := time.Date(2026, 4, 11, 0, 0, 0, 0, loc)
	ics := BuildCalendar("n", "America/Los_Angeles", []Event{{
		AllDay:      true,
		AllDayDate:  day,
		AllDayEnd:   day.AddDate(0, 0, 1),
		Summary:     "S",
		Description: "D",
	}})
	if !strings.Contains(ics, "DTSTART;VALUE=DATE:20260411") {
		t.Fatalf("%q", ics)
	}
}
