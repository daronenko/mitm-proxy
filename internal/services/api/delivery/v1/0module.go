package httpdelivery

import (
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"api.delivery.v1",
		fx.Invoke(Init),
	)
}
