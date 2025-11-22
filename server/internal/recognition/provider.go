package recognition

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewRecognitionService, NewRecognitionHandler),
	fx.Invoke(func(h *RecognitionHandler) {}),
)
