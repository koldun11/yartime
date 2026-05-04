package config

import (
	"encoding/json"
	"os"
)

// ServConfig конфигурация сервера
type ServConfig struct {
	Port      string `json:"port"`
	BinaryDir string `json:"binary_dir"`
}

// ClientConfig конфигурация клиента
type ClientConfig struct {
	ClientID          string   `json:"client_id"`
	DailyLimitMinutes int      `json:"daily_limit_minutes"`
	AllowedHoursStart string   `json:"allowed_hours_start"`
	AllowedHoursEnd   string   `json:"allowed_hours_end"`
	ExecuteOnStart    string   `json:"execute_on_start"`
	CronLine          string   `json:"cron_line"`
	ControlledApps    []string `json:"controlled_apps"`
}

// AppConfig общая структура конфигурации
type AppConfig struct {
	ConfigPath string       `json:"-"`
	Server     ServConfig   `json:"server"`
	Client     ClientConfig `json:"client"`
}

// NewAppConfig загружает конфигурацию из JSON
func NewAppConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config AppConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	config.ConfigPath = path // Сохраняем путь к файлу
	// TODO: Валидация конфигурации, дефолтные значения
	return &config, nil
}
