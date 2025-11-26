package song

import (
	"fmt"
	"go-shazam/internal/queue"

	"go.uber.org/fx"
)

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

var QueueModule = fx.Module(
	"song-queue",
	fx.Provide(NewAddSongTaskHandler),
	fx.Invoke(func(w queue.WorkerServer, h *AddSongTaskHandler) {
		fmt.Printf("[Queue] Registering handler for task type: %s\n", AddSongTaskType)
		w.RegisterServiceHandler(AddSongTaskType, h)
	}),
)
