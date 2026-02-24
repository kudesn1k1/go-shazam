package song

type SongMetadata struct {
	Title      string
	Artist     string
	DurationMs int
}

type DownloadedSong struct {
	Filename string
	Path     string
	SourceID string
}

type GetSongRequest struct {
	Link string `json:"link" binding:"required"`
}

type SongResponse struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Artist   string  `json:"artist"`
	Duration int     `json:"duration"`
	SourceID string  `json:"source_id"`
	UploadedBy *string `json:"uploaded_by,omitempty"`
}
