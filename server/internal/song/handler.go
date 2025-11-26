package song

import "github.com/gin-gonic/gin"

type SongHandler struct {
	songService *SongService
}

func NewSongHandler(songService *SongService) *SongHandler {
	return &SongHandler{songService: songService}
}

func RegisterRoutes(r *gin.Engine, h *SongHandler) {
	r.POST("/api/song/add", h.Add)
}

func (h *SongHandler) Add(c *gin.Context) {
	songRequest := GetSongRequest{}

	if err := c.ShouldBind(&songRequest); err != nil {
		c.JSON(422, gin.H{"error": err.Error()})
		return
	}

	err := h.songService.EnqueueSong(c.Request.Context(), songRequest.Link)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}

	c.JSON(200, gin.H{"message": "We will add this song soon"})
}
