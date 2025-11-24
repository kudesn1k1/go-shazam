package queue

import "go.uber.org/fx"

var Module = fx.Module("queue",
	fx.Provide(
		NewConfig,
		NewQueueService,
		NewWorkerServer,
	),
)
