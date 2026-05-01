package user

import (
	"errors"
	"net/http"

	"go-shazam/internal/auth"
	"go-shazam/internal/logger"
	"go-shazam/internal/utils/pagination"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	userGroup := r.Group("/api/user")
	userGroup.Use(auth.AuthMiddleware(jwtService))
	{
		userGroup.GET("/me", h.GetCurrentUser)
		userGroup.PUT("/me/avatar", h.SetOwnAvatar)
		userGroup.DELETE("/me/avatar", h.ClearOwnAvatar)
	}

	adminGroup := r.Group("/api/users")
	adminGroup.Use(auth.AuthMiddleware(jwtService), auth.RoleMiddleware("admin"))
	{
		adminGroup.GET("", h.ListUsers)
		adminGroup.GET("/:id", h.GetUser)
		adminGroup.POST("/:id/roles", h.UpdateUserRoles)
		adminGroup.PUT("/:id/avatar", h.SetUserAvatar)
		adminGroup.DELETE("/:id/avatar", h.ClearUserAvatar)
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

func (h *UserHandler) ListUsers(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())

	p := pagination.ParseParams(c)
	users, total, err := h.userService.GetAllUsers(c.Request.Context(), p.Page, p.Limit)
	if err != nil {
		log.Error("failed to list users", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}

	c.JSON(http.StatusOK, pagination.NewResponse(users, total, p.Page, p.Limit))
}

func (h *UserHandler) GetUser(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		log.Error("failed to get user", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) SetOwnAvatar(c *gin.Context) {
	userID, _ := auth.GetUserIDFromContext(c.Request.Context())
	h.setAvatar(c, userID)
}

func (h *UserHandler) ClearOwnAvatar(c *gin.Context) {
	userID, _ := auth.GetUserIDFromContext(c.Request.Context())
	h.clearAvatar(c, userID)
}

func (h *UserHandler) SetUserAvatar(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	h.setAvatar(c, userID)
}

func (h *UserHandler) ClearUserAvatar(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	h.clearAvatar(c, userID)
}

func (h *UserHandler) setAvatar(c *gin.Context, userID uuid.UUID) {
	log := logger.FromContext(c.Request.Context())

	var req SetAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	url, err := h.userService.SetAvatar(c.Request.Context(), userID, req.FileHash)
	if err != nil {
		switch {
		case errors.Is(err, ErrAvatarHashInvalid):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		case errors.Is(err, ErrAvatarFileNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		default:
			log.Error("set avatar failed", "error", err, "user_id", userID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set avatar"})
		}
		return
	}

	c.JSON(http.StatusOK, AvatarResponse{AvatarURL: url})
}

func (h *UserHandler) clearAvatar(c *gin.Context, userID uuid.UUID) {
	log := logger.FromContext(c.Request.Context())
	if err := h.userService.ClearAvatar(c.Request.Context(), userID); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		log.Error("clear avatar failed", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear avatar"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *UserHandler) UpdateUserRoles(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req UpdateRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.UpdateUserRoles(c.Request.Context(), userID, req.Roles); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		log.Error("failed to update user roles", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "roles updated"})
}
