package calendar

import (
	"strings"
	"testing"
)

func TestEmptyCalendarIsValid(t *testing.T) {
	ics := BuildCalendar("TeslaMate Calendar", "Asia/Shanghai", nil)
	if !strings.Contains(ics, "BEGIN:VCALENDAR") || !strings.Contains(ics, "END:VCALENDAR") {
		t.Fatal("invalid empty calendar")
	}
	if !strings.Contains(ics, "\r\n") {
		t.Fatal("ics should use CRLF")
	}
}
