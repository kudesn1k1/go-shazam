package fingerprint

import (
	"context"
	"go-shazam/internal/audio"

	"github.com/google/uuid"
)

type FingerprintService struct {
	repo *Repository
}

func NewFingerprintService(repo *Repository) *FingerprintService {
	return &FingerprintService{repo: repo}
}

// CreateFingerprints calculates fingerprints (peaks and hashes) from audio fragments.
// This method is CPU-bound and should be called before starting a database transaction if possible.
func (s *FingerprintService) CreateFingerprints(fragments []audio.ProcessedFragment, songID uuid.UUID, sampleRate int) []Hash {
	peaks := ExtractPeaks(fragments, sampleRate)
	return CreateHashes(peaks, songID)
}

// SaveFingerprints saves the pre-calculated hashes to the database.
func (s *FingerprintService) SaveFingerprints(ctx context.Context, hashes []Hash) error {
	return s.repo.SaveFingerprints(ctx, hashes)
}

// GetMatchingHashes finds all hashes in the database that match the given sample hashes.
func (s *FingerprintService) GetMatchingHashes(ctx context.Context, sampleHashes []Hash) ([]Hash, error) {
	hashValues := make([]int64, len(sampleHashes))
	for i, h := range sampleHashes {
		hashValues[i] = h.HashValue
	}
	return s.repo.FindHashesByValues(ctx, hashValues)
}
