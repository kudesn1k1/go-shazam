package song

import "github.com/gin-gonic/gin"

type SongHandler struct {
	songService *SongService
}

func NewSongHandler(r *gin.Engine, songService *SongService) *SongHandler {
	h := &SongHandler{songService: songService}

	r.POST("/api/song/add", h.Get)

	return h
}

func (h *SongHandler) Get(c *gin.Context) {
	songRequest := GetSongRequest{}

	if err := c.ShouldBind(&songRequest); err != nil {
		c.JSON(422, gin.H{"error": err.Error()})
		return
	}

	song, err := h.songService.GetSongsMetadata(c.Request.Context(), songRequest.Link)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}

	c.JSON(200, song)
}
