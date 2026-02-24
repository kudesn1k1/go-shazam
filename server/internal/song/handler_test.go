package song

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-shazam/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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
