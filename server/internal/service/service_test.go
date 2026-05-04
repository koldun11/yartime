package service

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/koldun11/yartime/server/config"
	"github.com/koldun11/yartime/server/internal/models"
	"go.uber.org/zap"
)

func TestNewService(t *testing.T) {
	cfg := &config.AppConfig{
		ConfigPath: "config.json",
		Server:     config.ServConfig{Port: "8080"},
		Client: config.ClientConfig{
			ClientID:          "test_client",
			DailyLimitMinutes: 180,
			AllowedHoursStart: "09:00",
			AllowedHoursEnd:   "21:00",
			CronLine:          "*/5 * * * *",
			ControlledApps:    []string{"firefox"},
		},
	}

	logger, _ := zap.NewDevelopment()
	svc := NewService(cfg, logger)

	if svc == nil {
		t.Fatal("Expected service to be created")
	}
}

func TestGetClientConfig(t *testing.T) {
	cfg := &config.AppConfig{
		ConfigPath: "config.json",
		Client: config.ClientConfig{
			ClientID:          "test_client",
			DailyLimitMinutes: 180,
			AllowedHoursStart: "09:00",
			AllowedHoursEnd:   "21:00",
			CronLine:          "*/5 * * * *",
			ControlledApps:    []string{"firefox", "steam"},
		},
	}

	logger, _ := zap.NewDevelopment()
	svc := NewService(cfg, logger)

	resp, err := svc.GetClientConfig()
	if err != nil {
		t.Fatalf("GetClientConfig failed: %v", err)
	}

	if resp.ClientID != "test_client" {
		t.Errorf("Expected ClientID=test_client, got=%s", resp.ClientID)
	}
	if resp.DailyLimitMinutes != 180 {
		t.Errorf("Expected DailyLimitMinutes=180, got=%d", resp.DailyLimitMinutes)
	}
	if len(resp.ControlledApps) != 2 {
		t.Errorf("Expected 2 controlled apps, got=%d", len(resp.ControlledApps))
	}
}

func TestGetClientConfig_NilControlledApps(t *testing.T) {
	cfg := &config.AppConfig{
		ConfigPath: "config.json",
		Client: config.ClientConfig{
			ClientID:       "test_client",
			ControlledApps: nil,
		},
	}

	logger, _ := zap.NewDevelopment()
	svc := NewService(cfg, logger)

	resp, err := svc.GetClientConfig()
	if err != nil {
		t.Fatalf("GetClientConfig failed: %v", err)
	}

	if resp.ControlledApps == nil {
		t.Error("Expected empty slice, got nil")
	}
}

func TestSetAllowedHours(t *testing.T) {
	tmpFile := "/tmp/test_config.json"
	defer os.Remove(tmpFile)

	cfg := &config.AppConfig{
		ConfigPath: tmpFile,
		Client:     config.ClientConfig{},
	}

	logger, _ := zap.NewDevelopment()
	svc := NewService(cfg, logger)

	err := svc.SetAllowedHours("09:00", "21:00")
	if err != nil {
		t.Fatalf("SetAllowedHours failed: %v", err)
	}

	cfg2, _ := config.NewAppConfig(tmpFile)
	if cfg2.Client.AllowedHoursStart != "09:00" {
		t.Errorf("Expected AllowedHoursStart=09:00, got=%s", cfg2.Client.AllowedHoursStart)
	}
	if cfg2.Client.AllowedHoursEnd != "21:00" {
		t.Errorf("Expected AllowedHoursEnd=21:00, got=%s", cfg2.Client.AllowedHoursEnd)
	}
}

func TestSetAllowedHours_InvalidFormat(t *testing.T) {
	tmpFile := "/tmp/test_config.json"
	defer os.Remove(tmpFile)

	cfg := &config.AppConfig{
		ConfigPath: tmpFile,
		Client:     config.ClientConfig{},
	}

	logger, _ := zap.NewDevelopment()
	svc := NewService(cfg, logger)

	err := svc.SetAllowedHours("25:00", "09:00")
	if err == nil {
		t.Fatal("Expected error for invalid time format")
	}

	err = svc.SetAllowedHours("09:60", "21:00")
	if err == nil {
		t.Fatal("Expected error for invalid time format")
	}
}

func TestSetAllowedHours_EqualTimes(t *testing.T) {
	tmpFile := "/tmp/test_config.json"
	defer os.Remove(tmpFile)

	cfg := &config.AppConfig{
		ConfigPath: tmpFile,
		Client:     config.ClientConfig{},
	}

	logger, _ := zap.NewDevelopment()
	svc := NewService(cfg, logger)

	err := svc.SetAllowedHours("10:00", "10:00")
	if err == nil {
		t.Fatal("Expected error when start equals end")
	}
}

func TestSetAllowedHours_CrossMidnight(t *testing.T) {
	tmpFile := "/tmp/test_config.json"
	defer os.Remove(tmpFile)

	cfg := &config.AppConfig{
		ConfigPath: tmpFile,
		Client:     config.ClientConfig{},
	}

	logger, _ := zap.NewDevelopment()
	svc := NewService(cfg, logger)

	err := svc.SetAllowedHours("23:00", "08:00")
	if err != nil {
		t.Fatalf("SetAllowedHours failed for cross-midnight: %v", err)
	}

	cfg2, _ := config.NewAppConfig(tmpFile)
	if cfg2.Client.AllowedHoursStart != "23:00" {
		t.Errorf("Expected AllowedHoursStart=23:00, got=%s", cfg2.Client.AllowedHoursStart)
	}
	if cfg2.Client.AllowedHoursEnd != "08:00" {
		t.Errorf("Expected AllowedHoursEnd=08:00, got=%s", cfg2.Client.AllowedHoursEnd)
	}
}

func TestSetCronLine(t *testing.T) {
	tmpFile := "/tmp/test_config.json"
	defer os.Remove(tmpFile)

	cfg := &config.AppConfig{
		ConfigPath: tmpFile,
		Client:     config.ClientConfig{},
	}

	logger, _ := zap.NewDevelopment()
	svc := NewService(cfg, logger)

	err := svc.SetCronLine("*/5 * * * *")
	if err != nil {
		t.Fatalf("SetCronLine failed: %v", err)
	}

	cfg2, _ := config.NewAppConfig(tmpFile)
	if cfg2.Client.CronLine != "*/5 * * * *" {
		t.Errorf("Expected CronLine=*/5 * * * *, got=%s", cfg2.Client.CronLine)
	}
}

func TestSetCronLine_Empty(t *testing.T) {
	tmpFile := "/tmp/test_config.json"
	defer os.Remove(tmpFile)

	cfg := &config.AppConfig{
		ConfigPath: tmpFile,
		Client:     config.ClientConfig{},
	}

	logger, _ := zap.NewDevelopment()
	svc := NewService(cfg, logger)

	err := svc.SetCronLine("")
	if err == nil {
		t.Fatal("Expected error for empty cron line")
	}
}

func TestSetDailyLimit(t *testing.T) {
	tmpFile := "/tmp/test_config.json"
	defer os.Remove(tmpFile)

	cfg := &config.AppConfig{
		ConfigPath: tmpFile,
		Client:     config.ClientConfig{},
	}

	logger, _ := zap.NewDevelopment()
	svc := NewService(cfg, logger)

	err := svc.SetDailyLimit(180)
	if err != nil {
		t.Fatalf("SetDailyLimit failed: %v", err)
	}

	cfg2, _ := config.NewAppConfig(tmpFile)
	if cfg2.Client.DailyLimitMinutes != 180 {
		t.Errorf("Expected DailyLimitMinutes=180, got=%d", cfg2.Client.DailyLimitMinutes)
	}
}

func TestSetDailyLimit_Invalid(t *testing.T) {
	tmpFile := "/tmp/test_config.json"
	defer os.Remove(tmpFile)

	cfg := &config.AppConfig{
		ConfigPath: tmpFile,
		Client:     config.ClientConfig{},
	}

	logger, _ := zap.NewDevelopment()
	svc := NewService(cfg, logger)

	err := svc.SetDailyLimit(0)
	if err == nil {
		t.Fatal("Expected error for zero limit")
	}

	err = svc.SetDailyLimit(-10)
	if err == nil {
		t.Fatal("Expected error for negative limit")
	}
}

func TestSetControlledApps(t *testing.T) {
	tmpFile := "/tmp/test_config.json"
	defer os.Remove(tmpFile)

	cfg := &config.AppConfig{
		ConfigPath: tmpFile,
		Client:     config.ClientConfig{},
	}

	logger, _ := zap.NewDevelopment()
	svc := NewService(cfg, logger)

	apps := []string{"firefox", "google-chrome", "steam"}
	err := svc.SetControlledApps(apps)
	if err != nil {
		t.Fatalf("SetControlledApps failed: %v", err)
	}

	cfg2, _ := config.NewAppConfig(tmpFile)
	if len(cfg2.Client.ControlledApps) != 3 {
		t.Errorf("Expected 3 controlled apps, got=%d", len(cfg2.Client.ControlledApps))
	}
	if cfg2.Client.ControlledApps[0] != "firefox" {
		t.Errorf("Expected first app=firefox, got=%s", cfg2.Client.ControlledApps[0])
	}
}

func TestSetControlledApps_Empty(t *testing.T) {
	tmpFile := "/tmp/test_config.json"
	defer os.Remove(tmpFile)

	cfg := &config.AppConfig{
		ConfigPath: tmpFile,
		Client:     config.ClientConfig{},
	}

	logger, _ := zap.NewDevelopment()
	svc := NewService(cfg, logger)

	err := svc.SetControlledApps([]string{})
	if err != nil {
		t.Fatalf("SetControlledApps failed: %v", err)
	}

	cfg2, _ := config.NewAppConfig(tmpFile)
	if cfg2.Client.ControlledApps == nil {
		t.Error("Expected empty slice, got nil")
	}
}

func TestGetVersion_NoBinaryDir(t *testing.T) {
	cfg := &config.AppConfig{
		ConfigPath: "config.json",
		Server:     config.ServConfig{BinaryDir: ""},
		Client:     config.ClientConfig{},
	}

	logger, _ := zap.NewDevelopment()
	svc := NewService(cfg, logger)

	v, err := svc.GetVersion()
	if err != nil {
		t.Fatalf("GetVersion failed: %v", err)
	}
	if v.Version != "0.0.0" {
		t.Errorf("Expected version 0.0.0, got=%s", v.Version)
	}
}

func TestGetVersion_WithVersionFile(t *testing.T) {
	tmpDir := "/tmp/test_binary_dir"
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	versionInfo := models.VersionInfo{Version: "1.2.3", Checksum: "abc123"}
	data, _ := json.Marshal(versionInfo)
	os.WriteFile(tmpDir+"/version.json", data, 0644)

	cfg := &config.AppConfig{
		ConfigPath: "config.json",
		Server:     config.ServConfig{BinaryDir: tmpDir},
		Client:     config.ClientConfig{},
	}

	logger, _ := zap.NewDevelopment()
	svc := NewService(cfg, logger)

	v, err := svc.GetVersion()
	if err != nil {
		t.Fatalf("GetVersion failed: %v", err)
	}
	if v.Version != "1.2.3" {
		t.Errorf("Expected version 1.2.3, got=%s", v.Version)
	}
	if v.Checksum != "abc123" {
		t.Errorf("Expected checksum abc123, got=%s", v.Checksum)
	}
}

func TestIsValidTime(t *testing.T) {
	tests := []struct {
		name     string
		timeStr  string
		expected bool
	}{
		{"Valid time", "09:00", true},
		{"Valid time", "23:59", true},
		{"Invalid hour", "24:00", false},
		{"Invalid minute", "09:60", false},
		{"Invalid format", "9:0", false},
		{"Invalid format", "09:0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidTime(tt.timeStr)
			if result != tt.expected {
				t.Errorf("isValidTime(%q) = %v, want %v", tt.timeStr, result, tt.expected)
			}
		})
	}
}

func TestService_Mutex(t *testing.T) {
	tmpFile := "/tmp/test_config.json"
	defer os.Remove(tmpFile)

	cfg := &config.AppConfig{
		ConfigPath: tmpFile,
		Client: config.ClientConfig{
			ClientID:          "test",
			DailyLimitMinutes: 180,
		},
	}

	logger, _ := zap.NewDevelopment()
	svc := NewService(cfg, logger)

	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			_, err := svc.GetClientConfig()
			if err != nil {
				t.Errorf("Concurrent read %d failed: %v", id, err)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	writeDone := make(chan bool)
	for i := 0; i < 5; i++ {
		go func(id int) {
			err := svc.SetAllowedHours("09:00", "21:00")
			if err != nil {
				t.Errorf("Concurrent write %d failed: %v", id, err)
			}
			writeDone <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		<-writeDone
	}

	close(done)
	close(writeDone)
}

func TestSetAllowedHours_TimeComparison(t *testing.T) {
	tmpFile := "/tmp/test_config.json"
	defer os.Remove(tmpFile)

	cfg := &config.AppConfig{
		ConfigPath: tmpFile,
		Client:     config.ClientConfig{},
	}

	logger, _ := zap.NewDevelopment()
	svc := NewService(cfg, logger)

	err := svc.SetAllowedHours("10:00", "20:00")
	if err != nil {
		t.Fatalf("SetAllowedHours failed: %v", err)
	}

	startTime, _ := time.Parse("15:04", "10:00")
	endTime, _ := time.Parse("15:04", "20:00")

	if !startTime.Before(endTime) {
		t.Error("Expected start time before end time")
	}
}

func TestComputeBinaryChecksum(t *testing.T) {
	tmpFile := "/tmp/test_binary"
	content := []byte("test binary content")
	os.WriteFile(tmpFile, content, 0755)
	defer os.Remove(tmpFile)

	checksum, err := ComputeBinaryChecksum(tmpFile)
	if err != nil {
		t.Fatalf("ComputeBinaryChecksum failed: %v", err)
	}

	if checksum == "" {
		t.Error("Expected non-empty checksum")
	}

	// Same content should produce same checksum
	checksum2, err := ComputeBinaryChecksum(tmpFile)
	if err != nil {
		t.Fatalf("ComputeBinaryChecksum failed: %v", err)
	}
	if checksum != checksum2 {
		t.Errorf("Expected same checksum, got %s and %s", checksum, checksum2)
	}
}
