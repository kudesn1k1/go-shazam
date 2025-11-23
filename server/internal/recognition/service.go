package recognition

import (
	"context"
	"fmt"
	"go-shazam/internal/audio"
	"go-shazam/internal/fingerprint"
	"go-shazam/internal/logger"
	"go-shazam/internal/song"
	"math"

	"github.com/google/uuid"
)

type RecognitionService struct {
	fingerprintService *fingerprint.FingerprintService
	songRepository     song.SongRepositoryInterface
}

func NewRecognitionService(
	fingerprintService *fingerprint.FingerprintService,
	songRepository song.SongRepositoryInterface,
) *RecognitionService {
	return &RecognitionService{
		fingerprintService: fingerprintService,
		songRepository:     songRepository,
	}
}

// TODO: return song metadata instead of song entity
type MatchResult struct {
	Song       *song.SongEntity `json:"song"`
	TimeOffset float64          `json:"time_offset"`
	Score      int              `json:"score"`
}

// IdentifySong processes audio fragments and returns the best matching song.
func (s *RecognitionService) IdentifySong(ctx context.Context, fragments []audio.ProcessedFragment, sampleRate int) (*MatchResult, error) {
	// Use a dummy ID since we don't know the song yet
	logger := logger.FromContext(ctx)

	dummyID := uuid.Nil
	sampleHashes := s.fingerprintService.CreateFingerprints(fragments, dummyID, sampleRate)

	if len(sampleHashes) == 0 {
		return nil, fmt.Errorf("no fingerprints generated from audio")
	}

	dbHashes, err := s.fingerprintService.GetMatchingHashes(ctx, sampleHashes)
	if err != nil {
		return nil, fmt.Errorf("failed to get matching hashes: %w", err)
	}

	if len(dbHashes) == 0 {
		return nil, nil
	}

	scores := make(map[uuid.UUID]map[int]int)

	sampleHashMap := make(map[uint32][]float64)
	for _, h := range sampleHashes {
		sampleHashMap[h.HashValue] = append(sampleHashMap[h.HashValue], h.TimeOffset)
	}

	bestScore := 0
	var bestSongID uuid.UUID
	var bestTimeOffset float64

	for _, dbHash := range dbHashes {
		if sampleOffsets, ok := sampleHashMap[dbHash.HashValue]; ok {
			for _, sampleOffset := range sampleOffsets {
				diff := dbHash.TimeOffset - sampleOffset

				bin := int(math.Round(diff * 20)) // 50ms bins

				if scores[dbHash.SongID] == nil {
					scores[dbHash.SongID] = make(map[int]int)
				}

				scores[dbHash.SongID][bin]++
				count := scores[dbHash.SongID][bin]

				if count > bestScore {
					bestScore = count
					bestSongID = dbHash.SongID
					bestTimeOffset = dbHash.TimeOffset
				}
			}
		}
	}

	logger.Info(fmt.Sprintf("Recognition result: BestScore=%d, TotalCandidates=%d", bestScore, len(scores)))

	if bestScore < 5 { // Minimum threshold to consider it a match
		logger.Info("Score too low", "bestScore", bestScore)
		return nil, nil
	}

	songEntity, err := s.songRepository.FindByID(ctx, bestSongID)
	if err != nil {
		return nil, fmt.Errorf("failed to find song %s: %w", bestSongID, err)
	}

	return &MatchResult{
		Song:       songEntity,
		TimeOffset: bestTimeOffset,
		Score:      bestScore,
	}, nil
}
