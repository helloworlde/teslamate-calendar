package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/helloworlde/teslamate-calendar/internal/model"
	"github.com/helloworlde/teslamate-calendar/internal/util"
)

type Query struct {
	StartDate   time.Time
	EndDate     time.Time
	HasStartEnd bool
	MinDistance string
	MaxDistance string
}

type Client struct {
	BaseURL    *url.URL
	HTTPClient *http.Client
	Token      string
	AuthHeader string
	AuthScheme string
}

func New(baseURL, token, authHeader, authScheme string, timeout time.Duration) (*Client, error) {
	n, err := util.NormalizeTeslaMateAPIBase(baseURL)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(n)
	if err != nil {
		return nil, err
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &Client{
		BaseURL:    u,
		HTTPClient: &http.Client{Timeout: timeout},
		Token:      token,
		AuthHeader: authHeader,
		AuthScheme: authScheme,
	}, nil
}

// apiURL 在 BaseURL（仅 scheme+host）下拼接 TeslaMateApi 的 /api/... 路径。
func (c *Client) apiURL(elem ...string) *url.URL {
	return c.BaseURL.JoinPath(append([]string{"api"}, elem...)...)
}

func (c *Client) Healthz(ctx context.Context) error {
	_, err := c.get(ctx, c.apiURL("healthz"), nil)
	return err
}

func (c *Client) Readyz(ctx context.Context) error {
	_, err := c.get(ctx, c.apiURL("readyz"), nil)
	return err
}

func (c *Client) ListCars(ctx context.Context) ([]model.Car, error) {
	body, err := c.get(ctx, c.apiURL("v1", "cars"), nil)
	if err != nil {
		return nil, err
	}
	var resp apiCarsResponse
	if err := decodeJSON(body, &resp); err != nil {
		return nil, err
	}
	out := make([]model.Car, 0, len(resp.Data.Cars))
	for _, row := range resp.Data.Cars {
		name := ""
		if row.Name != nil {
			name = *row.Name
		}
		out = append(out, model.Car{
			ID:   row.CarID,
			Name: name,
		})
	}
	return out, nil
}

func (c *Client) GetCar(ctx context.Context, carID string) (model.CarProfile, error) {
	wantID, err := strconv.ParseInt(strings.TrimSpace(carID), 10, 64)
	if err != nil || wantID < 1 {
		return model.CarProfile{}, fmt.Errorf("invalid car id")
	}
	body, err := c.get(ctx, c.apiURL("v1", "cars", carID), nil)
	if err != nil {
		return model.CarProfile{}, err
	}
	var resp apiCarsResponse
	if err := decodeJSON(body, &resp); err != nil {
		return model.CarProfile{}, err
	}
	for _, row := range resp.Data.Cars {
		if row.CarID == wantID {
			return teslaMateCarRowToProfile(row), nil
		}
	}
	return model.CarProfile{}, fmt.Errorf("car id %d not found in response", wantID)
}

func (c *Client) ListDrives(ctx context.Context, carID string, q Query) ([]model.Drive, error) {
	v := url.Values{}
	if q.HasStartEnd {
		util.AddQueryTime(v, "startDate", q.StartDate)
		util.AddQueryTime(v, "endDate", q.EndDate)
	}
	if q.MinDistance != "" {
		v.Set("minDistance", q.MinDistance)
	}
	if q.MaxDistance != "" {
		v.Set("maxDistance", q.MaxDistance)
	}
	body, err := c.get(ctx, c.apiURL("v1", "cars", carID, "drives"), v)
	if err != nil {
		return nil, err
	}
	var resp apiDrivesResponse
	if err := decodeJSON(body, &resp); err != nil {
		return nil, err
	}
	out := make([]model.Drive, 0, len(resp.Data.Drives))
	for _, drive := range resp.Data.Drives {
		durationSeconds := drive.DurationSec
		if durationSeconds == nil && drive.DurationMin != nil {
			sec := *drive.DurationMin * 60
			durationSeconds = &sec
		}
		out = append(out, model.Drive{
			ID:                drive.DriveID,
			StartDate:         parseAPITime(drive.StartDate),
			EndDate:           parseAPITime(drive.EndDate),
			Distance:          drive.Distance,
			DurationSeconds:   durationSeconds,
			StartBatteryLevel: drive.StartBatteryLevel,
			EndBatteryLevel:   drive.EndBatteryLevel,
			Consumption:       drive.Consumption,
			AvgSpeed:          drive.AvgSpeed,
			MaxSpeed:          drive.MaxSpeed,
			AvgPower:          drive.AvgPower,
			MaxPower:          drive.MaxPower,
			StartAddress:      drive.StartAddress,
			EndAddress:        drive.EndAddress,
			StartGeofence:     drive.StartGeofence,
			EndGeofence:       drive.EndGeofence,
			StartLat:          drive.StartLat,
			StartLng:          drive.StartLng,
			EndLat:            drive.EndLat,
			EndLng:            drive.EndLng,
			HasRoute:          strings.TrimSpace(drive.Polyline) != "",
		})
	}
	return out, nil
}

func (c *Client) ListCharges(ctx context.Context, carID string, q Query) ([]model.Charge, error) {
	v := url.Values{}
	if q.HasStartEnd {
		util.AddQueryTime(v, "startDate", q.StartDate)
		util.AddQueryTime(v, "endDate", q.EndDate)
	}
	body, err := c.get(ctx, c.apiURL("v1", "cars", carID, "charges"), v)
	if err != nil {
		return nil, err
	}
	var resp apiChargesResponse
	if err := decodeJSON(body, &resp); err != nil {
		return nil, err
	}
	out := make([]model.Charge, 0, len(resp.Data.Charges))
	for _, charge := range resp.Data.Charges {
		out = append(out, model.Charge{
			ID:                charge.ChargeID,
			StartDate:         parseAPITime(charge.StartDate),
			EndDate:           parseAPITime(charge.EndDate),
			DurationMinutes:   charge.DurationMin,
			StartBatteryLevel: charge.StartBatteryLevel,
			EndBatteryLevel:   charge.EndBatteryLevel,
			KwhAdded:          charge.ChargeEnergyAdded,
			Cost:              charge.Cost,
			MaxPower:          charge.MaxPower,
			AvgPower:          charge.AveragePower,
			Address:           charge.Address,
			Geofence:          charge.Geofence,
			Lat:               charge.Latitude,
			Lng:               charge.Longitude,
		})
	}
	return out, nil
}

func (c *Client) ListUpdates(ctx context.Context, carID string) ([]model.Update, error) {
	body, err := c.get(ctx, c.apiURL("v1", "cars", carID, "updates"), nil)
	if err != nil {
		return nil, err
	}
	var resp apiUpdatesResponse
	if err := decodeJSON(body, &resp); err != nil {
		return nil, err
	}
	out := make([]model.Update, 0, len(resp.Data.Updates))
	for _, update := range resp.Data.Updates {
		out = append(out, model.Update{
			ID:           update.UpdateID,
			Version:      update.Version,
			StartDate:    parseAPITime(update.StartDate),
			EndDate:      parseAPITime(update.EndDate),
			Status:       update.Status,
			ReleaseNotes: update.ReleaseNotes,
		})
	}
	return out, nil
}

func (c *Client) GetGlobalSettings(ctx context.Context) (model.GlobalSettings, error) {
	body, err := c.get(ctx, c.apiURL("v1", "globalsettings"), nil)
	if err != nil {
		return model.GlobalSettings{}, err
	}
	var resp apiGlobalSettingsResponse
	if err := decodeJSON(body, &resp); err != nil {
		return model.GlobalSettings{}, err
	}
	return model.GlobalSettings{
		GrafanaURL: firstNonEmpty(resp.Data.Settings.TeslaMateURLs.GrafanaURL, resp.Data.Settings.GrafanaURL),
	}, nil
}

func (c *Client) get(ctx context.Context, target *url.URL, query url.Values) ([]byte, error) {
	body, code, msg, err := c.doGet(ctx, target, query)
	if err != nil {
		return nil, err
	}
	if code >= 200 && code <= 299 {
		return body, nil
	}
	return nil, fmt.Errorf("non-2xx from teslamateapi: %d %s", code, msg)
}

func (c *Client) doGet(ctx context.Context, target *url.URL, query url.Values) ([]byte, int, string, error) {
	u := target
	if len(query) > 0 {
		u2 := *target
		u2.RawQuery = query.Encode()
		u = &u2
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, 0, "", err
	}
	if c.Token != "" && !strings.EqualFold(c.AuthScheme, "None") {
		header := c.AuthHeader
		if header == "" {
			header = "Authorization"
		}
		scheme := c.AuthScheme
		if scheme == "" {
			scheme = "Bearer"
		}
		req.Header.Set(header, strings.TrimSpace(scheme+" "+c.Token))
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, 0, "", fmt.Errorf("request timeout: %w", err)
		}
		return nil, 0, "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, "", err
	}
	return body, resp.StatusCode, strings.TrimSpace(string(body)), nil
}

func decodeJSON(body []byte, out any) error {
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("failed to decode teslamateapi response: %w", err)
	}
	return nil
}

func parseAPITime(raw string) *time.Time {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return &t
	}
	if t, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return &t
	}
	return nil
}

func teslaMateCarRowToProfile(row apiTeslaMateCarRow) model.CarProfile {
	name := ""
	if row.Name != nil {
		name = *row.Name
	}
	return model.CarProfile{
		ID:          row.CarID,
		Name:        name,
		DisplayName: "",
		VIN:         row.CarDetails.Vin,
		Model:       row.CarDetails.Model,
	}
}

func firstNonEmpty(v ...string) string {
	for _, s := range v {
		if strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	return ""
}
