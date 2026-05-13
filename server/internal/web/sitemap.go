package web

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"go-shazam/internal/logger"
	"go-shazam/internal/song"

	"github.com/gin-gonic/gin"
)

// sitemapEntry is one <url> in a sitemap.
type sitemapEntry struct {
	XMLName    xml.Name `xml:"url"`
	Loc        string   `xml:"loc"`
	LastMod    string   `xml:"lastmod,omitempty"`
	ChangeFreq string   `xml:"changefreq,omitempty"`
	Priority   string   `xml:"priority,omitempty"`
}

type sitemapDoc struct {
	XMLName xml.Name       `xml:"urlset"`
	XMLNS   string         `xml:"xmlns,attr"`
	URLs    []sitemapEntry `xml:""`
}

// sitemapCache holds the most recently rendered sitemap to avoid hitting the DB
// on every crawler request. TTL is short enough that newly added songs appear
// reasonably fast, long enough to absorb crawler bursts.
type sitemapCache struct {
	mu      sync.RWMutex
	body    []byte
	expires time.Time
}

const sitemapTTL = 15 * time.Minute

func (h *WebHandler) Sitemap(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())

	if cached := h.sitemapFromCache(); cached != nil {
		c.Data(http.StatusOK, "application/xml; charset=utf-8", cached)
		return
	}

	body, err := h.buildSitemap(c)
	if err != nil {
		log.Error("failed to build sitemap", "error", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	h.cacheSitemap(body)
	c.Data(http.StatusOK, "application/xml; charset=utf-8", body)
}

func (h *WebHandler) sitemapFromCache() []byte {
	h.smCache.mu.RLock()
	defer h.smCache.mu.RUnlock()
	if h.smCache.body == nil || time.Now().After(h.smCache.expires) {
		return nil
	}
	return h.smCache.body
}

func (h *WebHandler) cacheSitemap(body []byte) {
	h.smCache.mu.Lock()
	defer h.smCache.mu.Unlock()
	h.smCache.body = body
	h.smCache.expires = time.Now().Add(sitemapTTL)
}

func (h *WebHandler) buildSitemap(c *gin.Context) ([]byte, error) {
	origin := h.origin(c)

	// Static public routes.
	urls := []sitemapEntry{
		{Loc: origin + "/", ChangeFreq: "weekly", Priority: "1.0"},
		{Loc: origin + "/catalog", ChangeFreq: "daily", Priority: "0.8"},
	}

	// One entry per public song. We pull a single big page rather than
	// streaming because the sitemap is small (sitemaps cap at 50k URLs/50MB
	// per file, and this app is nowhere near that).
	songs, _, err := h.songService.ListSongs(c.Request.Context(), song.SongFilter{
		Sort:   song.SortCreatedAt,
		Order:  song.SortDesc,
		Limit:  1000,
		Offset: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("list songs for sitemap: %w", err)
	}

	for _, s := range songs {
		urls = append(urls, sitemapEntry{
			Loc:        fmt.Sprintf("%s/catalog/%s", origin, s.ID),
			LastMod:    truncateToDate(s.CreatedAt),
			ChangeFreq: "monthly",
			Priority:   "0.6",
		})
	}

	doc := sitemapDoc{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}

	body, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, err
	}
	return []byte(xml.Header + string(body)), nil
}

// truncateToDate yields the YYYY-MM-DD prefix of an RFC3339 timestamp, which
// is what the sitemap spec recommends for <lastmod>. Falls back to the raw
// value if it's shorter than 10 chars (should never happen for our DB output).
func truncateToDate(rfc3339 string) string {
	if len(rfc3339) < 10 {
		return rfc3339
	}
	if !strings.Contains(rfc3339[:10], "-") {
		return rfc3339
	}
	return rfc3339[:10]
}
