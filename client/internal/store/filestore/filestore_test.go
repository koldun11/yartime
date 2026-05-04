package filestore

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/koldun11/yartime/client/internal/store"
)

func tmpDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return dir
}

func TestOpen(t *testing.T) {
	dir := tmpDir(t)
	fs, err := Open(dir)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	if fs == nil {
		t.Fatal("Expected non-nil FileStore")
	}
	defer fs.Close()

	if _, err := os.Stat(filepath.Join(dir, "logs")); os.IsNotExist(err) {
		t.Fatal("Expected logs directory to be created")
	}
}

func TestGetConfig_NoFile(t *testing.T) {
	dir := tmpDir(t)
	fs, _ := Open(dir)

	cfg, err := fs.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}

	if cfg.ClientID != "default-client" {
		t.Errorf("Expected default-client, got=%s", cfg.ClientID)
	}
	if cfg.AllowedHoursStart != "08:00" {
		t.Errorf("Expected 08:00, got=%s", cfg.AllowedHoursStart)
	}
}

func TestSaveAndGetConfig(t *testing.T) {
	dir := tmpDir(t)
	fs, _ := Open(dir)

	cfg := store.Config{
		ClientID:          "test_child",
		AllowedHoursStart: "10:00",
		AllowedHoursEnd:   "20:00",
		ControlledApps:    []string{"firefox", "steam"},
		ExecuteOnStart:    "echo hello",
		ServerURL:         "http://example.com",
	}

	if err := fs.SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	got, err := fs.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}

	if got.ClientID != "test_child" {
		t.Errorf("ClientID: expected test_child, got=%s", got.ClientID)
	}
	if got.AllowedHoursStart != "10:00" {
		t.Errorf("AllowedHoursStart: expected 10:00, got=%s", got.AllowedHoursStart)
	}
	if got.AllowedHoursEnd != "20:00" {
		t.Errorf("AllowedHoursEnd: expected 20:00, got=%s", got.AllowedHoursEnd)
	}
	if len(got.ControlledApps) != 2 {
		t.Errorf("ControlledApps: expected 2, got=%d", len(got.ControlledApps))
	}
	if got.ExecuteOnStart != "echo hello" {
		t.Errorf("ExecuteOnStart: expected 'echo hello', got=%s", got.ExecuteOnStart)
	}
}

func TestGetConfig_CorruptedFile(t *testing.T) {
	dir := tmpDir(t)
	fs, _ := Open(dir)

	os.WriteFile(fs.configPath, []byte("not json{"), 0600)

	cfg, err := fs.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig should not error on corrupted file: %v", err)
	}

	if cfg.ClientID != "default-client" {
		t.Errorf("Expected default fallback, got=%s", cfg.ClientID)
	}
}

func TestLogEvent(t *testing.T) {
	dir := tmpDir(t)
	fs, _ := Open(dir)

	ev := store.Event{
		Timestamp: time.Date(2026, 5, 4, 15, 30, 0, 0, time.Local),
		Type:      "shutdown",
		Details:   "outside allowed hours",
	}

	if err := fs.LogEvent(ev); err != nil {
		t.Fatalf("LogEvent failed: %v", err)
	}

	events, err := fs.GetEvents(time.Date(2026, 5, 4, 0, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("GetEvents failed: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got=%d", len(events))
	}
	if events[0].Type != "shutdown" {
		t.Errorf("Expected type shutdown, got=%s", events[0].Type)
	}
	if events[0].Details != "outside allowed hours" {
		t.Errorf("Expected details 'outside allowed hours', got=%s", events[0].Details)
	}
}

func TestLogEvent_MultipleDays(t *testing.T) {
	dir := tmpDir(t)
	fs, _ := Open(dir)

	day1 := time.Date(2026, 5, 3, 10, 0, 0, 0, time.Local)
	day2 := time.Date(2026, 5, 4, 10, 0, 0, 0, time.Local)

	fs.LogEvent(store.Event{Timestamp: day1, Type: "error", Details: "day1"})
	fs.LogEvent(store.Event{Timestamp: day2, Type: "error", Details: "day2"})

	events1, _ := fs.GetEvents(day1)
	events2, _ := fs.GetEvents(day2)

	if len(events1) != 1 {
		t.Errorf("Expected 1 event for day1, got=%d", len(events1))
	}
	if len(events2) != 1 {
		t.Errorf("Expected 1 event for day2, got=%d", len(events2))
	}

	files, _ := fs.ListEventFiles()
	if len(files) != 2 {
		t.Errorf("Expected 2 log files, got=%d", len(files))
	}
}

func TestGetEvents_NoFile(t *testing.T) {
	dir := tmpDir(t)
	fs, _ := Open(dir)

	events, err := fs.GetEvents(time.Now())
	if err != nil {
		t.Fatalf("GetEvents should not error on missing file: %v", err)
	}
	if events != nil {
		t.Errorf("Expected nil events, got %v", events)
	}
}

func TestLogEvent_Rotation(t *testing.T) {
	dir := tmpDir(t)
	fs, _ := Open(dir)

	oldDate := time.Now().AddDate(0, 0, -(retentionDays + 5))
	fs.LogEvent(store.Event{Timestamp: oldDate, Type: "old"})

	recentDate := time.Now().AddDate(0, 0, -5)
	fs.LogEvent(store.Event{Timestamp: recentDate, Type: "recent"})

	fs.LogEvent(store.Event{Timestamp: time.Now(), Type: "today"})

	files, _ := fs.ListEventFiles()

	if len(files) != 2 {
		t.Errorf("Expected 2 files after rotation (old should be deleted), got=%d", len(files))
	}

	for _, f := range files {
		if f == "events-"+oldDate.Format("2006-01-02")+".jsonl" {
			t.Error("Old log file should have been deleted")
		}
	}
}

func TestClose(t *testing.T) {
	dir := tmpDir(t)
	fs, _ := Open(dir)

	if err := fs.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}
