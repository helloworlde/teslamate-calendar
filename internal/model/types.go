package model

import "time"

type Car struct {
	ID   int64          `json:"id"`
	Name string         `json:"name"`
	Raw  map[string]any `json:"-"`
}

type Drive struct {
	ID                any
	StartDate         *time.Time
	EndDate           *time.Time
	Distance          *float64
	DurationSeconds   *float64
	StartBatteryLevel *float64
	EndBatteryLevel   *float64
	Consumption       *float64
	AvgSpeed          *float64
	MaxSpeed          *float64
	AvgPower          *float64
	MaxPower          *float64
	StartAddress      string
	EndAddress        string
	StartGeofence     string
	EndGeofence       string
	StartLat          *float64
	StartLng          *float64
	EndLat            *float64
	EndLng            *float64
	Raw               map[string]any
}

type Charge struct {
	ID                any
	StartDate         *time.Time
	EndDate           *time.Time
	DurationMinutes   *float64
	StartBatteryLevel *float64
	EndBatteryLevel   *float64
	KwhAdded          *float64
	Cost              *float64
	MaxPower          *float64
	AvgPower          *float64
	Address           string
	Geofence          string
	Lat               *float64
	Lng               *float64
	Raw               map[string]any
}

type Update struct {
	ID           any
	Version      string
	StartDate    *time.Time
	EndDate      *time.Time
	Status       string
	ReleaseNotes string
	Raw          map[string]any
}

type DailySummary struct {
	Day            time.Time
	DriveCount     int
	ChargeCount    int
	UpdateCount    int
	Distance       float64
	DriveSeconds   float64
	ChargeSeconds  float64
	KwhAdded       float64
	Cost           float64
	MaxSpeed       float64
	MaxChargePower float64
	Consumption    float64
	StartBattery   *float64
	EndBattery     *float64
	UpdateVersions []string
	DriveDetails   []string
	ChargeDetails  []string
}

type GlobalSettings struct {
	Raw map[string]any
}
