package httpdelivery

import (
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"proxy.delivery.v1",
		fx.Provide(New),
	)
}
