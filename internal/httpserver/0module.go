package httpserver

import (
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"httpserver",

		fx.Provide(NewProxyServer),

		fx.Provide(NewApiRouter),
		fx.Provide(NewApiServer),
	)
}
