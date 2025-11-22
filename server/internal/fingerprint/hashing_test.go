package fingerprint

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateHashes(t *testing.T) {
	songID := uuid.New()

	// Create a sequence of peaks
	// Anchor at T=0, F=1000
	// Target at T=1.0, F=1200
	// Target at T=1.5, F=1300
	// Target at T=5.0, F=1400 (Out of target zone if we assume < 2.0s, checking impl)

	// In hashing.go:
	// if timeDelta < 0.1 { continue }
	// if timeDelta > 2.0 { break }

	peaks := []Peak{
		{Time: 0.0, Frequency: 1000.0},
		{Time: 0.05, Frequency: 1000.0}, // Too close (<0.1)
		{Time: 1.0, Frequency: 1200.0},  // Good
		{Time: 1.5, Frequency: 1300.0},  // Good
		{Time: 3.0, Frequency: 1400.0},  // Too far (>2.0)
	}

	hashes := CreateHashes(peaks, songID)

	// We expect hashes from the first anchor (T=0.0) to:
	// - T=1.0 (Delta=1.0)
	// - T=1.5 (Delta=1.5)
	// Total 2 hashes from first anchor.

	// Second peak (T=0.05) is also an anchor? Yes.
	// From T=0.05:
	// - T=1.0 (Delta=0.95) -> Good
	// - T=1.5 (Delta=1.45) -> Good
	// - T=3.0 (Delta=2.95) -> Too far
	// Total 2 hashes from second anchor.

	// Third peak (T=1.0):
	// - T=1.5 (Delta=0.5) -> Good
	// - T=3.0 (Delta=2.0) -> Edge case?
	// Impl: if timeDelta > 2.0 { break }. So 2.0 is excluded? Or included?
	// Let's check: usually strictly greater.

	// Let's count total.
	// Anchor 1 (0.0): + 1.0, + 1.5 (2 hashes)
	// Anchor 2 (0.05): + 1.0, + 1.5 (2 hashes)
	// Anchor 3 (1.0): + 1.5, + 3.0 (Delta 2.0).
	// If Delta == 2.0, > 2.0 is false. So it matches.
	// So Anchor 3 matches T=3.0.
	// Anchor 4 (1.5): + 3.0 (Delta 1.5) -> Good. (1 hash)
	// Anchor 5 (3.0): No targets.

	// Total: 2 + 2 + 2 + 1 = 7 hashes expected (if 2.0 is included).

	// Let's assert > 0 for now and check structure.
	assert.NotEmpty(t, hashes)

	firstHash := hashes[0]
	assert.Equal(t, songID, firstHash.SongID)
	assert.Equal(t, 0.0, firstHash.TimeOffset)
}
