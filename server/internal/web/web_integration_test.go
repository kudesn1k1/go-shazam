//go:build integration

package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-shazam/internal/auth"
	"go-shazam/internal/core/db"
	"go-shazam/internal/song"
	authtest "go-shazam/test/auth"
	"go-shazam/test/containers"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildTestRouter wires the routes that the integration suite exercises
// against a real DB + real auth middleware. Spotify, yt-dlp, S3 and the queue
// are not wired — the endpoints under test do not invoke them.
func buildTestRouter(t *testing.T) (*gin.Engine, *auth.JWTService) {
	t.Helper()
	pgDB := containers.StartPostgres(t)

	txm := db.NewTransactionManager(pgDB)
	repoFactory := db.NewRepository(txm)
	songRepo := song.NewSongRepository(repoFactory)

	jwtSvc := auth.NewJWTService(&auth.Config{
		AccessTokenSecret:  "test-access-secret-32characters!",
		RefreshTokenSecret: "test-refresh-secret-32character!",
		AccessTokenTTL:     15 * time.Minute,
		RefreshTokenTTL:    7 * 24 * time.Hour,
	})

	gin.SetMode(gin.TestMode)
	r := gin.New()

	songSvc := song.NewSongService(nil, nil, songRepo, nil, txm, nil)
	songHandler := song.NewSongHandler(songSvc)
	song.RegisterRoutes(r, songHandler, jwtSvc)

	return r, jwtSvc
}

func TestWeb_PostSongAdd_NoTokenReturns401_INT(t *testing.T) {
	r, _ := buildTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/song/add",
		bytes.NewBufferString(`{"link":"https://example.com/x"}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWeb_GetAdminSongs_AsUserReturns403_INT(t *testing.T) {
	r, jwt := buildTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/songs", nil)
	req.Header.Set("Authorization", authtest.MustIssueBearer(t, jwt, uuid.New(), "user"))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestWeb_GetAdminSongs_AsAdminReturns200WithPaginatedShape_INT(t *testing.T) {
	r, jwt := buildTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/songs", nil)
	req.Header.Set("Authorization", authtest.MustIssueBearer(t, jwt, uuid.New(), "user", "admin"))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Data  []any `json:"data"`
		Total int   `json:"total"`
		Page  int   `json:"page"`
		Limit int   `json:"limit"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 0, resp.Total)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 20, resp.Limit)
	assert.NotNil(t, resp.Data)
}

func TestWeb_GetPublicSongs_DoesNotLeakUploaderEmail_INT(t *testing.T) {
	r, _ := buildTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/public/songs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.NotContains(t, body, "uploader_email")
	assert.NotContains(t, body, "uploaded_by")
}
