package store

import "time"

// Config represents client configuration.
// All fields except ServerURL are persisted to the local store.
type Config struct {
	ClientID          string   `json:"client_id"`
	AllowedHoursStart string   `json:"allowed_hours_start"`
	AllowedHoursEnd   string   `json:"allowed_hours_end"`
	ControlledApps    []string `json:"controlled_apps"`
	ExecuteOnStart    string   `json:"execute_on_start"`
	ServerURL         string   `json:"-"`
}

// Event represents a single log event stored locally.
type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Details   string    `json:"details"`
}

// Store defines the interface for local persistence.
// Implementations may use files, SQLite, or any other backend.
type Store interface {
	GetConfig() (Config, error)
	SaveConfig(Config) error
	LogEvent(Event) error
	GetEvents(date time.Time) ([]Event, error)
	Close() error
}

// DefaultConfig returns a fallback config used when no local data exists
// and the server is unreachable.
func DefaultConfig(clientID string) Config {
	return Config{
		ClientID:          clientID,
		AllowedHoursStart: "08:00",
		AllowedHoursEnd:   "21:00",
		ControlledApps:    []string{},
		ExecuteOnStart:    "",
	}
}
