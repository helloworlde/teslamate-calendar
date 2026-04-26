package calendar

import (
	"fmt"
	"math"
	"strings"
)

func rawValue(m map[string]any, keys ...string) any {
	cur := any(m)
	for _, key := range keys {
		mv, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		v, ok := mv[key]
		if !ok {
			return nil
		}
		cur = v
	}
	return cur
}

func rawFloat(m map[string]any, keys ...string) *float64 {
	v := rawValue(m, keys...)
	switch x := v.(type) {
	case float64:
		return &x
	case float32:
		y := float64(x)
		return &y
	case int:
		y := float64(x)
		return &y
	case int64:
		y := float64(x)
		return &y
	}
	return nil
}

func rawString(m map[string]any, keys ...string) string {
	v := rawValue(m, keys...)
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(s)
}

func valueOr(v *float64, fallback *float64) *float64 {
	if v != nil {
		return v
	}
	return fallback
}

func percentDelta(start, end *float64) string {
	if start == nil || end == nil {
		return ""
	}
	d := *end - *start
	return fmt.Sprintf("%.0f%%→%.0f%% (%+.0f%%)", *start, *end, d)
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}
