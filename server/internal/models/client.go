package models

// ClientConfigResponse представляет конфигурацию клиента для API
type ClientConfigResponse struct {
	ClientID          string   `json:"client_id"`
	DailyLimitMinutes int      `json:"daily_limit_minutes"`
	AllowedHoursStart string   `json:"allowed_hours_start"`
	AllowedHoursEnd   string   `json:"allowed_hours_end"`
	ExecuteOnStart    string   `json:"execute_on_start"`
	CronLine          string   `json:"cron_line"`
	ControlledApps    []string `json:"controlled_apps"`
}

// VersionInfo представляет информацию о версии клиента для self-update
type VersionInfo struct {
	Version  string `json:"version"`
	Checksum string `json:"checksum"`
}
