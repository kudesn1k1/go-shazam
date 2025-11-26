package song

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-shazam/internal/audio"
	"go-shazam/internal/core/db"
	"go-shazam/internal/fingerprint"
	"go-shazam/internal/queue"
	"go-shazam/internal/utils/converter"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

var ErrSongTaskAlreadyExists = errors.New("Song is already being processed")

type SongMetadataSource interface {
	GetSongMetadata(ctx context.Context, sourceID string) (*SongMetadata, error)
	ExtractSourceID(link string) (string, error)
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
	queue              queue.QueueService
}

func NewSongService(
	songMetadataSource SongMetadataSource,
	songDownloader SongDownloader,
	songRepository SongRepositoryInterface,
	fingerprintService *fingerprint.FingerprintService,
	transactionManager *db.TransactionManager,
	queue queue.QueueService,
) *SongService {
	return &SongService{
		songMetadataSource: songMetadataSource,
		songDownloader:     songDownloader,
		songRepository:     songRepository,
		fingerprintService: fingerprintService,
		transactionManager: transactionManager,
		queue:              queue,
	}
}

func (s *SongService) GetSongMetadata(ctx context.Context, sourceID string) (*SongMetadata, error) {
	return s.songMetadataSource.GetSongMetadata(ctx, sourceID)
}

func (s *SongService) EnqueueSong(ctx context.Context, link string) error {
	sourceID, err := s.songMetadataSource.ExtractSourceID(link)
	if err != nil {
		return fmt.Errorf("failed to extract source ID from link: %w", err)
	}

	payload, err := json.Marshal(AddSongTaskPayload{ID: sourceID})
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	if _, err = s.queue.Enqueue(AddSongTaskType, payload, asynq.TaskID(sourceID)); err != nil {
		if errors.Is(err, asynq.ErrTaskIDConflict) {
			return ErrSongTaskAlreadyExists
		}
		return fmt.Errorf("failed to enqueue song task: %w", err)
	}

	return nil
}

func (s *SongService) AddSong(ctx context.Context, sourceID string) (*SongMetadata, error) {
	songMeta, err := s.songMetadataSource.GetSongMetadata(ctx, sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get song metadata: %w", err)
	}

	existingSong, err := s.songRepository.FindByTitleAndArtist(ctx, songMeta.Title, songMeta.Artist)
	if err != nil {
		return nil, err
	}
	if existingSong != nil {
		return &SongMetadata{
			Title:      existingSong.Title,
			Artist:     existingSong.Artist,
			DurationMs: existingSong.Duration,
		}, nil
	}

	downloadedSong, err := s.songDownloader.DownloadSong(ctx, songMeta, os.TempDir())
	if err != nil {
		return nil, fmt.Errorf("failed to download song: %w", err)
	}

	// Convert to WAV
	fullPath := filepath.Join(downloadedSong.Path, downloadedSong.Filename)
	defer os.Remove(fullPath) // Cleanup downloaded file

	wavPath, err := converter.ConvertToWav(fullPath, audio.TargetSampleRate)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to wav: %w", err)
	}
	defer os.Remove(wavPath) // Cleanup converted WAV file

	// Load WAV
	samples, sampleRate, err := audio.LoadWav(wavPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load wav: %w", err)
	}

	// Process audio (FFT) - CPU bound, done outside transaction
	fragments, err := audio.ProcessAudio(samples, sampleRate)
	if err != nil {
		return nil, fmt.Errorf("failed to process audio: %w", err)
	}

	songID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate uuid: %w", err)
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
		if err := s.songRepository.Save(txCtx, songEntity); err != nil {
			return nil, fmt.Errorf("failed to save song: %w", err)
		}

		if err := s.fingerprintService.SaveFingerprints(txCtx, hashes); err != nil {
			return nil, fmt.Errorf("failed to save fingerprints: %w", err)
		}

		return nil, nil
	})

	if err != nil {
		return nil, err
	}

	fmt.Printf("Processed %d fragments and saved %d hashes for song %s\n", len(fragments), len(hashes), songMeta.Title)

	return songMeta, nil
}
