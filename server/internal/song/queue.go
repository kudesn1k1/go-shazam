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
	fmt.Printf("[Worker] Received task: %s\n", task.Type())

	var payload AddSongTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		fmt.Printf("[Worker] Failed to unmarshal payload: %v\n", err)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	fmt.Printf("[Worker] Processing song with ID: %s\n", payload.ID)

	songMeta, err := h.songService.AddSong(ctx, payload.ID)
	if err != nil {
		fmt.Printf("[Worker] Failed to add song: %v\n", err)
		return fmt.Errorf("failed to add song: %w", err)
	}

	fmt.Printf("[Worker] Successfully added song: %s - %s\n", songMeta.Artist, songMeta.Title)
	return nil
}
