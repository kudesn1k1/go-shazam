package recognition

import "go.uber.org/fx"

var Module = fx.Module("recognition",
	fx.Provide(NewRecognitionService),
)

var HttpModule = fx.Module("recognition-http",
	fx.Provide(NewRecognitionHandler),
	fx.Invoke(RegisterRoutes),
)
