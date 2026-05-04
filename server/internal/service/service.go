package service

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/koldun11/yartime/server/config"
	"github.com/koldun11/yartime/server/internal/models"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

// Servicer интерфейс для слоя сервиса
type Servicer interface {
	GetClientConfig() (*models.ClientConfigResponse, error)
	SetAllowedHours(start, end string) error
	SetCronLine(cronLine string) error
	SetDailyLimit(limit int) error
	SetControlledApps(apps []string) error
	GetVersion() (*models.VersionInfo, error)
	GetBinaryPath(osName, arch string) (string, error)
}

// Service реализует интерфейс handler.Service
type Service struct {
	config *config.AppConfig
	logger *zap.Logger
	mutex  sync.RWMutex
}

// NewService создаёт новый Service
func NewService(config *config.AppConfig, logger *zap.Logger) *Service {
	return &Service{
		config: config,
		logger: logger,
	}
}

// GetClientConfig возвращает конфигурацию клиента
func (s *Service) GetClientConfig() (*models.ClientConfigResponse, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	apps := s.config.Client.ControlledApps
	if apps == nil {
		apps = []string{}
	}

	return &models.ClientConfigResponse{
		ClientID:          s.config.Client.ClientID,
		DailyLimitMinutes: s.config.Client.DailyLimitMinutes,
		AllowedHoursStart: s.config.Client.AllowedHoursStart,
		AllowedHoursEnd:   s.config.Client.AllowedHoursEnd,
		ExecuteOnStart:    s.config.Client.ExecuteOnStart,
		CronLine:          s.config.Client.CronLine,
		ControlledApps:    apps,
	}, nil
}

// SetAllowedHours устанавливает разрешённое время
func (s *Service) SetAllowedHours(start, end string) error {
	if !isValidTime(start) || !isValidTime(end) {
		return fmt.Errorf("invalid time format: start=%s, end=%s", start, end)
	}

	startTime, _ := time.Parse("15:04", start)
	endTime, _ := time.Parse("15:04", end)
	if startTime.Equal(endTime) {
		return fmt.Errorf("start time (%s) cannot equal end time (%s)", start, end)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.config.Client.AllowedHoursStart = start
	s.config.Client.AllowedHoursEnd = end

	if err := s.saveConfig(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	s.logger.Info("Allowed hours updated",
		zap.String("start", start),
		zap.String("end", end))
	return nil
}

// SetCronLine устанавливает строку cron
func (s *Service) SetCronLine(cronLine string) error {
	if cronLine == "" {
		return fmt.Errorf("cron line cannot be empty")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.config.Client.CronLine = cronLine

	if err := s.saveConfig(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	s.logger.Info("Cron line updated", zap.String("cron_line", cronLine))
	return nil
}

// SetDailyLimit устанавливает дневной лимит
func (s *Service) SetDailyLimit(limit int) error {
	if limit <= 0 {
		return fmt.Errorf("daily limit must be positive, got %d", limit)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.config.Client.DailyLimitMinutes = limit

	if err := s.saveConfig(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	s.logger.Info("Daily limit updated", zap.Int("daily_limit_minutes", limit))
	return nil
}

// SetControlledApps устанавливает список контролируемых приложений
func (s *Service) SetControlledApps(apps []string) error {
	if apps == nil {
		apps = []string{}
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.config.Client.ControlledApps = apps

	if err := s.saveConfig(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	s.logger.Info("Controlled apps updated", zap.Strings("apps", apps))
	return nil
}

// GetVersion возвращает информацию о версии клиента
func (s *Service) GetVersion() (*models.VersionInfo, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	binaryDir := s.config.Server.BinaryDir
	if binaryDir == "" {
		return &models.VersionInfo{Version: "0.0.0", Checksum: ""}, nil
	}

	versionPath := filepath.Join(binaryDir, "version.json")
	data, err := os.ReadFile(versionPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &models.VersionInfo{Version: "0.0.0", Checksum: ""}, nil
		}
		return nil, fmt.Errorf("failed to read version.json: %w", err)
	}

	var info models.VersionInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("failed to parse version.json: %w", err)
	}

	return &info, nil
}

// GetBinaryPath возвращает путь к бинарнику клиента для указанной ОС и архитектуры
func (s *Service) GetBinaryPath(osName, arch string) (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	binaryDir := s.config.Server.BinaryDir
	if binaryDir == "" {
		return "", fmt.Errorf("binary_dir not configured")
	}

	filename := fmt.Sprintf("yartime-client-%s-%s", osName, arch)
	path := filepath.Join(binaryDir, filename)

	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("binary not found: %s", filename)
	}

	return path, nil
}

// ComputeBinaryChecksum вычисляет SHA-256 checksum файла
func ComputeBinaryChecksum(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash), nil
}

// saveConfig сохраняет конфигурацию в config.json
func (s *Service) saveConfig() error {
	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.config.ConfigPath, data, 0644)
}

// isValidTime проверяет формат времени HH:MM
func isValidTime(t string) bool {
	matched, _ := regexp.MatchString(`^([0-1][0-9]|2[0-3]):[0-5][0-9]$`, t)
	return matched
}
