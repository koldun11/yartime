package run

import (
	"context"
	"github.com/koldun11/yartime/server/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type App struct {
	conf   *config.AppConfig
	logger *zap.Logger

	// TODO: implement
}

func NewApp(conf *config.AppConfig, logger *zap.Logger) *App {
	return &App{
		conf:   conf,
		logger: logger,
	}
}

func Start(lc fx.Lifecycle, app *App) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return app.start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return app.stop(ctx)
		},
	})
}

// start при запуске приложения
func (a *App) start(ctx context.Context) error {
	// TODO: Реализовать запуск HTTP-сервера
	a.logger.Info("Starting server", // TODO: заменить на актуальные параметры
		zap.String("port", a.conf.Server.Port),
		zap.String("client_id", a.conf.Client.ClientID),
		zap.Int("daily_limit", a.conf.Client.DailyLimitMinutes))
	return nil
}

// stop при остановке приложения
func (a *App) stop(ctx context.Context) error {
	a.logger.Info("Stopping server")
	return nil
}
