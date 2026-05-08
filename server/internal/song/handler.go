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

	// Public catalog: unauthenticated, served via PublicSongResponse so uploader
	// identity is never exposed. Drives the SEO-indexable /catalog frontend.
	publicGroup := r.Group("/api/public/songs")
	{
		publicGroup.GET("", h.GetPublicSongs)
		publicGroup.GET("/:id", h.GetPublicSong)
	}

	userSongGroup := r.Group("/api/user")
	userSongGroup.Use(authMiddleware)
	{
		userSongGroup.GET("/songs", h.GetMySongs)
	}

	adminSongGroup := r.Group("/api/songs")
	adminSongGroup.Use(authMiddleware, adminMiddleware)
	{
		adminSongGroup.GET("", h.GetAllSongs)
		adminSongGroup.DELETE("/:id", h.Delete)
	}

	adminUserSongGroup := r.Group("/api/users")
	adminUserSongGroup.Use(authMiddleware, adminMiddleware)
	{
		adminUserSongGroup.GET("/:id/songs", h.GetUserSongs)
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

	filter, err := parseBaseSongFilter(c)
	if err != nil {
		respondFilterError(c, err)
		return
	}
	filter.UploadedBy = &userID

	songs, total, err := h.songService.ListSongs(c.Request.Context(), filter)
	if err != nil {
		log.Error("failed to list my songs", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get songs"})
		return
	}

	page := filter.Offset/filter.Limit + 1
	c.JSON(http.StatusOK, pagination.NewResponse(songs, total, page, filter.Limit))
}

func (h *SongHandler) GetAllSongs(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())

	filter, err := parseBaseSongFilter(c)
	if err != nil {
		respondFilterError(c, err)
		return
	}
	uploader, err := parseUploadedByQuery(c)
	if err != nil {
		respondFilterError(c, err)
		return
	}
	filter.UploadedBy = uploader

	songs, total, err := h.songService.ListSongs(c.Request.Context(), filter)
	if err != nil {
		log.Error("failed to list all songs", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get songs"})
		return
	}

	page := filter.Offset/filter.Limit + 1
	c.JSON(http.StatusOK, pagination.NewResponse(songs, total, page, filter.Limit))
}

func (h *SongHandler) Delete(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid song id"})
		return
	}

	if err := h.songService.DeleteSong(c.Request.Context(), id); err != nil {
		log.Error("failed to delete song", "error", err, "song_id", id)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete song"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "song deleted successfully"})
}

func (h *SongHandler) GetPublicSongs(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())

	filter, err := parseBaseSongFilter(c)
	if err != nil {
		respondFilterError(c, err)
		return
	}

	songs, total, err := h.songService.ListSongs(c.Request.Context(), filter)
	if err != nil {
		log.Error("failed to list public songs", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get songs"})
		return
	}

	public := make([]PublicSongResponse, len(songs))
	for i, s := range songs {
		public[i] = ToPublicSongResponse(s)
	}

	page := filter.Offset/filter.Limit + 1
	c.JSON(http.StatusOK, pagination.NewResponse(public, total, page, filter.Limit))
}

func (h *SongHandler) GetPublicSong(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid song id"})
		return
	}

	sg, err := h.songService.GetSong(c.Request.Context(), id)
	if err != nil {
		log.Error("failed to get public song", "error", err, "song_id", id)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get song"})
		return
	}
	if sg == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.JSON(http.StatusOK, ToPublicSongResponse(*sg))
}

func (h *SongHandler) GetUserSongs(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	filter, err := parseBaseSongFilter(c)
	if err != nil {
		respondFilterError(c, err)
		return
	}
	filter.UploadedBy = &userID

	songs, total, err := h.songService.ListSongs(c.Request.Context(), filter)
	if err != nil {
		log.Error("failed to list user songs", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get songs"})
		return
	}

	page := filter.Offset/filter.Limit + 1
	c.JSON(http.StatusOK, pagination.NewResponse(songs, total, page, filter.Limit))
}

func respondFilterError(c *gin.Context, err error) {
	var ferr *FilterError
	if errors.As(err, &ferr) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "validation failed", "fields": ferr.Fields})
		return
	}
	c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
}
