package fingerprint

import (
	"go.uber.org/fx"
)

var Module = fx.Module("fingerprint",
	fx.Provide(
		NewRepository,
		NewFingerprintService,
	),
)
