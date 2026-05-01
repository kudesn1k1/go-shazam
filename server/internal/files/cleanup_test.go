package files

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// newTestCleanupHandler builds a handler whose txRunner passes straight through
// (no real DB). That lets us exercise the ordering/rollback logic without
// spinning up a postgres for unit tests.
func newTestCleanupHandler(repo FilesRepositoryInterface, s3 S3Client, cfg *Config) *CleanupTaskHandler {
	return &CleanupTaskHandler{
		repo: repo,
		s3:   s3,
		cfg:  cfg,
		runTx: func(ctx context.Context, fn func(ctx context.Context) error) error {
			return fn(ctx)
		},
	}
}

func TestCleanupTaskHandler_DeletesExpired_DBBeforeS3(t *testing.T) {
	repo := new(mockRepo)
	s3c := new(mockS3)
	h := newTestCleanupHandler(repo, s3c, &Config{CleanupOutdate: 24 * time.Hour, Bucket: "b"})

	var order []string

	repo.On("ListExpired", mock.Anything, mock.AnythingOfType("time.Time"), 500).
		Return([]string{"a"}, nil).Once()
	repo.On("DeleteByHash", mock.Anything, "a").Run(func(args mock.Arguments) {
		order = append(order, "db")
	}).Return(nil).Once()
	s3c.On("DeleteObject", mock.Anything, "files/a").Run(func(args mock.Arguments) {
		order = append(order, "s3")
	}).Return(nil).Once()

	require.NoError(t, h.ProcessTask(context.Background(), asynq.NewTask(FileCleanupTaskType, nil)))
	assert.Equal(t, []string{"db", "s3"}, order)
	repo.AssertExpectations(t)
	s3c.AssertExpectations(t)
}

func TestCleanupTaskHandler_S3FailureRollsBack(t *testing.T) {
	repo := new(mockRepo)
	s3c := new(mockS3)
	// Capture what runTx saw so we can assert the tx body returned an error
	// (simulating a real transaction rollback).
	var txErr error
	h := &CleanupTaskHandler{
		repo: repo, s3: s3c, cfg: &Config{CleanupOutdate: 24 * time.Hour, Bucket: "b"},
		runTx: func(ctx context.Context, fn func(ctx context.Context) error) error {
			txErr = fn(ctx)
			return txErr
		},
	}

	repo.On("ListExpired", mock.Anything, mock.Anything, 500).Return([]string{"a"}, nil).Once()
	repo.On("DeleteByHash", mock.Anything, "a").Return(nil).Once()
	s3c.On("DeleteObject", mock.Anything, "files/a").Return(errors.New("boom")).Once()

	// ProcessTask does not propagate per-hash errors (by design: logs and continues).
	require.NoError(t, h.ProcessTask(context.Background(), asynq.NewTask(FileCleanupTaskType, nil)))
	// But the tx body DID return an error, meaning a real *sqlx.Tx would have rolled back.
	require.Error(t, txErr)
	repo.AssertExpectations(t)
	s3c.AssertExpectations(t)
}

func TestCleanupTaskHandler_DBFailureSkipsS3(t *testing.T) {
	repo := new(mockRepo)
	s3c := new(mockS3)
	h := newTestCleanupHandler(repo, s3c, &Config{CleanupOutdate: 24 * time.Hour, Bucket: "b"})

	repo.On("ListExpired", mock.Anything, mock.Anything, 500).Return([]string{"a"}, nil).Once()
	repo.On("DeleteByHash", mock.Anything, "a").Return(errors.New("db down")).Once()

	require.NoError(t, h.ProcessTask(context.Background(), asynq.NewTask(FileCleanupTaskType, nil)))
	s3c.AssertNotCalled(t, "DeleteObject", mock.Anything, mock.Anything)
	repo.AssertExpectations(t)
}

func TestCleanupTaskHandler_EmptyRun(t *testing.T) {
	repo := new(mockRepo)
	s3c := new(mockS3)
	h := newTestCleanupHandler(repo, s3c, &Config{CleanupOutdate: 24 * time.Hour})

	repo.On("ListExpired", mock.Anything, mock.Anything, 500).Return([]string{}, nil).Once()

	require.NoError(t, h.ProcessTask(context.Background(), asynq.NewTask(FileCleanupTaskType, nil)))
	s3c.AssertNotCalled(t, "DeleteObject", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "DeleteByHash", mock.Anything, mock.Anything)
}
