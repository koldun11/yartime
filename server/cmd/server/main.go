package main

import (
	"github.com/koldun11/yartime/server/config"
	"github.com/koldun11/yartime/server/internal/handler"
	"github.com/koldun11/yartime/server/internal/infrastructure/server"
	"github.com/koldun11/yartime/server/internal/router"
	"github.com/koldun11/yartime/server/internal/service"
	"github.com/koldun11/yartime/server/run"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var version = "0.0.0"

const configPath = "config/config.json"

func main() {
	fx.New(
		fx.Provide(
			func() (*config.AppConfig, error) {
				return config.NewAppConfig(configPath)
			},
			zap.NewDevelopment,
			run.NewApp,
			service.NewService,
			handler.NewHandler,
			router.NewRouter,
			server.NewServer,
		),
		fx.Invoke(run.Start),
		fx.WithLogger(func() *zap.Logger {
			logger, _ := zap.NewDevelopment()
			logger.Info("Starting yartime server", zap.String("version", version))
			return logger
		}),
	).Run()
}
