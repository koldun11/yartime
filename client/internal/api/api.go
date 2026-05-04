package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/koldun11/yartime/client/internal/store"
)

// VersionInfo holds version and checksum for self-update.
type VersionInfo struct {
	Version  string `json:"version"`
	Checksum string `json:"checksum"`
}

// Client communicates with the yartime server API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New creates a new API client with the given base URL.
func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// FetchConfig retrieves the client configuration from the server.
func (c *Client) FetchConfig(clientID string) (store.Config, error) {
	url := fmt.Sprintf("%s/client/config", c.baseURL)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return store.Config{}, fmt.Errorf("failed to fetch config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return store.Config{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var raw struct {
		ClientID          string   `json:"client_id"`
		DailyLimitMinutes int      `json:"daily_limit_minutes"`
		AllowedHoursStart string   `json:"allowed_hours_start"`
		AllowedHoursEnd   string   `json:"allowed_hours_end"`
		ExecuteOnStart    string   `json:"execute_on_start"`
		CronLine          string   `json:"cron_line"`
		ControlledApps    []string `json:"controlled_apps"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return store.Config{}, fmt.Errorf("failed to decode config: %w", err)
	}

	apps := raw.ControlledApps
	if apps == nil {
		apps = []string{}
	}

	return store.Config{
		ClientID:          raw.ClientID,
		AllowedHoursStart: raw.AllowedHoursStart,
		AllowedHoursEnd:   raw.AllowedHoursEnd,
		ControlledApps:    apps,
		ExecuteOnStart:    raw.ExecuteOnStart,
	}, nil
}

// FetchVersion retrieves the current client version from the server.
func (c *Client) FetchVersion() (VersionInfo, error) {
	url := fmt.Sprintf("%s/client/version", c.baseURL)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return VersionInfo{}, fmt.Errorf("failed to fetch version: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return VersionInfo{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var info VersionInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return VersionInfo{}, fmt.Errorf("failed to decode version: %w", err)
	}

	return info, nil
}

// DownloadBinary downloads the client binary for the given OS and architecture.
func (c *Client) DownloadBinary(osName, arch string) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/client/binary?os=%s&arch=%s", c.baseURL, osName, arch)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download binary: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}
