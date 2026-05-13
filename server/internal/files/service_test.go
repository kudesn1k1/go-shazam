package files

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockRepo struct{ mock.Mock }

func (m *mockRepo) FindByHash(ctx context.Context, hash string) (*FileEntity, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*FileEntity), args.Error(1)
}
func (m *mockRepo) Insert(ctx context.Context, e *FileEntity) error {
	return m.Called(ctx, e).Error(0)
}
func (m *mockRepo) BumpUpdatedAt(ctx context.Context, hash string) error {
	return m.Called(ctx, hash).Error(0)
}
func (m *mockRepo) UpdateStatus(ctx context.Context, hash, status string) error {
	return m.Called(ctx, hash, status).Error(0)
}
func (m *mockRepo) IsReferenced(ctx context.Context, hash string) (bool, error) {
	args := m.Called(ctx, hash)
	return args.Bool(0), args.Error(1)
}
func (m *mockRepo) ListExpired(ctx context.Context, olderThan time.Time, limit int) ([]string, error) {
	args := m.Called(ctx, olderThan, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}
func (m *mockRepo) DeleteByHash(ctx context.Context, hash string) error {
	return m.Called(ctx, hash).Error(0)
}

type mockS3 struct{ mock.Mock }

func (m *mockS3) PutObject(ctx context.Context, key string, body io.Reader, contentType string, size int64) error {
	buf, _ := io.ReadAll(body)
	return m.Called(ctx, key, buf, contentType, size).Error(0)
}
func (m *mockS3) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}
func (m *mockS3) DeleteObject(ctx context.Context, key string) error {
	return m.Called(ctx, key).Error(0)
}
func (m *mockS3) ObjectExists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

// Minimal SOI + APP0 prefix that http.DetectContentType recognises as image/jpeg.
func jpegBytes() []byte {
	return []byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00}
}

func TestFilesService_Upload_FreshInsert(t *testing.T) {
	repo := new(mockRepo)
	s3c := new(mockS3)
	svc := NewFilesService(repo, s3c, &Config{MaxUploadBytes: 1024, Bucket: "b"})

	data := jpegBytes()
	repo.On("FindByHash", mock.Anything, mock.AnythingOfType("string")).Return((*FileEntity)(nil), nil).Once()
	s3c.On("PutObject", mock.Anything, mock.AnythingOfType("string"), data, "image/jpeg", int64(len(data))).Return(nil).Once()
	repo.On("Insert", mock.Anything, mock.AnythingOfType("*files.FileEntity")).Return(nil).Once()

	resp, err := svc.Upload(context.Background(), bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)
	assert.Equal(t, "image/jpeg", resp.ContentType)
	assert.Equal(t, len(data), resp.SizeBytes)
	assert.Len(t, resp.Hash, 64)
	repo.AssertExpectations(t)
	s3c.AssertExpectations(t)
}

func TestFilesService_Upload_DedupHitTemporaryBumpsUpdatedAt(t *testing.T) {
	repo := new(mockRepo)
	s3c := new(mockS3)
	svc := NewFilesService(repo, s3c, &Config{MaxUploadBytes: 1024, Bucket: "b"})

	data := jpegBytes()
	existing := &FileEntity{Status: StatusTemporary, ContentType: "image/jpeg", SizeBytes: len(data)}
	repo.On("FindByHash", mock.Anything, mock.Anything).Return(existing, nil).Once()
	repo.On("BumpUpdatedAt", mock.Anything, mock.Anything).Return(nil).Once()
	s3c.On("ObjectExists", mock.Anything, mock.Anything).Return(true, nil).Once()

	resp, err := svc.Upload(context.Background(), bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)
	assert.Len(t, resp.Hash, 64)
	assert.Equal(t, "image/jpeg", resp.ContentType)
	assert.Equal(t, len(data), resp.SizeBytes)
	s3c.AssertNotCalled(t, "PutObject", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	repo.AssertExpectations(t)
	s3c.AssertExpectations(t)
}

func TestFilesService_Upload_DedupHitConfirmedNoBump(t *testing.T) {
	repo := new(mockRepo)
	s3c := new(mockS3)
	svc := NewFilesService(repo, s3c, &Config{MaxUploadBytes: 1024, Bucket: "b"})

	data := jpegBytes()
	existing := &FileEntity{Hash: "x", Status: StatusConfirmed}
	repo.On("FindByHash", mock.Anything, mock.Anything).Return(existing, nil).Once()
	s3c.On("ObjectExists", mock.Anything, mock.Anything).Return(true, nil).Once()

	_, err := svc.Upload(context.Background(), bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)
	s3c.AssertNotCalled(t, "PutObject", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "BumpUpdatedAt", mock.Anything, mock.Anything)
}

func TestFilesService_Upload_DedupHitObjectMissingReuploads(t *testing.T) {
	repo := new(mockRepo)
	s3c := new(mockS3)
	svc := NewFilesService(repo, s3c, &Config{MaxUploadBytes: 1024, Bucket: "b"})

	data := jpegBytes()
	existing := &FileEntity{Status: StatusConfirmed, ContentType: "image/jpeg", SizeBytes: len(data)}
	repo.On("FindByHash", mock.Anything, mock.Anything).Return(existing, nil).Once()
	s3c.On("ObjectExists", mock.Anything, mock.Anything).Return(false, nil).Once()
	s3c.On("PutObject", mock.Anything, mock.Anything, data, "image/jpeg", int64(len(data))).Return(nil).Once()

	resp, err := svc.Upload(context.Background(), bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)
	assert.Len(t, resp.Hash, 64)
	repo.AssertNotCalled(t, "Insert", mock.Anything, mock.Anything)
	repo.AssertExpectations(t)
	s3c.AssertExpectations(t)
}

func TestFilesService_Upload_InsertFailureRollsBackS3(t *testing.T) {
	repo := new(mockRepo)
	s3c := new(mockS3)
	svc := NewFilesService(repo, s3c, &Config{MaxUploadBytes: 1024, Bucket: "b"})

	data := jpegBytes()
	repo.On("FindByHash", mock.Anything, mock.Anything).Return((*FileEntity)(nil), nil).Once()
	s3c.On("PutObject", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	repo.On("Insert", mock.Anything, mock.Anything).Return(errors.New("db down")).Once()
	s3c.On("DeleteObject", mock.Anything, mock.Anything).Return(nil).Once()

	_, err := svc.Upload(context.Background(), bytes.NewReader(data), int64(len(data)))
	require.Error(t, err)
	repo.AssertExpectations(t)
	s3c.AssertExpectations(t)
}

func TestFilesService_Upload_RejectsUnsupportedMIME(t *testing.T) {
	repo := new(mockRepo)
	s3c := new(mockS3)
	svc := NewFilesService(repo, s3c, &Config{MaxUploadBytes: 1024, Bucket: "b"})

	data := []byte("just some plain text, definitely not an image")
	_, err := svc.Upload(context.Background(), bytes.NewReader(data), int64(len(data)))
	require.ErrorIs(t, err, ErrUnsupportedMIME)
}

func TestFilesService_Upload_RejectsOversize(t *testing.T) {
	repo := new(mockRepo)
	s3c := new(mockS3)
	svc := NewFilesService(repo, s3c, &Config{MaxUploadBytes: 5, Bucket: "b"})

	data := jpegBytes()
	_, err := svc.Upload(context.Background(), bytes.NewReader(data), int64(len(data)))
	require.ErrorIs(t, err, ErrFileTooLarge)
}

func TestFilesService_Upload_S3FailureSkipsInsert(t *testing.T) {
	repo := new(mockRepo)
	s3c := new(mockS3)
	svc := NewFilesService(repo, s3c, &Config{MaxUploadBytes: 1024, Bucket: "b"})

	data := jpegBytes()
	repo.On("FindByHash", mock.Anything, mock.Anything).Return((*FileEntity)(nil), nil).Once()
	s3c.On("PutObject", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("boom")).Once()

	_, err := svc.Upload(context.Background(), bytes.NewReader(data), int64(len(data)))
	require.Error(t, err)
	repo.AssertNotCalled(t, "Insert", mock.Anything, mock.Anything)
}

func TestFilesService_Dereference_LastReference(t *testing.T) {
	repo := new(mockRepo)
	svc := NewFilesService(repo, new(mockS3), &Config{})
	repo.On("IsReferenced", mock.Anything, "h").Return(false, nil).Once()
	repo.On("UpdateStatus", mock.Anything, "h", StatusTemporary).Return(nil).Once()

	require.NoError(t, svc.Dereference(context.Background(), "h"))
	repo.AssertExpectations(t)
}

func TestFilesService_Dereference_OthersStillReference(t *testing.T) {
	repo := new(mockRepo)
	svc := NewFilesService(repo, new(mockS3), &Config{})
	repo.On("IsReferenced", mock.Anything, "h").Return(true, nil).Once()

	require.NoError(t, svc.Dereference(context.Background(), "h"))
	repo.AssertNotCalled(t, "UpdateStatus", mock.Anything, mock.Anything, mock.Anything)
}

func TestFilesService_Confirm(t *testing.T) {
	repo := new(mockRepo)
	svc := NewFilesService(repo, new(mockS3), &Config{})
	repo.On("UpdateStatus", mock.Anything, "h", StatusConfirmed).Return(nil).Once()

	require.NoError(t, svc.Confirm(context.Background(), "h"))
	repo.AssertExpectations(t)
}
