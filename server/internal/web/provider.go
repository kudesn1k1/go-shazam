package web

import "go.uber.org/fx"

var Module = fx.Module("web",
	fx.Provide(LoadConfig, NewWebHandler),
)

var HttpModule = fx.Module("web-http",
	fx.Invoke(RegisterRoutes),
)
