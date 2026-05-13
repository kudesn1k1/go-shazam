package song

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SongMetadata struct {
	Title      string
	Artist     string
	DurationMs int
}

type DownloadedSong struct {
	Filename string
	Path     string
	SourceID string
}

type GetSongRequest struct {
	Link string `json:"link" binding:"required"`
}

type SongResponse struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	Artist     string  `json:"artist"`
	Duration   int     `json:"duration"`
	SourceID   string  `json:"source_id"`
	UploadedBy *string `json:"uploaded_by,omitempty"`
	CreatedAt  string  `json:"created_at"`
}

// PublicSongResponse omits uploader identity so the public catalog endpoint
// doesn't leak which user uploaded which song.
type PublicSongResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Artist    string `json:"artist"`
	Duration  int    `json:"duration"`
	SourceID  string `json:"source_id"`
	CreatedAt string `json:"created_at"`
}

func ToPublicSongResponse(s SongResponse) PublicSongResponse {
	return PublicSongResponse{
		ID:        s.ID,
		Title:     s.Title,
		Artist:    s.Artist,
		Duration:  s.Duration,
		SourceID:  s.SourceID,
		CreatedAt: s.CreatedAt,
	}
}

type SortField string
type SortOrder string

const (
	SortTitle     SortField = "title"
	SortArtist    SortField = "artist"
	SortDuration  SortField = "duration"
	SortCreatedAt SortField = "created_at"

	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)

var validSortFields = map[SortField]struct{}{
	SortTitle: {}, SortArtist: {}, SortDuration: {}, SortCreatedAt: {},
}

var validSortOrders = map[SortOrder]struct{}{
	SortAsc: {}, SortDesc: {},
}

type SongFilter struct {
	Q             string
	Artist        string
	UploadedBy    *uuid.UUID
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	Sort          SortField
	Order         SortOrder
	Limit         int
	Offset        int
}

// FilterError carries per-field messages for 422 responses.
type FilterError struct {
	Fields map[string]string
}

func (e *FilterError) Error() string {
	parts := make([]string, 0, len(e.Fields))
	for k, v := range e.Fields {
		parts = append(parts, fmt.Sprintf("%s: %s", k, v))
	}
	return strings.Join(parts, "; ")
}

func newFilterError(field, msg string) *FilterError {
	return &FilterError{Fields: map[string]string{field: msg}}
}

// parseBaseSongFilter parses and validates everything except the UploadedBy scope.
// Callers that need to scope to a specific uploader set filter.UploadedBy themselves
// after this function returns.
func parseBaseSongFilter(c *gin.Context) (SongFilter, error) {
	q := strings.TrimSpace(c.Query("q"))
	if len(q) > 200 {
		return SongFilter{}, newFilterError("q", "max 200 chars")
	}

	artist := strings.TrimSpace(c.Query("artist"))
	if len(artist) > 200 {
		return SongFilter{}, newFilterError("artist", "max 200 chars")
	}

	var createdAfter, createdBefore *time.Time
	if raw := c.Query("created_after"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return SongFilter{}, newFilterError("created_after", "must be RFC3339")
		}
		t = t.UTC()
		createdAfter = &t
	}
	if raw := c.Query("created_before"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return SongFilter{}, newFilterError("created_before", "must be RFC3339")
		}
		t = t.UTC()
		createdBefore = &t
	}
	if createdAfter != nil && createdBefore != nil && createdBefore.Before(*createdAfter) {
		return SongFilter{}, newFilterError("created_before", "must be >= created_after")
	}

	sort := SortField(strings.ToLower(c.DefaultQuery("sort", string(SortCreatedAt))))
	if _, ok := validSortFields[sort]; !ok {
		return SongFilter{}, newFilterError("sort", "must be one of: title, artist, duration, created_at")
	}

	order := SortOrder(strings.ToLower(c.DefaultQuery("order", string(SortDesc))))
	if _, ok := validSortOrders[order]; !ok {
		return SongFilter{}, newFilterError("order", "must be asc or desc")
	}

	page := parseIntDefault(c.Query("page"), 1)
	limit := parseIntDefault(c.Query("limit"), 20)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return SongFilter{
		Q:             q,
		Artist:        artist,
		CreatedAfter:  createdAfter,
		CreatedBefore: createdBefore,
		Sort:          sort,
		Order:         order,
		Limit:         limit,
		Offset:        (page - 1) * limit,
	}, nil
}

// parseUploadedByQuery parses an optional ?uploaded_by=<uuid> query param.
// Returns nil if the param is absent; returns a FilterError if the value is malformed.
func parseUploadedByQuery(c *gin.Context) (*uuid.UUID, error) {
	raw := c.Query("uploaded_by")
	if raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, newFilterError("uploaded_by", "must be a valid UUID")
	}
	return &id, nil
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
