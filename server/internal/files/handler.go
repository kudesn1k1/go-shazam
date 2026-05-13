package files

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"go-shazam/internal/auth"

	"github.com/gin-gonic/gin"
)

var hashRegex = regexp.MustCompile(`^[0-9a-f]{64}$`)

// extensionFor maps the content types we accept into the extension browsers
// expect to see on a downloaded file. Unknown types fall back to "bin".
func extensionFor(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/webp":
		return "webp"
	default:
		return "bin"
	}
}

type FilesHandler struct {
	svc *FilesService
	cfg *Config
}

func NewFilesHandler(svc *FilesService, cfg *Config) *FilesHandler {
	return &FilesHandler{svc: svc, cfg: cfg}
}

// RegisterRoutes wires the files endpoints.
//
// GET /api/files/:hash is intentionally unauthenticated: stored objects are
// content-addressed by SHA-256, so the hash itself is the minor access capability.
// Leaving it public lets browsers
// and any future CDN cache the immutable response without an auth round-trip.
// Treat this endpoint as serving public, non-secret blobs only.
//
// POST /api/files stays authenticated to gate who can add new objects.
func RegisterRoutes(r *gin.Engine, h *FilesHandler, jwtService *auth.JWTService) {
	r.GET("/api/files/:hash", h.Get)

	authed := r.Group("/api/files")
	authed.Use(auth.AuthMiddleware(jwtService))
	{
		authed.POST("", h.Upload)
	}
}

func (h *FilesHandler) Upload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.cfg.MaxUploadBytes)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file too large"})
			return
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "missing file field"})
		return
	}
	defer file.Close()

	resp, err := h.svc.Upload(c.Request.Context(), file, header.Size)
	if err != nil {
		switch {
		case errors.Is(err, ErrFileTooLarge):
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error()})
		case errors.Is(err, ErrUnsupportedMIME):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
		}
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *FilesHandler) Get(c *gin.Context) {
	hash := c.Param("hash")
	if !hashRegex.MatchString(hash) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hash format"})
		return
	}

	entity, err := h.svc.GetByHash(c.Request.Context(), hash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "lookup failed"})
		return
	}
	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	etag := fmt.Sprintf("\"%s\"", hash)
	if c.GetHeader("If-None-Match") == etag {
		c.Status(http.StatusNotModified)
		return
	}

	c.Header("ETag", etag)
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Header("Content-Type", entity.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", entity.SizeBytes))
	c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s.%s"`, hash[:12], extensionFor(entity.ContentType)))

	reader, err := h.svc.OpenObject(c.Request.Context(), hash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "storage error"})
		return
	}
	defer reader.Close()

	c.Status(http.StatusOK)
	if _, err := io.Copy(c.Writer, reader); err != nil {
		return
	}
}
