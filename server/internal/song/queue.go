package song

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

const (
	AddSongTaskType = "song:add_song"
)

type AddSongTaskPayload struct {
	ID string `json:"id"`
}

type AddSongTaskHandler struct {
	songService *SongService
}

func NewAddSongTaskHandler(songService *SongService) *AddSongTaskHandler {
	return &AddSongTaskHandler{songService: songService}
}

func (h *AddSongTaskHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload AddSongTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	_, err := h.songService.AddSong(ctx, payload.ID)
	if err != nil {
		return fmt.Errorf("failed to add song: %w", err)
	}

	return nil
}
