package util

import "testing"

func TestNormalizeTeslaMateAPIBase(t *testing.T) {
	out, err := NormalizeTeslaMateAPIBase("http://teslamateapi:8080/api")
	if err != nil {
		t.Fatal(err)
	}
	if out != "http://teslamateapi:8080" {
		t.Fatalf("got %q", out)
	}
	out, err = NormalizeTeslaMateAPIBase("http://host/path/sub")
	if err != nil {
		t.Fatal(err)
	}
	if out != "http://host" {
		t.Fatalf("got %q", out)
	}
}

func TestNormalizeTeslaMateAPIBaseRejects(t *testing.T) {
	_, err := NormalizeTeslaMateAPIBase("nohost")
	if err == nil {
		t.Fatal("expected error")
	}
}
