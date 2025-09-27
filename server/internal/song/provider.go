package song

import "go.uber.org/fx"

var Module = fx.Module(
	"song",
	fx.Provide(NewSongService, NewSongHandler),
	fx.Invoke(func(h *SongHandler) {}),
)
