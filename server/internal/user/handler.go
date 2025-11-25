package user

import (
	"errors"
	"net/http"

	"go-shazam/internal/auth"
	"go-shazam/internal/logger"

	"github.com/gin-gonic/gin"
)

const (
	refreshTokenCookieName = "refresh_token"
	cookiePath             = "/api/auth"
)

type UserHandler struct {
	userService *UserService
	authConfig  *auth.Config
}

func NewUserHandler(userService *UserService, authConfig *auth.Config) *UserHandler {
	return &UserHandler{
		userService: userService,
		authConfig:  authConfig,
	}
}

func RegisterRoutes(r *gin.Engine, h *UserHandler, jwtService *auth.JWTService) {
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/register", h.Register)
		authGroup.POST("/login", h.Login)
		authGroup.POST("/refresh", h.RefreshToken)
		authGroup.POST("/logout", h.Logout)
	}

	// Protected routes
	userGroup := r.Group("/api/user")
	userGroup.Use(auth.AuthMiddleware(jwtService))
	{
		userGroup.GET("/me", h.GetCurrentUser)
	}
}

func (h *UserHandler) setRefreshTokenCookie(c *gin.Context, refreshToken string) {
	maxAge := int(h.authConfig.RefreshTokenTTL.Seconds())

	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		refreshTokenCookieName,
		refreshToken,
		maxAge,
		cookiePath,
		h.authConfig.CookieDomain,
		h.authConfig.CookieSecure,
		true,
	)
}

func (h *UserHandler) clearRefreshTokenCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		refreshTokenCookieName,
		"",
		-1,
		cookiePath,
		h.authConfig.CookieDomain,
		h.authConfig.CookieSecure,
		true,
	)
}

func (h *UserHandler) Register(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())

	var dto CreateUser
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	tokens, err := h.userService.Register(c.Request.Context(), &dto)
	if err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "user with this email already exists"})
			return
		}
		log.Error("failed to register user", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
		return
	}

	h.setRefreshTokenCookie(c, tokens.RefreshToken)

	c.JSON(http.StatusCreated, TokenResponse{
		AccessToken: tokens.AccessToken,
		ExpiresIn:   tokens.ExpiresIn,
	})
}

func (h *UserHandler) Login(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())

	var dto LoginRequest
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	tokens, err := h.userService.Login(c.Request.Context(), &dto)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		log.Error("failed to login user", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
		return
	}

	h.setRefreshTokenCookie(c, tokens.RefreshToken)

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken: tokens.AccessToken,
		ExpiresIn:   tokens.ExpiresIn,
	})
}

func (h *UserHandler) RefreshToken(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())

	refreshToken, err := c.Cookie(refreshTokenCookieName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token not found"})
		return
	}

	tokens, err := h.userService.RefreshTokens(c.Request.Context(), refreshToken)
	if err != nil {
		h.clearRefreshTokenCookie(c)

		if errors.Is(err, auth.ErrExpiredToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token has expired"})
			return
		}
		if errors.Is(err, auth.ErrInvalidToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
			return
		}
		log.Error("failed to refresh tokens", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to refresh tokens"})
		return
	}

	h.setRefreshTokenCookie(c, tokens.RefreshToken)

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken: tokens.AccessToken,
		ExpiresIn:   tokens.ExpiresIn,
	})
}

func (h *UserHandler) Logout(c *gin.Context) {
	h.clearRefreshTokenCookie(c)
	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())

	userID, ok := auth.GetUserIDFromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	user, err := h.userService.GetCurrentUser(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		log.Error("failed to get current user", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	c.JSON(http.StatusOK, user)
}
