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

const (
	TimeBinResolution = 20    // 50ms bins (1/0.05)
	MinAbsoluteScore  = 5     // Minimum absolute score to consider a match
	MinScoreRatio     = 0.015 // Minimum score as ratio of sample hashes (1.5%)
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
	log := logger.FromContext(ctx)

	// Generate fingerprints from sample (use Nil UUID since we don't know the song)
	sampleHashes := s.fingerprintService.CreateFingerprints(fragments, uuid.Nil, sampleRate)

	if len(sampleHashes) == 0 {
		return nil, fmt.Errorf("no fingerprints generated from audio")
	}

	dbHashes, err := s.fingerprintService.GetMatchingHashes(ctx, sampleHashes)
	if err != nil {
		return nil, fmt.Errorf("failed to get matching hashes: %w", err)
	}

	if len(dbHashes) == 0 {
		log.Info("No matching hashes found in database")
		return nil, nil
	}

	sampleHashMap := make(map[int64][]float64)
	for _, h := range sampleHashes {
		sampleHashMap[h.HashValue] = append(sampleHashMap[h.HashValue], h.TimeOffset)
	}

	// scores[songID][timeBin] = count
	scores := make(map[uuid.UUID]map[int]int)

	bestScore := 0
	var bestSongID uuid.UUID
	var bestTimeOffset float64

	for _, dbHash := range dbHashes {
		sampleOffsets, ok := sampleHashMap[dbHash.HashValue]
		if !ok {
			continue
		}

		for _, sampleOffset := range sampleOffsets {
			diff := dbHash.TimeOffset - sampleOffset

			bin := int(math.Round(diff * TimeBinResolution)) // 50 ms resolution

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

	log.Info("Recognition analysis complete",
		"bestScore", bestScore,
		"totalCandidates", len(scores),
		"sampleHashes", len(sampleHashes),
	)

	// Adaptive threshold: max(MinAbsoluteScore, sampleHashes * MinScoreRatio)
	minThreshold := max(MinAbsoluteScore, int(float64(len(sampleHashes))*MinScoreRatio))

	if bestScore < minThreshold {
		log.Info("Score below threshold",
			"bestScore", bestScore,
			"threshold", minThreshold,
		)
		return nil, nil
	}

	// Fetch song details
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
