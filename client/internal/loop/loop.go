package loop

import (
	"reflect"
	"time"

	"github.com/koldun11/yartime/client/internal/api"
	"github.com/koldun11/yartime/client/internal/enforce"
	"github.com/koldun11/yartime/client/internal/store"
	"github.com/koldun11/yartime/client/internal/updater"
	"go.uber.org/zap"
)

// Run starts the main client loop. It polls the server for config updates,
// checks time restrictions, kills controlled apps, and checks for self-updates.
// The loop runs until the stop channel is closed.
func Run(
	cfg store.Config,
	st store.Store,
	apiClient *api.Client,
	enforcer enforce.Enforcer,
	interval time.Duration,
	currentVersion string,
	binaryPath string,
	logger *zap.Logger,
	stop <-chan struct{},
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	lastUpdateCheck := time.Time{}
	updateCheckInterval := time.Hour

	for {
		select {
		case <-stop:
			logger.Info("Loop stopped")
			return
		case <-ticker.C:
			tick(cfg, st, apiClient, enforcer, currentVersion, binaryPath,
				&lastUpdateCheck, updateCheckInterval, logger)
		}
	}
}

func tick(
	cfg store.Config,
	st store.Store,
	apiClient *api.Client,
	enforcer enforce.Enforcer,
	currentVersion string,
	binaryPath string,
	lastUpdateCheck *time.Time,
	updateCheckInterval time.Duration,
	logger *zap.Logger,
) {
	newCfg, err := apiClient.FetchConfig(cfg.ClientID)
	if err != nil {
		logger.Debug("Failed to sync config from server, using local", zap.Error(err))
	} else {
		if !configsEqual(cfg, newCfg) {
			newCfg.ServerURL = cfg.ServerURL
			if err := st.SaveConfig(newCfg); err != nil {
				logger.Error("Failed to save config", zap.Error(err))
			} else {
				cfg = newCfg
				logger.Info("Config synced from server")
				st.LogEvent(store.Event{
					Timestamp: time.Now(),
					Type:      "config_synced",
					Details:   "config updated from server",
				})
			}
		}
	}

	if err := checkTime(cfg, enforcer, st, logger); err != nil {
		logger.Error("Time check failed", zap.Error(err))
	}

	if len(cfg.ControlledApps) > 0 {
		if err := enforcer.KillApps(cfg.ControlledApps); err != nil {
			logger.Error("KillApps failed", zap.Error(err))
		}
	}

	if time.Since(*lastUpdateCheck) >= updateCheckInterval {
		*lastUpdateCheck = time.Now()
		if err := updater.CheckAndUpdate(currentVersion, cfg.ServerURL, binaryPath, logger); err != nil {
			logger.Debug("Update check failed", zap.Error(err))
		}
	}
}

func checkTime(cfg store.Config, enforcer enforce.Enforcer, st store.Store, logger *zap.Logger) error {
	now := time.Now()
	start, err := time.Parse("15:04", cfg.AllowedHoursStart)
	if err != nil {
		return err
	}
	end, err := time.Parse("15:04", cfg.AllowedHoursEnd)
	if err != nil {
		return err
	}

	start = time.Date(now.Year(), now.Month(), now.Day(), start.Hour(), start.Minute(), 0, 0, now.Location())
	end = time.Date(now.Year(), now.Month(), now.Day(), end.Hour(), end.Minute(), 0, 0, now.Location())

	outside := false
	if start.Before(end) {
		outside = now.Before(start) || now.After(end)
	} else {
		outside = now.After(end) && now.Before(start)
	}

	if outside {
		logger.Warn("Outside allowed hours, initiating shutdown",
			zap.String("now", now.Format("15:04")),
			zap.String("start", cfg.AllowedHoursStart),
			zap.String("end", cfg.AllowedHoursEnd))

		st.LogEvent(store.Event{
			Timestamp: now,
			Type:      "shutdown",
			Details:   "outside allowed hours",
		})

		return enforcer.Shutdown(1)
	}

	return nil
}

func configsEqual(a, b store.Config) bool {
	if a.ClientID != b.ClientID ||
		a.AllowedHoursStart != b.AllowedHoursStart ||
		a.AllowedHoursEnd != b.AllowedHoursEnd ||
		a.ExecuteOnStart != b.ExecuteOnStart {
		return false
	}
	return reflect.DeepEqual(a.ControlledApps, b.ControlledApps)
}
