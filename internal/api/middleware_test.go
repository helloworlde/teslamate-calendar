package api

import "testing"

func TestConstantTimeEquals(t *testing.T) {
	if !ConstantTimeEquals("abc", "abc") {
		t.Fatal("expected equal")
	}
	if ConstantTimeEquals("abc", "abd") {
		t.Fatal("expected not equal")
	}
}
