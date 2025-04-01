package proxy

import (
	httpdelivery "github.com/daronenko/https-proxy/internal/services/proxy/delivery"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"proxy",
		httpdelivery.Module(),
	)
}
