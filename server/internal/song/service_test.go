package song

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockSongMetadataSource struct {
	mock.Mock
}

func (m *MockSongMetadataSource) GetSongsMetadata(ctx context.Context, link string) (*SongMetadata, error) {
	args := m.Called(ctx, link)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*SongMetadata), args.Error(1)
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

func TestSongService_GetSongsMetadata_Success(t *testing.T) {
	mockMetadataSource := new(MockSongMetadataSource)

	expectedMetadata := &SongMetadata{
		Title:      "Test Song",
		Artist:     "Test Artist",
		DurationMs: 180000,
	}

	mockDownloader := new(MockSongDownloader)

	mockMetadataSource.On("GetSongsMetadata", mock.Anything, "spotify-link").Return(expectedMetadata, nil)

	service := NewSongService(mockMetadataSource, mockDownloader)

	result, err := service.GetSongsMetadata(context.Background(), "spotify-link")

	assert.NoError(t, err)
	assert.Equal(t, expectedMetadata, result)
	mockMetadataSource.AssertExpectations(t)
}

func TestSongService_GetSongsMetadata_MetadataSourceError(t *testing.T) {
	mockMetadataSource := new(MockSongMetadataSource)
	mockDownloader := new(MockSongDownloader)

	expectedError := errors.New("failed to get metadata")
	mockMetadataSource.On("GetSongsMetadata", mock.Anything, "invalid-link").Return(nil, expectedError)

	service := NewSongService(mockMetadataSource, mockDownloader)

	result, err := service.GetSongsMetadata(context.Background(), "invalid-link")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedError, err)
	mockMetadataSource.AssertExpectations(t)
	mockDownloader.AssertNotCalled(t, "DownloadSong")
}
