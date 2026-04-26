package client

// apiCarsResponse 对应 definitions 中 main.SwaggerDataResponse：
// { "data": { "cars": [ ... ] } }，与 TeslaMateApi 兼容路由 GET /v1/cars、GET /v1/cars/{CarID} 实际返回一致（单车查询仍在 cars 数组中返回一条）。
type apiCarsResponse struct {
	Data struct {
		Cars []apiTeslaMateCarRow `json:"cars"`
	} `json:"data"`
}

type apiTeslaMateCarRow struct {
	CarID            int64                   `json:"car_id"`
	Name             *string                 `json:"name"`
	CarDetails       apiTeslaMateCarDetails  `json:"car_details"`
	CarExterior      apiTeslaMateCarExterior `json:"car_exterior"`
	CarSettings      apiTeslaMateCarSettings `json:"car_settings"`
	TeslaMateDetails apiTeslaMateTimestamps  `json:"teslamate_details"`
	TeslaMateStats   apiTeslaMateCarStats    `json:"teslamate_stats"`
}

type apiTeslaMateCarDetails struct {
	EID         int64   `json:"eid"`
	VID         int64   `json:"vid"`
	Vin         string  `json:"vin"`
	Model       string  `json:"model"`
	TrimBadging string  `json:"trim_badging"`
	Efficiency  float64 `json:"efficiency"`
}

type apiTeslaMateCarExterior struct {
	ExteriorColor string `json:"exterior_color"`
	SpoilerType   string `json:"spoiler_type"`
	WheelType     string `json:"wheel_type"`
}

type apiTeslaMateCarSettings struct {
	SuspendMin          int  `json:"suspend_min"`
	SuspendAfterIdleMin int  `json:"suspend_after_idle_min"`
	ReqNotUnlocked      bool `json:"req_not_unlocked"`
	FreeSupercharging   bool `json:"free_supercharging"`
	UseStreamingAPI     bool `json:"use_streaming_api"`
}

type apiTeslaMateTimestamps struct {
	InsertedAt string `json:"inserted_at"`
	UpdatedAt  string `json:"updated_at"`
}

type apiTeslaMateCarStats struct {
	TotalCharges int `json:"total_charges"`
	TotalDrives  int `json:"total_drives"`
	TotalUpdates int `json:"total_updates"`
}

type apiDrivesResponse struct {
	Data struct {
		Drives []apiDrive `json:"drives"`
	} `json:"data"`
}

type apiDrive struct {
	DriveID           any      `json:"drive_id"`
	StartDate         string   `json:"start_date"`
	EndDate           string   `json:"end_date"`
	Distance          *float64 `json:"distance"`
	DurationSec       *float64 `json:"duration_sec"`
	DurationMin       *float64 `json:"duration_min"`
	StartBatteryLevel *float64 `json:"start_battery_level"`
	EndBatteryLevel   *float64 `json:"end_battery_level"`
	Consumption       *float64 `json:"consumption"`
	AvgSpeed          *float64 `json:"average_speed"`
	MaxSpeed          *float64 `json:"max_speed"`
	AvgPower          *float64 `json:"average_power"`
	MaxPower          *float64 `json:"max_power"`
	StartAddress      string   `json:"start_address"`
	EndAddress        string   `json:"end_address"`
	StartGeofence     string   `json:"start_geofence"`
	EndGeofence       string   `json:"end_geofence"`
	StartLat          *float64 `json:"start_lat"`
	StartLng          *float64 `json:"start_lng"`
	EndLat            *float64 `json:"end_lat"`
	EndLng            *float64 `json:"end_lng"`
	Polyline          string   `json:"polyline"`
}

type apiChargesResponse struct {
	Data struct {
		Charges []apiCharge `json:"charges"`
	} `json:"data"`
}

type apiCharge struct {
	ChargeID           any      `json:"charge_id"`
	StartDate          string   `json:"start_date"`
	EndDate            string   `json:"end_date"`
	DurationMin        *float64 `json:"duration_min"`
	StartBatteryLevel  *float64 `json:"start_battery_level"`
	EndBatteryLevel    *float64 `json:"end_battery_level"`
	ChargeEnergyAdded  *float64 `json:"charge_energy_added"`
	Cost               *float64 `json:"cost"`
	MaxPower           *float64 `json:"max_power"`
	AveragePower       *float64 `json:"average_power"`
	Address            string   `json:"address"`
	Geofence           string   `json:"geofence"`
	Latitude           *float64 `json:"latitude"`
	Longitude          *float64 `json:"longitude"`
}

type apiUpdatesResponse struct {
	Data struct {
		Updates []apiUpdate `json:"updates"`
	} `json:"data"`
}

type apiUpdate struct {
	UpdateID      any    `json:"update_id"`
	Version       string `json:"version"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	Status        string `json:"status"`
	ReleaseNotes  string `json:"release_notes"`
}

type apiGlobalSettingsResponse struct {
	Data struct {
		Settings apiGlobalSettings `json:"settings"`
	} `json:"data"`
}

type apiGlobalSettings struct {
	TeslaMateURLs apiTeslaMateURLs `json:"teslamate_urls"`
	GrafanaURL    string           `json:"grafana_url"`
}

type apiTeslaMateURLs struct {
	GrafanaURL string `json:"grafana_url"`
}
