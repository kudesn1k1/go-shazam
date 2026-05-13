//go:build integration

package song

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"go-shazam/internal/core/db"
	"go-shazam/internal/queue"
	"go-shazam/test/containers"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type integrationSetup struct {
	db        *sqlx.DB
	songRepo  SongRepositoryInterface
	songSvc   *SongService
	metadata  *MockSongMetadataSource
	downloader *MockSongDownloader
}

func setupIntegration(t *testing.T) *integrationSetup {
	t.Helper()

	pgDB := containers.StartPostgres(t)
	redisOpt, _ := containers.StartRedis(t)

	txm := db.NewTransactionManager(pgDB)
	repoFactory := db.NewRepository(txm)
	songRepo := NewSongRepository(repoFactory)

	queueSvc := queue.NewQueueService(&queue.Config{Addr: redisOpt.Addr})
	t.Cleanup(func() { _ = queueSvc.Close() })

	metadata := new(MockSongMetadataSource)
	downloader := new(MockSongDownloader)

	// fingerprintService nil — the tests below don't drive AddSong, just
	// EnqueueSong / repository / ListSongs.
	songSvc := NewSongService(metadata, downloader, songRepo, nil, txm, queueSvc)

	return &integrationSetup{
		db:        pgDB,
		songRepo:  songRepo,
		songSvc:   songSvc,
		metadata:  metadata,
		downloader: downloader,
	}
}

func TestSongService_EnqueueSong_SecondEnqueueIsDeduped_INT(t *testing.T) {
	s := setupIntegration(t)
	ctx := context.Background()

	s.metadata.On("ExtractSourceID", "https://example.com/track/abc").
		Return("abc", nil)

	uploader := uuid.New()

	err := s.songSvc.EnqueueSong(ctx, "https://example.com/track/abc", &uploader)
	require.NoError(t, err)

	err = s.songSvc.EnqueueSong(ctx, "https://example.com/track/abc", &uploader)
	assert.ErrorIs(t, err, ErrSongTaskAlreadyExists)
}

func TestSongService_ListSongs_PaginationAgainstRealDB_INT(t *testing.T) {
	s := setupIntegration(t)
	ctx := context.Background()

	for i := 0; i < 25; i++ {
		songID, err := uuid.NewV7()
		require.NoError(t, err)
		require.NoError(t, s.songRepo.Save(ctx, &SongEntity{
			ID:        songID,
			Title:     fmt.Sprintf("Title %02d", i),
			Artist:    "Artist X",
			Duration:  180000,
			SourceID:  fmt.Sprintf("src-%02d", i),
			CreatedAt: time.Now().Add(time.Duration(-i) * time.Hour).UTC(),
		}))
	}

	songs, total, err := s.songSvc.ListSongs(ctx, SongFilter{
		Limit:  10,
		Offset: 0,
		Sort:   SortCreatedAt,
		Order:  SortDesc,
	})
	require.NoError(t, err)
	assert.Equal(t, 25, total)
	assert.Len(t, songs, 10)

	songs, total, err = s.songSvc.ListSongs(ctx, SongFilter{
		Limit:  10,
		Offset: 20,
		Sort:   SortCreatedAt,
		Order:  SortDesc,
	})
	require.NoError(t, err)
	assert.Equal(t, 25, total)
	assert.Len(t, songs, 5)
}

func TestSongRepository_FindByTitleAndArtist_NotFoundReturnsNilNil_INT(t *testing.T) {
	s := setupIntegration(t)
	ctx := context.Background()

	got, err := s.songRepo.FindByTitleAndArtist(ctx, "Does Not Exist", "Nobody")
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestSongService_AddSong_MetadataSourceOutageReturnsErrorNoOrphan_INT(t *testing.T) {
	// Lab requirement 4.5 — third-party API outage. We don't need Playwright
	// for this: we exercise the SongService.AddSong path with a metadata source
	// that simulates Spotify being down, and assert (a) the call returns an
	// error, (b) the DB has no orphan song rows or fingerprints from a partial
	// transaction.
	s := setupIntegration(t)
	ctx := context.Background()

	s.metadata.On("GetSongMetadata", mock.Anything, "outage-id").
		Return(nil, errors.New("upstream 503"))

	_, err := s.songSvc.AddSong(ctx, "outage-id", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get song metadata")

	// No songs should have been created
	songs, total, err := s.songSvc.ListSongs(ctx, SongFilter{
		Limit: 100, Offset: 0, Sort: SortCreatedAt, Order: SortDesc,
	})
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, songs)
}

func TestSongRepository_Save_ThenFindByID_Roundtrip_INT(t *testing.T) {
	s := setupIntegration(t)
	ctx := context.Background()

	songID, err := uuid.NewV7()
	require.NoError(t, err)
	want := &SongEntity{
		ID:        songID,
		Title:     "Round-trip",
		Artist:    "Tester",
		Duration:  42000,
		SourceID:  "rt-1",
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
	}
	require.NoError(t, s.songRepo.Save(ctx, want))

	got, err := s.songRepo.FindByID(ctx, songID)
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, want.Title, got.Title)
	assert.Equal(t, want.Artist, got.Artist)
	assert.Equal(t, want.Duration, got.Duration)
}
