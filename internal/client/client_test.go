package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientWithMockServer(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/readyz":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok":true}`))
		case "/api/v1/cars":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{"cars":[{"car_id":1,"name":"M3","car_details":{"vin":"","model":null}}]}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer s.Close()

	c, err := New(s.URL, "", "Authorization", "Bearer", 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Readyz(context.Background()); err != nil {
		t.Fatal(err)
	}
	cars, err := c.ListCars(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(cars) != 1 || cars[0].ID != 1 {
		t.Fatalf("unexpected cars: %+v", cars)
	}
}
