package song

import "go.uber.org/fx"

var Module = fx.Module(
	"song",
	fx.Provide(
		NewSongRepository,
		NewSongService,
	),
)

var HttpModule = fx.Module(
	"song-http",
	fx.Provide(NewSongHandler),
	fx.Invoke(RegisterRoutes),
)
