package spotify

import "go.uber.org/fx"

var Module = fx.Module("spotify",
	fx.Provide(NewSpotifySongMetadataSource, LoadConfig),
)
