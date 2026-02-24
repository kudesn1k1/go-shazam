package song

import (
	"errors"
	"net/http"

	"go-shazam/internal/auth"
	"go-shazam/internal/logger"
	"go-shazam/internal/utils/pagination"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SongHandler struct {
	songService *SongService
}

func NewSongHandler(songService *SongService) *SongHandler {
	return &SongHandler{songService: songService}
}

func RegisterRoutes(r *gin.Engine, h *SongHandler, jwtService *auth.JWTService) {
	authMiddleware := auth.AuthMiddleware(jwtService)
	adminMiddleware := auth.RoleMiddleware("admin")

	r.POST("/api/song/add", authMiddleware, h.Add)

	userSongGroup := r.Group("/api/user")
	userSongGroup.Use(authMiddleware)
	{
		userSongGroup.GET("/songs", h.GetMySongs)
	}

	adminSongGroup := r.Group("/api/users")
	adminSongGroup.Use(authMiddleware, adminMiddleware)
	{
		adminSongGroup.GET("/:id/songs", h.GetUserSongs)
	}
}

func (h *SongHandler) Add(c *gin.Context) {
	songRequest := GetSongRequest{}

	if err := c.ShouldBind(&songRequest); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	userID, _ := auth.GetUserIDFromContext(c.Request.Context())
	err := h.songService.EnqueueSong(c.Request.Context(), songRequest.Link, &userID)
	if err != nil {
		if errors.Is(err, ErrSongTaskAlreadyExists) {
			c.JSON(http.StatusBadRequest, gin.H{"error": ErrSongTaskAlreadyExists.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "We will add this song soon"})
}

func (h *SongHandler) GetMySongs(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())

	userID, ok := auth.GetUserIDFromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	p := pagination.ParseParams(c)
	songs, total, err := h.songService.GetUserSongs(c.Request.Context(), userID, p.Page, p.Limit)
	if err != nil {
		log.Error("failed to get user songs", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get songs"})
		return
	}

	c.JSON(http.StatusOK, pagination.NewResponse(songs, total, p.Page, p.Limit))
}

func (h *SongHandler) GetUserSongs(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	p := pagination.ParseParams(c)
	songs, total, err := h.songService.GetUserSongs(c.Request.Context(), userID, p.Page, p.Limit)
	if err != nil {
		log.Error("failed to get user songs", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get songs"})
		return
	}

	c.JSON(http.StatusOK, pagination.NewResponse(songs, total, p.Page, p.Limit))
}
