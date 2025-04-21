package repo

import (
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"api.delivery",
		fx.Provide(
			New,
		),
	)
}
