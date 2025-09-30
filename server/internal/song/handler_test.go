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

func TestSongHandler_Get_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	NewSongHandler(router, nil)

	req, _ := http.NewRequest(http.MethodPost, "/api/song/add", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestSongHandler_Get_MissingLink(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	NewSongHandler(router, nil)

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
