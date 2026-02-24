package song

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockSongMetadataSource struct {
	mock.Mock
}

func (m *MockSongMetadataSource) GetSongMetadata(ctx context.Context, sourceID string) (*SongMetadata, error) {
	args := m.Called(ctx, sourceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*SongMetadata), args.Error(1)
}

func (m *MockSongMetadataSource) ExtractSourceID(link string) (string, error) {
	args := m.Called(link)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.Get(0).(string), args.Error(1)
}

type MockSongDownloader struct {
	mock.Mock
}

func (m *MockSongDownloader) DownloadSong(ctx context.Context, data *SongMetadata, dir string) (*DownloadedSong, error) {
	args := m.Called(ctx, data, dir)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*DownloadedSong), args.Error(1)
}

type MockSongRepository struct {
	mock.Mock
}

func (m *MockSongRepository) Save(ctx context.Context, song *SongEntity) error {
	args := m.Called(ctx, song)
	return args.Error(0)
}

func (m *MockSongRepository) FindByID(ctx context.Context, id uuid.UUID) (*SongEntity, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*SongEntity), args.Error(1)
}

func (m *MockSongRepository) FindByTitleAndArtist(ctx context.Context, title string, artist string) (*SongEntity, error) {
	args := m.Called(ctx, title, artist)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*SongEntity), args.Error(1)
}

func (m *MockSongRepository) FindByUploadedBy(ctx context.Context, userID uuid.UUID, limit, offset int) ([]SongEntity, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]SongEntity), args.Error(1)
}

func (m *MockSongRepository) CountByUploadedBy(ctx context.Context, userID uuid.UUID) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

type MockQueueService struct {
	mock.Mock
}

func (m *MockQueueService) Enqueue(taskType string, payload []byte, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	args := m.Called(taskType, payload)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*asynq.TaskInfo), args.Error(1)
}

func (m *MockQueueService) Close() error {
	return nil
}

func TestSongService_GetSongMetadata_Success(t *testing.T) {
	mockMetadataSource := new(MockSongMetadataSource)
	mockDownloader := new(MockSongDownloader)

	expectedMetadata := &SongMetadata{
		Title:      "Test Song",
		Artist:     "Test Artist",
		DurationMs: 180000,
	}

	mockMetadataSource.On("GetSongMetadata", mock.Anything, "spotify-id").Return(expectedMetadata, nil)

	service := NewSongService(mockMetadataSource, mockDownloader, nil, nil, nil, nil)

	result, err := service.GetSongMetadata(context.Background(), "spotify-id")

	assert.NoError(t, err)
	assert.Equal(t, expectedMetadata, result)
	mockMetadataSource.AssertExpectations(t)
}

func TestSongService_GetSongMetadata_Error(t *testing.T) {
	mockMetadataSource := new(MockSongMetadataSource)
	mockDownloader := new(MockSongDownloader)

	expectedError := errors.New("failed to get metadata")
	mockMetadataSource.On("GetSongMetadata", mock.Anything, "invalid-id").Return(nil, expectedError)

	service := NewSongService(mockMetadataSource, mockDownloader, nil, nil, nil, nil)

	result, err := service.GetSongMetadata(context.Background(), "invalid-id")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedError, err)
	mockMetadataSource.AssertExpectations(t)
}

func TestSongService_EnqueueSong_Success(t *testing.T) {
	mockMetadataSource := new(MockSongMetadataSource)
	mockQueue := new(MockQueueService)

	mockMetadataSource.On("ExtractSourceID", "https://open.spotify.com/track/123").Return("123", nil)
	mockQueue.On("Enqueue", AddSongTaskType, mock.Anything).Return(nil, nil)

	service := NewSongService(mockMetadataSource, nil, nil, nil, nil, mockQueue)

	err := service.EnqueueSong(context.Background(), "https://open.spotify.com/track/123", nil)

	assert.NoError(t, err)
	mockMetadataSource.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
}

func TestSongService_EnqueueSong_WithUploadedBy(t *testing.T) {
	mockMetadataSource := new(MockSongMetadataSource)
	mockQueue := new(MockQueueService)

	userID := uuid.New()
	mockMetadataSource.On("ExtractSourceID", "https://open.spotify.com/track/456").Return("456", nil)
	mockQueue.On("Enqueue", AddSongTaskType, mock.Anything).Return(nil, nil)

	service := NewSongService(mockMetadataSource, nil, nil, nil, nil, mockQueue)

	err := service.EnqueueSong(context.Background(), "https://open.spotify.com/track/456", &userID)

	assert.NoError(t, err)
	mockMetadataSource.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
}

func TestSongService_EnqueueSong_InvalidLink(t *testing.T) {
	mockMetadataSource := new(MockSongMetadataSource)

	mockMetadataSource.On("ExtractSourceID", "invalid-link").Return("", errors.New("invalid link"))

	service := NewSongService(mockMetadataSource, nil, nil, nil, nil, nil)

	err := service.EnqueueSong(context.Background(), "invalid-link", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to extract source ID")
	mockMetadataSource.AssertExpectations(t)
}

func TestSongService_GetUserSongs_Success(t *testing.T) {
	mockRepo := new(MockSongRepository)
	userID := uuid.New()
	songID := uuid.New()

	expectedSongs := []SongEntity{
		{ID: songID, Title: "Song 1", Artist: "Artist 1", Duration: 180000, SourceID: "src1", UploadedBy: &userID},
	}

	mockRepo.On("FindByUploadedBy", mock.Anything, userID, 20, 0).Return(expectedSongs, nil)
	mockRepo.On("CountByUploadedBy", mock.Anything, userID).Return(1, nil)

	service := NewSongService(nil, nil, mockRepo, nil, nil, nil)

	songs, total, err := service.GetUserSongs(context.Background(), userID, 1, 20)

	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, songs, 1)
	assert.Equal(t, "Song 1", songs[0].Title)
	mockRepo.AssertExpectations(t)
}

func TestSongService_GetUserSongs_Empty(t *testing.T) {
	mockRepo := new(MockSongRepository)
	userID := uuid.New()

	mockRepo.On("FindByUploadedBy", mock.Anything, userID, 20, 0).Return([]SongEntity{}, nil)
	mockRepo.On("CountByUploadedBy", mock.Anything, userID).Return(0, nil)

	service := NewSongService(nil, nil, mockRepo, nil, nil, nil)

	songs, total, err := service.GetUserSongs(context.Background(), userID, 1, 20)

	assert.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, songs)
	mockRepo.AssertExpectations(t)
}
