package app

import (
	"log"
	"os"
	"time"

	"github.com/daronenko/https-proxy/internal/app/config"
	"github.com/daronenko/https-proxy/internal/httpserver"
	"github.com/daronenko/https-proxy/internal/infra"
	"github.com/daronenko/https-proxy/internal/services/api"
	"github.com/daronenko/https-proxy/internal/services/proxy"
	"github.com/daronenko/https-proxy/pkg/logger"
	"go.uber.org/fx"
)

func Options(extraOpts ...fx.Option) []fx.Option {
	conf, err := config.New()
	if err != nil {
		log.Printf("error: failed to parse config: %v\n", err)
		os.Exit(1)
	}

	logger.Setup(&conf.App.Logger)

	baseOpts := []fx.Option{
		fx.Supply(conf),

		fx.WithLogger(logger.Fx),

		infra.Module(),
		httpserver.Module(),

		proxy.Module(),
		api.Module(),

		fx.StartTimeout(1 * time.Second),
	}

	return append(baseOpts, extraOpts...)
}

func New(extraOpts ...fx.Option) *fx.App {
	return fx.New(Options(extraOpts...)...)
}
