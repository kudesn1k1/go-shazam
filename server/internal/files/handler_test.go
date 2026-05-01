package files

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go-shazam/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestJWTService() *auth.JWTService {
	return auth.NewJWTService(&auth.Config{
		AccessTokenSecret:  "test-access-secret-key-32chars!",
		RefreshTokenSecret: "test-refresh-secret-key-32chars!",
		AccessTokenTTL:     15 * time.Minute,
		RefreshTokenTTL:    7 * 24 * time.Hour,
	})
}

func newBearerToken(t *testing.T, j *auth.JWTService) string {
	pair, err := j.GenerateTokenPair(uuid.New(), []string{"user"})
	require.NoError(t, err)
	return "Bearer " + pair.AccessToken
}

func setupRouter(svc *FilesService, cfg *Config) (*gin.Engine, *auth.JWTService) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	j := newTestJWTService()
	h := NewFilesHandler(svc, cfg)
	RegisterRoutes(r, h, j)
	return r, j
}

type readCloser struct{ io.Reader }

func (readCloser) Close() error { return nil }

func TestFilesHandler_Upload_HappyPath(t *testing.T) {
	repo := new(mockRepo)
	s3c := new(mockS3)
	cfg := &Config{MaxUploadBytes: 1 << 20, Bucket: "b"}
	svc := NewFilesService(repo, s3c, cfg)
	router, j := setupRouter(svc, cfg)

	data := jpegBytes()
	repo.On("FindByHash", mock.Anything, mock.Anything).Return((*FileEntity)(nil), nil).Once()
	s3c.On("PutObject", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	repo.On("Insert", mock.Anything, mock.Anything).Return(nil).Once()

	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	part, err := mw.CreateFormFile("file", "x.jpg")
	require.NoError(t, err)
	_, _ = part.Write(data)
	require.NoError(t, mw.Close())

	req, _ := http.NewRequest(http.MethodPost, "/api/files", body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", newBearerToken(t, j))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), `"content_type":"image/jpeg"`)
	assert.Contains(t, w.Body.String(), `"hash":"`)
}

func TestFilesHandler_Upload_Unauthenticated(t *testing.T) {
	cfg := &Config{MaxUploadBytes: 1 << 20}
	router, _ := setupRouter(NewFilesService(new(mockRepo), new(mockS3), cfg), cfg)

	req, _ := http.NewRequest(http.MethodPost, "/api/files", strings.NewReader(""))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestFilesHandler_Upload_PayloadTooLarge(t *testing.T) {
	repo := new(mockRepo)
	s3c := new(mockS3)
	cfg := &Config{MaxUploadBytes: 4, Bucket: "b"}
	svc := NewFilesService(repo, s3c, cfg)
	router, j := setupRouter(svc, cfg)

	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	part, _ := mw.CreateFormFile("file", "x.jpg")
	_, _ = part.Write(jpegBytes())
	_ = mw.Close()

	req, _ := http.NewRequest(http.MethodPost, "/api/files", body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", newBearerToken(t, j))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
}

func TestFilesHandler_Get_HappyPath(t *testing.T) {
	repo := new(mockRepo)
	s3c := new(mockS3)
	cfg := &Config{Bucket: "b"}
	svc := NewFilesService(repo, s3c, cfg)
	router, j := setupRouter(svc, cfg)

	hash := strings.Repeat("a", 64)
	data := []byte("binary-bytes")
	repo.On("FindByHash", mock.Anything, hash).Return(&FileEntity{
		Hash: hash, ContentType: "image/png", SizeBytes: len(data), Status: StatusConfirmed,
	}, nil).Once()
	s3c.On("GetObject", mock.Anything, ObjectKey(hash)).Return(readCloser{bytes.NewReader(data)}, nil).Once()

	req, _ := http.NewRequest(http.MethodGet, "/api/files/"+hash, nil)
	req.Header.Set("Authorization", newBearerToken(t, j))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "image/png", w.Header().Get("Content-Type"))
	assert.Equal(t, fmt.Sprintf("\"%s\"", hash), w.Header().Get("ETag"))
	assert.Equal(t, "public, max-age=31536000, immutable", w.Header().Get("Cache-Control"))
	assert.Equal(t, `inline; filename="`+hash[:12]+`.png"`, w.Header().Get("Content-Disposition"))
	assert.Equal(t, data, w.Body.Bytes())
}

func TestFilesHandler_Get_PublicNoAuthRequired(t *testing.T) {
	repo := new(mockRepo)
	s3c := new(mockS3)
	cfg := &Config{Bucket: "b"}
	svc := NewFilesService(repo, s3c, cfg)
	router, _ := setupRouter(svc, cfg)

	hash := strings.Repeat("d", 64)
	data := []byte("public-bytes")
	repo.On("FindByHash", mock.Anything, hash).Return(&FileEntity{
		Hash: hash, ContentType: "image/png", SizeBytes: len(data), Status: StatusConfirmed,
	}, nil).Once()
	s3c.On("GetObject", mock.Anything, ObjectKey(hash)).Return(readCloser{bytes.NewReader(data)}, nil).Once()

	req, _ := http.NewRequest(http.MethodGet, "/api/files/"+hash, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, data, w.Body.Bytes())
}

func TestFilesHandler_Get_NotModified(t *testing.T) {
	repo := new(mockRepo)
	s3c := new(mockS3)
	cfg := &Config{Bucket: "b"}
	svc := NewFilesService(repo, s3c, cfg)
	router, j := setupRouter(svc, cfg)

	hash := strings.Repeat("b", 64)
	repo.On("FindByHash", mock.Anything, hash).Return(&FileEntity{
		Hash: hash, ContentType: "image/png", SizeBytes: 12, Status: StatusConfirmed,
	}, nil).Once()

	req, _ := http.NewRequest(http.MethodGet, "/api/files/"+hash, nil)
	req.Header.Set("Authorization", newBearerToken(t, j))
	req.Header.Set("If-None-Match", fmt.Sprintf("\"%s\"", hash))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotModified, w.Code)
	s3c.AssertNotCalled(t, "GetObject", mock.Anything, mock.Anything)
}

func TestFilesHandler_Get_NotFound(t *testing.T) {
	repo := new(mockRepo)
	s3c := new(mockS3)
	cfg := &Config{Bucket: "b"}
	svc := NewFilesService(repo, s3c, cfg)
	router, j := setupRouter(svc, cfg)

	hash := strings.Repeat("c", 64)
	repo.On("FindByHash", mock.Anything, hash).Return((*FileEntity)(nil), nil).Once()

	req, _ := http.NewRequest(http.MethodGet, "/api/files/"+hash, nil)
	req.Header.Set("Authorization", newBearerToken(t, j))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestFilesHandler_Get_InvalidHashFormat(t *testing.T) {
	cfg := &Config{Bucket: "b"}
	svc := NewFilesService(new(mockRepo), new(mockS3), cfg)
	router, j := setupRouter(svc, cfg)

	req, _ := http.NewRequest(http.MethodGet, "/api/files/not-a-hash", nil)
	req.Header.Set("Authorization", newBearerToken(t, j))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
