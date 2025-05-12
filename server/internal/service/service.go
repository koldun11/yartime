package service

import (
	"encoding/json"
	"fmt"
	"github.com/koldun11/yartime/server/config"
	"github.com/koldun11/yartime/server/internal/models"
	"go.uber.org/zap"
	"os"
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

	return &models.ClientConfigResponse{
		ClientID:          s.config.Client.ClientID,
		DailyLimitMinutes: s.config.Client.DailyLimitMinutes,
		AllowedHoursStart: s.config.Client.AllowedHoursStart,
		AllowedHoursEnd:   s.config.Client.AllowedHoursEnd,
		ExecuteOnStart:    s.config.Client.ExecuteOnStart,
		CronLine:          s.config.Client.CronLine,
	}, nil
}

// SetAllowedHours устанавливает разрешённое время
func (s *Service) SetAllowedHours(start, end string) error {
	// Валидация времени (формат HH:MM)
	if !isValidTime(start) || !isValidTime(end) {
		return fmt.Errorf("invalid time format: start=%s, end=%s", start, end)
	}

	// Проверка интервала
	startTime, _ := time.Parse("15:04", start)
	endTime, _ := time.Parse("15:04", end)
	if startTime.Equal(endTime) {
		return fmt.Errorf("start time (%s) cannot equal end time (%s)", start, end)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Обновление конфигурации
	s.config.Client.AllowedHoursStart = start
	s.config.Client.AllowedHoursEnd = end

	// Сохранение в config.json
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
	// Базовая валидация cron
	if cronLine == "" {
		return fmt.Errorf("cron line cannot be empty")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Обновление конфигурации
	s.config.Client.CronLine = cronLine

	// Сохранение в config.json
	if err := s.saveConfig(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	s.logger.Info("Cron line updated", zap.String("cron_line", cronLine))
	return nil
}

// SetDailyLimit устанавливает дневной лимит
func (s *Service) SetDailyLimit(limit int) error {
	// Валидация лимита
	if limit <= 0 {
		return fmt.Errorf("daily limit must be positive, got %d", limit)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Обновление конфигурации
	s.config.Client.DailyLimitMinutes = limit

	// Сохранение в config.json
	if err := s.saveConfig(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	s.logger.Info("Daily limit updated", zap.Int("daily_limit_minutes", limit))
	return nil
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
