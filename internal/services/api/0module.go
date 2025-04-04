package api

import (
	httpdelivery "github.com/daronenko/https-proxy/internal/services/api/delivery"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"api",
		httpdelivery.Module(),
	)
}
