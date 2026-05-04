package enforce

import (
	"fmt"
	"os/exec"
	"time"

	"go.uber.org/zap"
)

// LinuxEnforcer implements Enforcer for Linux systems.
type LinuxEnforcer struct {
	logger *zap.Logger
}

// NewLinuxEnforcer creates a new LinuxEnforcer.
func NewLinuxEnforcer(logger *zap.Logger) *LinuxEnforcer {
	return &LinuxEnforcer{logger: logger}
}

// Shutdown initiates a system shutdown after the given number of minutes.
// Uses the `shutdown` command which requires appropriate permissions.
func (e *LinuxEnforcer) Shutdown(minutes int) error {
	e.logger.Warn("Initiating system shutdown", zap.Int("minutes", minutes))

	cmd := exec.Command("shutdown", "-h", fmt.Sprintf("+%d", minutes))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("shutdown command failed: %w", err)
	}

	return nil
}

// KillApps terminates processes matching the given application names.
// For each app, it first sends SIGTERM via `pkill`, waits 5 seconds,
// then sends SIGKILL via `pkill -9` if the process is still running.
func (e *LinuxEnforcer) KillApps(apps []string) error {
	for _, app := range apps {
		if app == "" {
			continue
		}

		cmd := exec.Command("pgrep", app)
		if err := cmd.Run(); err != nil {
			continue
		}

		e.logger.Info("Terminating app", zap.String("app", app))

		cmd = exec.Command("pkill", app)
		if err := cmd.Run(); err != nil {
			e.logger.Debug("pkill failed, trying SIGKILL", zap.String("app", app), zap.Error(err))
		}

		time.Sleep(5 * time.Second)

		cmd = exec.Command("pgrep", app)
		if err := cmd.Run(); err == nil {
			e.logger.Warn("Process still alive after SIGTERM, sending SIGKILL", zap.String("app", app))
			exec.Command("pkill", "-9", app).Run()
		}
	}

	return nil
}
