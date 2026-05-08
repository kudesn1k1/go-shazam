package web

import (
	"errors"
	"io/fs"
	"net/http"
	"strings"

	"go-shazam/internal/logger"
	"go-shazam/internal/song"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WebHandler struct {
	tpl         *Template
	cfg         *Config
	songService *song.SongService
	staticFS    http.FileSystem
	smCache     sitemapCache
}

func NewWebHandler(cfg *Config, songService *song.SongService) (*WebHandler, error) {
	tpl, err := LoadTemplate(distFS)
	if err != nil {
		return nil, err
	}

	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return nil, err
	}

	return &WebHandler{
		tpl:         tpl,
		cfg:         cfg,
		songService: songService,
		staticFS:    http.FS(sub),
	}, nil
}

// RegisterRoutes wires the HTML, static-asset, sitemap, and robots routes.
//
// Why all this lives in Go and not nginx:
//   - Per-route meta injection (especially per-song JSON-LD on /catalog/:id)
//     has to happen on the initial HTML response, before any JS runs. Social
//     bots and most non-Google crawlers don't execute JS, so client-side
//     meta libraries don't reach them.
//   - Sitemap is data-driven (lists every public song), which is awkward in
//     nginx but trivial in the same process that owns the DB.
//
// nginx still proxies in front of this server; it just doesn't try to serve
// HTML or static assets out of its own copy of dist/.
func RegisterRoutes(r *gin.Engine, h *WebHandler) {
	r.GET("/sitemap.xml", h.Sitemap)
	r.GET("/robots.txt", h.Robots)

	// Static assets — anything under /assets/ comes straight from embedded dist.
	r.GET("/assets/*filepath", h.serveStatic)
	r.GET("/vite.svg", h.serveStatic)
	r.GET("/favicon.ico", h.serveStatic)

	// HTML routes. Each one resolves the right PageMeta and renders the
	// shared SPA shell template with that meta spliced in.
	r.GET("/", h.RenderHome)
	r.GET("/catalog", h.RenderCatalog)
	r.GET("/catalog/:id", h.RenderSongDetail)

	// Authenticated/admin SPA routes — rendered with noindex so crawlers don't
	// try to index a stripped-down login wall. The SPA itself enforces auth.
	for _, p := range []string{"/profile", "/my-songs", "/songs", "/users"} {
		r.GET(p, h.RenderPrivate)
	}
	r.GET("/users/:id", h.RenderPrivate)

	// NoRoute catches anything else — return 404 with noindex.
	r.NoRoute(h.RenderNotFound)
}

func (h *WebHandler) origin(c *gin.Context) string {
	if h.cfg.PublicOrigin != "" {
		return strings.TrimRight(h.cfg.PublicOrigin, "/")
	}
	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host
}

func (h *WebHandler) writeHTML(c *gin.Context, status int, meta PageMeta) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Status(status)
	_, _ = c.Writer.Write(h.tpl.Render(meta))
}

func (h *WebHandler) RenderHome(c *gin.Context) {
	h.writeHTML(c, http.StatusOK, HomeMeta(h.origin(c)))
}

func (h *WebHandler) RenderCatalog(c *gin.Context) {
	h.writeHTML(c, http.StatusOK, CatalogMeta(h.origin(c)))
}

func (h *WebHandler) RenderSongDetail(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())
	origin := h.origin(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		// Malformed ID → SPA can show a not-found state; we tell crawlers noindex
		// and return 404 so the URL doesn't get indexed.
		h.writeHTML(c, http.StatusNotFound, PrivateMeta(origin, c.Request.URL.Path))
		return
	}

	sg, err := h.songService.GetSong(c.Request.Context(), id)
	if err != nil {
		log.Error("failed to load song for HTML render", "error", err, "song_id", id)
		// Don't leak the failure into search; serve a generic shell with noindex.
		h.writeHTML(c, http.StatusInternalServerError, PrivateMeta(origin, c.Request.URL.Path))
		return
	}
	if sg == nil {
		h.writeHTML(c, http.StatusNotFound, PrivateMeta(origin, c.Request.URL.Path))
		return
	}

	h.writeHTML(c, http.StatusOK, SongMeta(origin, sg))
}

func (h *WebHandler) RenderPrivate(c *gin.Context) {
	h.writeHTML(c, http.StatusOK, PrivateMeta(h.origin(c), c.Request.URL.Path))
}

func (h *WebHandler) RenderNotFound(c *gin.Context) {
	// API 404s should keep their JSON shape; everything else gets the SPA shell
	// with noindex + 404 status so crawlers see "this page doesn't exist."
	if strings.HasPrefix(c.Request.URL.Path, "/api/") {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	h.writeHTML(c, http.StatusNotFound, PrivateMeta(h.origin(c), c.Request.URL.Path))
}

func (h *WebHandler) serveStatic(c *gin.Context) {
	// Strip the route prefix Gin gives us in *filepath, fall back to URL.Path
	// for the explicit /vite.svg-style routes.
	path := c.Param("filepath")
	if path == "" {
		path = c.Request.URL.Path
	} else {
		path = "/assets" + path
	}

	f, err := h.staticFS.Open(strings.TrimPrefix(path, "/"))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil || stat.IsDir() {
		c.Status(http.StatusNotFound)
		return
	}

	// Vite-hashed assets are immutable; everything else gets a short cache.
	if strings.HasPrefix(path, "/assets/") {
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		c.Header("Cache-Control", "public, max-age=3600")
	}
	http.ServeContent(c.Writer, c.Request, stat.Name(), stat.ModTime(), f.(interface {
		Read([]byte) (int, error)
		Seek(int64, int) (int64, error)
	}))
}
