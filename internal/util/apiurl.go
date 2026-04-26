package util

import (
	"fmt"
	"net/url"
	"strings"
)

func NormalizeTeslaMateAPIBase(s string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(s))
	if err != nil {
		return "", err
	}
	if u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("TESLAMATE_API_BASE_URL 须包含协议与主机，例如 http://teslamateapi:8080")
	}
	u.Path, u.RawPath, u.RawQuery, u.Fragment = "", "", "", ""
	return strings.TrimRight(u.String(), "/"), nil
}
