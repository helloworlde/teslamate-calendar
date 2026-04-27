package calendar

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// Event 表示一个日历事件
type Event struct {
	UID             string
	DTStamp         time.Time
	Start           time.Time
	End             time.Time
	AllDay          bool
	AllDayDate      time.Time
	AllDayEnd       time.Time
	Summary         string
	Description     string
	HTMLDescription string
	Location        string
	Geo             string
	URL             string
}

// StableUID 生成稳定的唯一标识符，用于日历事件的 UID
func StableUID(parts ...string) string {
	h := sha1.New()
	for _, p := range parts {
		_, _ = h.Write([]byte(p))
		_, _ = h.Write([]byte("|"))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// BuildCalendar 构建符合 RFC 5545 标准的 iCalendar 格式字符串
func BuildCalendar(name, timezone string, events []Event) string {
	loc := time.Local
	if strings.TrimSpace(timezone) != "" {
		if l, err := time.LoadLocation(timezone); err == nil {
			loc = l
		}
	}
	lines := []string{
		"BEGIN:VCALENDAR",
		"VERSION:2.0",
		"PRODID:-//TeslaMate Calendar//github.com/helloworlde/teslamate-calendar//CN",
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
			ds := e.AllDayDate.In(loc)
			de := e.AllDayEnd.In(loc)
			start := time.Date(ds.Year(), ds.Month(), ds.Day(), 0, 0, 0, 0, loc)
			endExcl := time.Date(de.Year(), de.Month(), de.Day(), 0, 0, 0, 0, loc)
			startS := start.Format("20060102")
			endS := endExcl.Format("20060102")
			if endS <= startS {
				endS = start.AddDate(0, 0, 1).Format("20060102")
			}
			lines = append(lines, "DTSTART;VALUE=DATE:"+startS)
			lines = append(lines, "DTEND;VALUE=DATE:"+endS)
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
		if strings.TrimSpace(e.HTMLDescription) != "" {
			lines = append(lines, "X-ALT-DESC;FMTTYPE=text/html:"+EscapeText(e.HTMLDescription))
		}
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
