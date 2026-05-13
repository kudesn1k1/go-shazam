package files

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var (
	ErrUnsupportedMIME = errors.New("unsupported file type")
	ErrFileTooLarge    = errors.New("file too large")
)

var allowedContentTypes = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/webp": {},
}

type FilesService struct {
	repo FilesRepositoryInterface
	s3   S3Client
	cfg  *Config
}

func NewFilesService(repo FilesRepositoryInterface, s3 S3Client, cfg *Config) *FilesService {
	return &FilesService{repo: repo, s3: s3, cfg: cfg}
}

// Upload reads body fully (bounded by size), hashes, detects MIME, dedups against
// existing rows, and uploads to S3 if the content is new.
//
// Atomicity:
//   - Dedup hit: if the DB row exists but the S3 object is missing (e.g., a prior
//     upload rolled back the wrong way), re-upload the bytes to self-heal.
//   - Fresh upload: PutObject first, then Insert. On Insert failure, best-effort
//     DeleteObject to avoid leaving an unreferenced S3 orphan.
//
// Because storage keys are content-addressed (sha256), any residual orphans in S3
// get harmlessly overwritten on the next upload of the same bytes.
func (s *FilesService) Upload(ctx context.Context, body io.Reader, declaredSize int64) (*UploadResponse, error) {
	if declaredSize > s.cfg.MaxUploadBytes {
		return nil, ErrFileTooLarge
	}

	buf := &bytes.Buffer{}
	n, err := io.Copy(buf, io.LimitReader(body, s.cfg.MaxUploadBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	if n > s.cfg.MaxUploadBytes {
		return nil, ErrFileTooLarge
	}

	data := buf.Bytes()

	sniffLen := 512
	if len(data) < sniffLen {
		sniffLen = len(data)
	}
	contentType := http.DetectContentType(data[:sniffLen])
	if _, ok := allowedContentTypes[contentType]; !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedMIME, contentType)
	}

	sum := sha256.Sum256(data)
	hash := hex.EncodeToString(sum[:])
	key := ObjectKey(hash)

	existing, err := s.repo.FindByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("find by hash: %w", err)
	}
	if existing != nil {
		if existing.Status == StatusTemporary {
			if err := s.repo.BumpUpdatedAt(ctx, hash); err != nil {
				return nil, fmt.Errorf("bump updated_at: %w", err)
			}
		}
		// Self-heal: if the DB row exists but the S3 object doesn't (orphan row from a
		// prior failed state), push the bytes again. Content is identical by hash, so
		// this is always safe.
		exists, err := s.s3.ObjectExists(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("check object: %w", err)
		}
		if !exists {
			if err := s.s3.PutObject(ctx, key, bytes.NewReader(data), contentType, int64(len(data))); err != nil {
				return nil, fmt.Errorf("s3 put (self-heal): %w", err)
			}
		}
		return &UploadResponse{Hash: hash, ContentType: existing.ContentType, SizeBytes: existing.SizeBytes}, nil
	}

	if err := s.s3.PutObject(ctx, key, bytes.NewReader(data), contentType, int64(len(data))); err != nil {
		return nil, fmt.Errorf("s3 put: %w", err)
	}

	now := time.Now().UTC()
	entity := &FileEntity{
		Hash:        hash,
		ContentType: contentType,
		SizeBytes:   len(data),
		Status:      StatusTemporary,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.Insert(ctx, entity); err != nil {
		// Best-effort rollback: delete the object we just uploaded so we don't
		// leak an S3 orphan with no tracking row.
		if delErr := s.s3.DeleteObject(ctx, key); delErr != nil {
			return nil, fmt.Errorf("insert: %w (s3 rollback also failed: %v)", err, delErr)
		}
		return nil, fmt.Errorf("insert: %w", err)
	}

	return &UploadResponse{Hash: hash, ContentType: contentType, SizeBytes: len(data)}, nil
}

func (s *FilesService) GetByHash(ctx context.Context, hash string) (*FileEntity, error) {
	return s.repo.FindByHash(ctx, hash)
}

func (s *FilesService) OpenObject(ctx context.Context, hash string) (io.ReadCloser, error) {
	return s.s3.GetObject(ctx, ObjectKey(hash))
}

func (s *FilesService) Confirm(ctx context.Context, hash string) error {
	return s.repo.UpdateStatus(ctx, hash, StatusConfirmed)
}

// Dereference: if no users still reference the hash, transition it back to TEMPORARY
// so the cleanup job can reap it after the grace window.
func (s *FilesService) Dereference(ctx context.Context, hash string) error {
	referenced, err := s.repo.IsReferenced(ctx, hash)
	if err != nil {
		return fmt.Errorf("is referenced: %w", err)
	}
	if referenced {
		return nil
	}
	return s.repo.UpdateStatus(ctx, hash, StatusTemporary)
}
