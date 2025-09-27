package youtube

import "go.uber.org/fx"

var Module = fx.Module("youtube",
	fx.Provide(NewYoutubeSongDownloader, LoadConfig),
)
