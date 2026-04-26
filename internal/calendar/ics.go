package calendar

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type Event struct {
	UID         string
	DTStamp     time.Time
	Start       time.Time
	End         time.Time
	AllDay      bool
	AllDayDate  time.Time
	AllDayEnd   time.Time
	Summary     string
	Description string
	Location    string
	Geo         string
	URL         string
}

func StableUID(parts ...string) string {
	h := sha1.New()
	for _, p := range parts {
		_, _ = h.Write([]byte(p))
		_, _ = h.Write([]byte("|"))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func BuildCalendar(name, timezone string, events []Event) string {
	lines := []string{
		"BEGIN:VCALENDAR",
		"VERSION:2.0",
		"PRODID:-//TeslaMate Calendar//teslamate-calendar//CN",
		"CALSCALE:GREGORIAN",
		"METHOD:PUBLISH",
		"X-WR-CALNAME:" + EscapeText(name),
		"X-WR-TIMEZONE:" + EscapeText(timezone),
	}
	for _, e := range events {
		uid := e.UID
		if uid == "" {
			uid = StableUID(e.Summary, e.Start.UTC().Format(time.RFC3339), e.End.UTC().Format(time.RFC3339))
		}
		stamp := e.DTStamp
		if stamp.IsZero() {
			stamp = time.Now().UTC()
		}
		lines = append(lines, "BEGIN:VEVENT")
		lines = append(lines, "UID:"+EscapeText(uid))
		lines = append(lines, "DTSTAMP:"+stamp.UTC().Format("20060102T150405Z"))
		if e.AllDay {
			start := e.AllDayDate.In(time.UTC).Format("20060102")
			end := e.AllDayEnd.In(time.UTC).Format("20060102")
			if end == "" || end <= start {
				end = e.AllDayDate.AddDate(0, 0, 1).In(time.UTC).Format("20060102")
			}
			lines = append(lines, "DTSTART;VALUE=DATE:"+start)
			lines = append(lines, "DTEND;VALUE=DATE:"+end)
		} else {
			lines = append(lines, "DTSTART:"+e.Start.UTC().Format("20060102T150405Z"))
			end := e.End
			if end.IsZero() || !end.After(e.Start) {
				end = e.Start.Add(time.Minute)
			}
			lines = append(lines, "DTEND:"+end.UTC().Format("20060102T150405Z"))
		}
		lines = append(lines, "SUMMARY:"+EscapeText(e.Summary))
		lines = append(lines, "DESCRIPTION:"+EscapeText(e.Description))
		if strings.TrimSpace(e.Location) != "" {
			lines = append(lines, "LOCATION:"+EscapeText(e.Location))
		}
		if strings.TrimSpace(e.Geo) != "" {
			lines = append(lines, "GEO:"+EscapeText(e.Geo))
		}
		if strings.TrimSpace(e.URL) != "" {
			lines = append(lines, "URL:"+EscapeText(e.URL))
		}
		lines = append(lines, "END:VEVENT")
	}
	lines = append(lines, "END:VCALENDAR", "")
	return strings.Join(lines, "\r\n")
}

func UIDWithFallback(prefix string, id any, fallbackParts ...string) string {
	if id != nil {
		s := strings.TrimSpace(fmt.Sprintf("%v", id))
		if s != "" && s != "<nil>" {
			return prefix + s
		}
	}
	return prefix + StableUID(fallbackParts...)
}
