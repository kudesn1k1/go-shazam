package song

import "context"

type SongMetadataSource interface {
	GetSongsMetadata(ctx context.Context, link string) (*SongMetadata, error)
}

type SongDownloader interface {
	DownloadSong(ctx context.Context, data *SongMetadata, dir string) (*DownloadedSong, error)
}

type SongService struct {
	songMetadataSource SongMetadataSource
	songDownloader     SongDownloader
}

func NewSongService(songMetadataSource SongMetadataSource, songDownloader SongDownloader) *SongService {
	return &SongService{
		songMetadataSource: songMetadataSource,
		songDownloader:     songDownloader,
	}
}

func (s *SongService) GetSongsMetadata(ctx context.Context, link string) (*SongMetadata, error) {
	songMeta, err := s.songMetadataSource.GetSongsMetadata(ctx, link)
	s.songDownloader.DownloadSong(ctx, songMeta, "/tmp")

	return songMeta, err
}
