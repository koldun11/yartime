package filestore

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/koldun11/yartime/client/internal/store"
)

const (
	configFile    = "config.json"
	logsDir       = "logs"
	retentionDays = 30
)

// FileStore implements store.Store using the local filesystem.
// Config is stored as JSON. Events are stored as JSONL files, one per day.
type FileStore struct {
	baseDir    string
	configPath string
	logsDir    string
}

// Open creates or opens a FileStore at the given base directory.
// The directory is created if it doesn't exist.
func Open(baseDir string) (*FileStore, error) {
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create base dir: %w", err)
	}

	logsDir := filepath.Join(baseDir, logsDir)
	if err := os.MkdirAll(logsDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create logs dir: %w", err)
	}

	return &FileStore{
		baseDir:    baseDir,
		configPath: filepath.Join(baseDir, configFile),
		logsDir:    logsDir,
	}, nil
}

// GetConfig reads the config from disk. Returns DefaultConfig if the file
// doesn't exist or is corrupted.
func (fs *FileStore) GetConfig() (store.Config, error) {
	data, err := os.ReadFile(fs.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return store.DefaultConfig("default-client"), nil
		}
		return store.Config{}, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg store.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return store.DefaultConfig("default-client"), nil
	}

	return cfg, nil
}

// SaveConfig writes the config to disk with 0600 permissions.
func (fs *FileStore) SaveConfig(cfg store.Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(fs.configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// LogEvent appends an event as a JSON line to the day's log file.
// Old log files (older than retentionDays) are deleted.
func (fs *FileStore) LogEvent(ev store.Event) error {
	if ev.Timestamp.IsZero() {
		ev.Timestamp = time.Now()
	}

	logFile := fs.eventFilePath(ev.Timestamp)

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer f.Close()

	line, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if _, err := f.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	fs.rotateOldLogs()

	return nil
}

// GetEvents returns all events for the given date.
func (fs *FileStore) GetEvents(date time.Time) ([]store.Event, error) {
	logFile := fs.eventFilePath(date)

	f, err := os.Open(logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer f.Close()

	var events []store.Event
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var ev store.Event
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			continue
		}
		events = append(events, ev)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read log file: %w", err)
	}

	return events, nil
}

// Close is a no-op for FileStore (files are closed after each operation).
func (fs *FileStore) Close() error {
	return nil
}

// eventFilePath returns the path to the JSONL log file for the given date.
func (fs *FileStore) eventFilePath(t time.Time) string {
	name := fmt.Sprintf("events-%s.jsonl", t.Format("2006-01-02"))
	return filepath.Join(fs.logsDir, name)
}

// rotateOldLogs deletes log files older than retentionDays.
func (fs *FileStore) rotateOldLogs() {
	entries, err := os.ReadDir(fs.logsDir)
	if err != nil {
		return
	}

	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "events-") || !strings.HasSuffix(name, ".jsonl") {
			continue
		}

		dateStr := strings.TrimPrefix(name, "events-")
		dateStr = strings.TrimSuffix(dateStr, ".jsonl")
		fileDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}

		if fileDate.Before(cutoff) {
			os.Remove(filepath.Join(fs.logsDir, name))
		}
	}
}

// ListEventFiles returns the names of all event log files (for testing).
func (fs *FileStore) ListEventFiles() ([]string, error) {
	entries, err := os.ReadDir(fs.logsDir)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".jsonl") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}
