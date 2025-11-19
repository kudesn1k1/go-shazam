package song

import (
	"context"
	"os"
)

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

	return songMeta, err
}

func (s *SongService) DownloadSong(ctx context.Context, link string) (*SongMetadata, *DownloadedSong, error) {
	songMeta, err := s.songMetadataSource.GetSongsMetadata(ctx, link)
	if err != nil {
		return nil, nil, err
	}
	downloadedSong, err := s.songDownloader.DownloadSong(ctx, songMeta, os.TempDir())
	if err != nil {
		return nil, nil, err
	}

	return songMeta, downloadedSong, nil
}
