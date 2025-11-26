package song

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter(handler *SongHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRoutes(router, handler)
	return router
}

func TestSongHandler_Add_InvalidJSON(t *testing.T) {
	handler := NewSongHandler(nil)
	router := setupTestRouter(handler)

	req, _ := http.NewRequest(http.MethodPost, "/api/song/add", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestSongHandler_Add_MissingLink(t *testing.T) {
	handler := NewSongHandler(nil)
	router := setupTestRouter(handler)

	requestBody := map[string]interface{}{
		"other_field": "value",
	}
	jsonBody, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest(http.MethodPost, "/api/song/add", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}
