package song

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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

func newTestBearerToken(jwtService *auth.JWTService) string {
	userID := uuid.New()
	pair, err := jwtService.GenerateTokenPair(userID, []string{"user"})
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("Bearer %s", pair.AccessToken)
}

func setupTestRouter(handler *SongHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	jwtService := newTestJWTService()
	RegisterRoutes(router, handler, jwtService)
	return router
}

func setupTestRouterWithToken(handler *SongHandler) (*gin.Engine, string) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	jwtService := newTestJWTService()
	RegisterRoutes(router, handler, jwtService)
	token := newTestBearerToken(jwtService)
	return router, token
}

func TestSongHandler_Add_Unauthorized(t *testing.T) {
	handler := NewSongHandler(nil)
	router := setupTestRouter(handler)

	req, _ := http.NewRequest(http.MethodPost, "/api/song/add", bytes.NewBufferString(`{"link":"https://open.spotify.com/track/123"}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSongHandler_Add_InvalidJSON(t *testing.T) {
	handler := NewSongHandler(nil)
	router, token := setupTestRouterWithToken(handler)

	req, _ := http.NewRequest(http.MethodPost, "/api/song/add", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestSongHandler_Add_MissingLink(t *testing.T) {
	handler := NewSongHandler(nil)
	router, token := setupTestRouterWithToken(handler)

	requestBody := map[string]interface{}{
		"other_field": "value",
	}
	jsonBody, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest(http.MethodPost, "/api/song/add", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func newServiceWithRepo(repo SongRepositoryInterface) *SongService {
	return NewSongService(nil, nil, repo, nil, nil, nil)
}

func TestSongHandler_GetPublicSongs_NoAuthRequired_RedactsUploader(t *testing.T) {
	uploader := uuid.New()
	songID := uuid.New()
	repo := new(MockSongRepository)
	repo.On("FindFiltered", mock.Anything, mock.Anything).Return([]SongEntity{
		{ID: songID, Title: "T", Artist: "A", Duration: 1234, SourceID: "src", UploadedBy: &uploader, CreatedAt: time.Now().UTC()},
	}, nil).Once()
	repo.On("CountFiltered", mock.Anything, mock.Anything).Return(1, nil).Once()

	handler := NewSongHandler(newServiceWithRepo(repo))
	router := setupTestRouter(handler)

	req, _ := http.NewRequest(http.MethodGet, "/api/public/songs", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, `"title":"T"`)
	assert.Contains(t, body, `"artist":"A"`)
	assert.False(t, strings.Contains(body, "uploaded_by"), "public response must not include uploaded_by, got: %s", body)
	assert.False(t, strings.Contains(body, uploader.String()), "public response must not leak uploader UUID, got: %s", body)
}

func TestSongHandler_GetPublicSong_NoAuthRequired_RedactsUploader(t *testing.T) {
	uploader := uuid.New()
	songID := uuid.New()
	repo := new(MockSongRepository)
	repo.On("FindByID", mock.Anything, songID).Return(&SongEntity{
		ID: songID, Title: "T", Artist: "A", Duration: 1234, SourceID: "src", UploadedBy: &uploader, CreatedAt: time.Now().UTC(),
	}, nil).Once()

	handler := NewSongHandler(newServiceWithRepo(repo))
	router := setupTestRouter(handler)

	req, _ := http.NewRequest(http.MethodGet, "/api/public/songs/"+songID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, `"title":"T"`)
	assert.False(t, strings.Contains(body, "uploaded_by"), "public response must not include uploaded_by")
	assert.False(t, strings.Contains(body, uploader.String()), "public response must not leak uploader UUID")
}

func TestSongHandler_GetPublicSong_NotFound(t *testing.T) {
	songID := uuid.New()
	repo := new(MockSongRepository)
	repo.On("FindByID", mock.Anything, songID).Return((*SongEntity)(nil), nil).Once()

	handler := NewSongHandler(newServiceWithRepo(repo))
	router := setupTestRouter(handler)

	req, _ := http.NewRequest(http.MethodGet, "/api/public/songs/"+songID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSongHandler_GetPublicSong_InvalidUUID(t *testing.T) {
	handler := NewSongHandler(newServiceWithRepo(new(MockSongRepository)))
	router := setupTestRouter(handler)

	req, _ := http.NewRequest(http.MethodGet, "/api/public/songs/not-a-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSongHandler_GetPublicSong_RepositoryError(t *testing.T) {
	songID := uuid.New()
	repo := new(MockSongRepository)
	repo.On("FindByID", mock.Anything, songID).Return((*SongEntity)(nil), errors.New("db is down")).Once()

	handler := NewSongHandler(newServiceWithRepo(repo))
	router := setupTestRouter(handler)

	req, _ := http.NewRequest(http.MethodGet, "/api/public/songs/"+songID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
