package main

import (
	"context"
	"fmt"
	"net"

	"github.com/daronenko/https-proxy/internal/app"
	"github.com/daronenko/https-proxy/internal/app/config"
	"github.com/daronenko/https-proxy/internal/server/httpserver"
	"github.com/rs/zerolog/log"
	"go.uber.org/fx"
)

func main() {
	app.New(fx.Invoke(run)).Run()
}

func run(proxyServer *httpserver.ProxyServer, apiServer *httpserver.ApiServer, conf *config.Config, lc fx.Lifecycle) {
	proxyListener, err := net.Listen("tcp", conf.App.ProxyServer.Address)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start proxy listener")
	}

	apiListener, err := net.Listen("tcp", conf.App.ApiServer.Address)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start api listener")
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				log.Info().Msg(fmt.Sprintf("starting proxy server on %s...", conf.App.ProxyServer.Address))
				proxyServer.Serve(proxyListener)
			}()

			go func() {
				log.Info().Msg(fmt.Sprintf("starting api server on %s...", conf.App.ApiServer.Address))
				apiServer.Serve(apiListener)
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info().Msg("shutting down servers...")

			// if err := proxyServer.Shutdown(ctx); err != nil {
			// 	log.Error().Err(err).Msg("failed to shutdown proxy server")
			// }

			if err := apiServer.Shutdown(ctx); err != nil {
				log.Error().Err(err).Msg("failed to shutdown api server")
			}

			return nil
		},
	})
}
