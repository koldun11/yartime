package run

import (
	"context"
	"github.com/koldun11/yartime/server/config"
	"github.com/koldun11/yartime/server/internal/infrastructure/server"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type App struct {
	conf   *config.AppConfig
	logger *zap.Logger
	server *server.Server

	// TODO: implement
}

func NewApp(conf *config.AppConfig, logger *zap.Logger, server *server.Server) *App {
	return &App{
		conf:   conf,
		logger: logger,
		server: server,
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
	return a.server.Start()
}

// stop при остановке приложения
func (a *App) stop(ctx context.Context) error {
	return a.server.Stop(ctx)
}
