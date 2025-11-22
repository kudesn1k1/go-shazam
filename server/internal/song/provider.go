package song

import "go.uber.org/fx"

var Module = fx.Module(
	"song",
	fx.Provide(
		NewSongRepository,
		NewSongService,
		NewSongHandler,
	),
	fx.Invoke(func(h *SongHandler) {}),
)
