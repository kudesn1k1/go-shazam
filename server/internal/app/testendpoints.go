package app

import (
	"net/http"
	"os"
	"time"

	"go-shazam/internal/song"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// registerTestEndpoints mounts /api/test/* endpoints only when APP_ENV=test.
// These exist exclusively for the Playwright suite:
//   - POST /api/test/reset      → truncates user-data tables, re-seeds roles
//   - POST /api/test/seed-song  → inserts a song row directly (bypasses worker)
//   - POST /api/test/promote    → grants admin role to a user by email_hash
//
// In production builds (APP_ENV unset or any value other than "test") the
// function is a no-op so nothing is exposed.
func registerTestEndpoints(r *gin.Engine, db *sqlx.DB) {
	if os.Getenv("APP_ENV") != "test" {
		return
	}

	r.POST("/api/test/reset", func(c *gin.Context) {
		_, err := db.Exec(`
			TRUNCATE TABLE fingerprints, songs, user_roles, users, files
			RESTART IDENTITY CASCADE
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// Roles were dropped by CASCADE — restore.
		_, err = db.Exec(`
			INSERT INTO roles (name) VALUES ('user'), ('admin')
			ON CONFLICT (name) DO NOTHING
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})

	r.POST("/api/test/seed-song", func(c *gin.Context) {
		var req struct {
			Title      string  `json:"title" binding:"required"`
			Artist     string  `json:"artist" binding:"required"`
			Duration   int     `json:"duration"`
			SourceID   string  `json:"source_id"`
			UploadedBy *string `json:"uploaded_by"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		songID, err := uuid.NewV7()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var uploadedBy *uuid.UUID
		if req.UploadedBy != nil {
			id, err := uuid.Parse(*req.UploadedBy)
			if err != nil {
				c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid uploaded_by"})
				return
			}
			uploadedBy = &id
		}
		entity := &song.SongEntity{
			ID:         songID,
			Title:      req.Title,
			Artist:     req.Artist,
			Duration:   req.Duration,
			SourceID:   req.SourceID,
			UploadedBy: uploadedBy,
			CreatedAt:  time.Now().UTC(),
		}
		_, err = db.NamedExec(`
			INSERT INTO songs (id, title, artist, duration, source_id, uploaded_by, created_at)
			VALUES (:id, :title, :artist, :duration, :source_id, :uploaded_by, :created_at)
		`, entity)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"id": songID.String()})
	})

	r.POST("/api/test/promote", func(c *gin.Context) {
		var req struct {
			EmailHash string `json:"email_hash" binding:"required"`
			Role      string `json:"role" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		_, err := db.Exec(`
			INSERT INTO user_roles (user_id, role_id)
			SELECT u.id, r.id FROM users u CROSS JOIN roles r
			WHERE u.email_hash = $1 AND r.name = $2
			ON CONFLICT DO NOTHING
		`, req.EmailHash, req.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})
}
