package httpdelivery

import (
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"proxy.delivery",
		fx.Provide(New),
	)
}
