package song

type SongMetadata struct {
	Title      string
	Artist     string
	DurationMs int
}

type DownloadedSong struct {
	Filename string
	Path     string
}

type GetSongRequest struct {
	Link string `json:"link" binding:"required"`
}
