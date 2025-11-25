package user

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"user",
	fx.Provide(
		NewUserRepository,
		NewUserService,
		NewCryptoService,
	),
)

var HttpModule = fx.Module(
	"user-http",
	fx.Provide(NewUserHandler),
	fx.Invoke(RegisterRoutes),
)
