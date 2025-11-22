package song

import (
	"context"
	"fmt"
	"go-shazam/internal/audio"
	"go-shazam/internal/core/db"
	"go-shazam/internal/fingerprint"
	"go-shazam/internal/utils/converter"
	"os"
	"path/filepath"

	"github.com/google/uuid"
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
	songRepository     SongRepositoryInterface
	fingerprintService *fingerprint.FingerprintService
	transactionManager *db.TransactionManager
}

func NewSongService(
	songMetadataSource SongMetadataSource,
	songDownloader SongDownloader,
	songRepository SongRepositoryInterface,
	fingerprintService *fingerprint.FingerprintService,
	transactionManager *db.TransactionManager,
) *SongService {
	return &SongService{
		songMetadataSource: songMetadataSource,
		songDownloader:     songDownloader,
		songRepository:     songRepository,
		fingerprintService: fingerprintService,
		transactionManager: transactionManager,
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

	// 1. Convert to WAV
	fullPath := filepath.Join(downloadedSong.Path, downloadedSong.Filename)
	wavPath, err := converter.ConvertToWav(fullPath, audio.TargetSampleRate)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert to wav: %w", err)
	}

	// 2. Load WAV
	samples, sampleRate, err := audio.LoadWav(wavPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load wav: %w", err)
	}

	// 3. Process (FFT) (CPU bound - do outside transaction)
	fragments, err := audio.ProcessAudio(samples, sampleRate)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to process audio: %w", err)
	}

	songID, err := uuid.NewV7()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate uuid: %w", err)
	}

	songEntity := &SongEntity{
		ID:       songID,
		Title:    songMeta.Title,
		Artist:   songMeta.Artist,
		Duration: songMeta.DurationMs,
		SourceID: downloadedSong.SourceID,
	}

	// Calculate fingerprints (CPU bound)
	hashes := s.fingerprintService.CreateFingerprints(fragments, songID, sampleRate)

	_, err = db.Transactional(ctx, s.transactionManager, func(txCtx context.Context) (interface{}, error) {
		// Save Song
		if err := s.songRepository.Save(txCtx, songEntity); err != nil {
			return nil, fmt.Errorf("failed to save song: %w", err)
		}

		// Save Fingerprints
		if err := s.fingerprintService.SaveFingerprints(txCtx, hashes); err != nil {
			return nil, fmt.Errorf("failed to save fingerprints: %w", err)
		}

		return nil, nil
	})

	if err != nil {
		return nil, nil, err
	}

	fmt.Printf("Processed %d fragments and saved %d hashes for song %s\n", len(fragments), len(hashes), songMeta.Title)

	return songMeta, downloadedSong, nil
}
