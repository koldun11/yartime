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

const configPath = "config/config.json"

func main() {
	fx.New(
		fx.Provide(
			// Предоставляем *config.AppConfig
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
	).Run()
}
