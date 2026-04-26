package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"teslamate-calendar/internal/model"
	"teslamate-calendar/internal/util"
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

func (c *Client) Healthz(ctx context.Context) error {
	_, err := c.get(ctx, []string{"healthz"}, nil)
	return err
}

func (c *Client) Readyz(ctx context.Context) error {
	_, err := c.get(ctx, []string{"readyz"}, nil)
	return err
}

func (c *Client) ListCars(ctx context.Context) ([]model.Car, error) {
	body, err := c.get(ctx, []string{"v1", "cars"}, nil)
	if err != nil {
		return nil, err
	}
	rows, err := decodeArrayPayload(body, "cars")
	if err != nil {
		return nil, err
	}
	return model.ParseCars(rows), nil
}

func (c *Client) GetCar(ctx context.Context, carID string) (map[string]any, error) {
	body, err := c.get(ctx, []string{"v1", "cars", carID}, nil)
	if err != nil {
		return nil, err
	}
	return decodeObjectPayload(body, "car")
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
	body, err := c.get(ctx, []string{"v1", "cars", carID, "drives"}, v)
	if err != nil {
		return nil, err
	}
	rows, err := decodeArrayPayload(body, "drives")
	if err != nil {
		return nil, err
	}
	return model.ParseDrives(rows), nil
}

func (c *Client) ListCharges(ctx context.Context, carID string, q Query) ([]model.Charge, error) {
	v := url.Values{}
	if q.HasStartEnd {
		util.AddQueryTime(v, "startDate", q.StartDate)
		util.AddQueryTime(v, "endDate", q.EndDate)
	}
	body, err := c.get(ctx, []string{"v1", "cars", carID, "charges"}, v)
	if err != nil {
		return nil, err
	}
	rows, err := decodeArrayPayload(body, "charges")
	if err != nil {
		return nil, err
	}
	return model.ParseCharges(rows), nil
}

func (c *Client) ListUpdates(ctx context.Context, carID string) ([]model.Update, error) {
	body, err := c.get(ctx, []string{"v1", "cars", carID, "updates"}, nil)
	if err != nil {
		return nil, err
	}
	rows, err := decodeArrayPayload(body, "updates")
	if err != nil {
		return nil, err
	}
	return model.ParseUpdates(rows), nil
}

func (c *Client) GetGlobalSettings(ctx context.Context) (model.GlobalSettings, error) {
	body, err := c.get(ctx, []string{"v1", "globalsettings"}, nil)
	if err != nil {
		return model.GlobalSettings{}, err
	}
	obj, err := decodeObjectPayload(body, "globalsettings", "settings")
	if err != nil {
		return model.GlobalSettings{}, err
	}
	return model.GlobalSettings{Raw: obj}, nil
}

func (c *Client) get(ctx context.Context, segments []string, query url.Values) ([]byte, error) {
	body, code, msg, err := c.doGet(ctx, segments, query)
	if err != nil {
		return nil, err
	}
	if code >= 200 && code <= 299 {
		return body, nil
	}
	return nil, fmt.Errorf("non-2xx from teslamateapi: %d %s", code, msg)
}

func (c *Client) doGet(ctx context.Context, segments []string, query url.Values) ([]byte, int, string, error) {
	u := *c.BaseURL
	joined := make([]string, 0, len(segments)+1)
	joined = append(joined, "api")
	joined = append(joined, segments...)
	u.Path = path.Join(joined...)
	if !strings.HasPrefix(u.Path, "/") {
		u.Path = "/" + u.Path
	}
	if len(query) > 0 {
		u.RawQuery = query.Encode()
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

func decodeArrayPayload(body []byte, preferredKeys ...string) ([]map[string]any, error) {
	var arr []map[string]any
	if err := json.Unmarshal(body, &arr); err == nil {
		return arr, nil
	}
	var wrapped struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(body, &wrapped); err == nil && wrapped.Data != nil {
		return wrapped.Data, nil
	}
	var wrappedRows struct {
		Rows []map[string]any `json:"rows"`
	}
	if err := json.Unmarshal(body, &wrappedRows); err == nil && wrappedRows.Rows != nil {
		return wrappedRows.Rows, nil
	}
	var root map[string]any
	if err := json.Unmarshal(body, &root); err == nil {
		for _, key := range preferredKeys {
			if rows, ok := arrayAt(root, "data", key); ok {
				return rows, nil
			}
			if rows, ok := arrayAt(root, key); ok {
				return rows, nil
			}
		}
		if rows, ok := firstArray(root); ok {
			return rows, nil
		}
	}
	return nil, errors.New("failed to decode array payload")
}

func decodeObjectPayload(body []byte, preferredKeys ...string) (map[string]any, error) {
	var obj map[string]any
	if err := json.Unmarshal(body, &obj); err == nil {
		for _, key := range preferredKeys {
			if nested, ok := objectAt(obj, "data", key); ok {
				return nested, nil
			}
			if nested, ok := objectAt(obj, key); ok {
				return nested, nil
			}
		}
		return obj, nil
	}
	var wrapped struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(body, &wrapped); err == nil && wrapped.Data != nil {
		return wrapped.Data, nil
	}
	return nil, errors.New("failed to decode object payload")
}

func arrayAt(root map[string]any, keys ...string) ([]map[string]any, bool) {
	cur := any(root)
	for _, key := range keys {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = m[key]
		if !ok {
			return nil, false
		}
	}
	rows, ok := cur.([]any)
	if !ok {
		return nil, false
	}
	out := make([]map[string]any, 0, len(rows))
	for _, item := range rows {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		out = append(out, m)
	}
	return out, true
}

func objectAt(root map[string]any, keys ...string) (map[string]any, bool) {
	cur := any(root)
	for _, key := range keys {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = m[key]
		if !ok {
			return nil, false
		}
	}
	m, ok := cur.(map[string]any)
	return m, ok
}

func firstArray(root map[string]any) ([]map[string]any, bool) {
	for _, v := range root {
		if rows, ok := v.([]any); ok {
			out := make([]map[string]any, 0, len(rows))
			for _, item := range rows {
				m, ok := item.(map[string]any)
				if ok {
					out = append(out, m)
				}
			}
			return out, true
		}
		if m, ok := v.(map[string]any); ok {
			if rows, ok := firstArray(m); ok {
				return rows, true
			}
		}
	}
	return nil, false
}
