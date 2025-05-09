package main

import (
	"github.com/koldun11/yartime/server/config"
	"github.com/koldun11/yartime/server/run"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		fx.Provide(
			// Предоставляем *config.AppConfig
			func() (*config.AppConfig, error) {
				return config.NewAppConfig("config/config.json") // TODO: заменить
			},
			zap.NewDevelopment,
			run.NewApp,

			//handler.NewHandler, // TODO: implement
			//service.NewService, // TODO: implement
		),
		fx.Invoke(run.Start),
	).Run()
}
