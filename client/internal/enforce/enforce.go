package enforce

// Enforcer defines the interface for system enforcement actions.
// Implementations are platform-specific (linux, windows, etc.)
type Enforcer interface {
	// Shutdown initiates a system shutdown after the given number of minutes.
	Shutdown(minutes int) error

	// KillApps terminates processes matching the given application names.
	KillApps(apps []string) error
}
