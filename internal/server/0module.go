package server

import (
	"github.com/daronenko/https-proxy/internal/server/httprouter"
	"github.com/daronenko/https-proxy/internal/server/httpserver"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"server",

		fx.Provide(httpserver.NewProxyServer),

		fx.Provide(httprouter.New),
		fx.Provide(httpserver.NewApiServer),
	)
}
