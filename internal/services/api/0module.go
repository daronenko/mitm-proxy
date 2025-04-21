package api

import (
	httpdelivery "github.com/daronenko/https-proxy/internal/services/api/delivery"
	"github.com/daronenko/https-proxy/internal/services/api/repo"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"api",
		repo.Module(),
		httpdelivery.Module(),
	)
}
