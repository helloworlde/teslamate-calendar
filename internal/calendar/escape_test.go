package calendar

import "testing"

func TestEscapeText(t *testing.T) {
	in := "a\\b;c,d\ne"
	out := EscapeText(in)
	if out != "a\\\\b\\;c\\,d\\ne" {
		t.Fatalf("unexpected escaped text: %s", out)
	}
}
