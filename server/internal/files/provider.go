package files

import "go.uber.org/fx"

var Module = fx.Module("files",
	fx.Provide(
		LoadConfig,
		NewS3Client,
		NewFilesRepository,
		NewFilesService,
	),
)

var HttpModule = fx.Module("files-http",
	fx.Provide(NewFilesHandler),
	fx.Invoke(RegisterRoutes),
)

var CleanupModule = fx.Module("files-cleanup",
	fx.Provide(NewCleanupTaskHandler),
	fx.Invoke(RegisterCleanupHandler),
	fx.Invoke(RegisterCleanupSchedule),
)
