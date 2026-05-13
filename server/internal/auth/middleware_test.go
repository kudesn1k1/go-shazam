package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAuthRouter(t *testing.T, handler gin.HandlerFunc) (*gin.Engine, *JWTService) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	svc := NewJWTService(newTestConfig())
	r := gin.New()
	r.GET("/probe", AuthMiddleware(svc), handler)
	return r, svc
}

func TestAuthMiddleware_MissingHeaderReturns401(t *testing.T) {
	r, _ := setupAuthRouter(t, func(c *gin.Context) { c.Status(200) })
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_MalformedHeaderReturns401(t *testing.T) {
	r, _ := setupAuthRouter(t, func(c *gin.Context) { c.Status(200) })
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.Header.Set("Authorization", "Basic xyz") // not Bearer
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ExpiredTokenReturns401(t *testing.T) {
	cfg := newTestConfig()
	cfg.AccessTokenTTL = -1 * time.Minute // expired
	svc := NewJWTService(cfg)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/probe", AuthMiddleware(svc), func(c *gin.Context) { c.Status(200) })

	pair, err := svc.GenerateTokenPair(uuid.New(), []string{"user"})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ValidTokenPopulatesContext(t *testing.T) {
	userID := uuid.New()
	var capturedID uuid.UUID
	var capturedRoles []string
	r, svc := setupAuthRouter(t, func(c *gin.Context) {
		capturedID, _ = GetUserIDFromContext(c.Request.Context())
		capturedRoles = GetRolesFromContext(c.Request.Context())
		c.Status(200)
	})

	pair, err := svc.GenerateTokenPair(userID, []string{"user", "admin"})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, userID, capturedID)
	assert.ElementsMatch(t, []string{"user", "admin"}, capturedRoles)
}

func TestRoleMiddleware_AdminRouteRejectsNonAdminWith403(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := NewJWTService(newTestConfig())
	r := gin.New()
	r.GET("/admin", AuthMiddleware(svc), RoleMiddleware("admin"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	pair, err := svc.GenerateTokenPair(uuid.New(), []string{"user"})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRoleMiddleware_AdminRouteAllowsAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := NewJWTService(newTestConfig())
	r := gin.New()
	r.GET("/admin", AuthMiddleware(svc), RoleMiddleware("admin"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	pair, err := svc.GenerateTokenPair(uuid.New(), []string{"user", "admin"})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRoleMiddleware_AcceptsAnyOfMultipleRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := NewJWTService(newTestConfig())
	r := gin.New()
	r.GET("/multi", AuthMiddleware(svc), RoleMiddleware("admin", "moderator"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	pair, err := svc.GenerateTokenPair(uuid.New(), []string{"moderator"})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/multi", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
