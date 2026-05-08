package web

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// robotsBody allows public catalog routes, blocks the API surface and every
// authenticated SPA route. The Sitemap line points to our dynamic sitemap.
const robotsBody = `User-agent: *
Allow: /
Allow: /catalog
Allow: /catalog/
Disallow: /api/
Disallow: /profile
Disallow: /my-songs
Disallow: /songs
Disallow: /users
Disallow: /users/

Sitemap: %s/sitemap.xml
`

func (h *WebHandler) Robots(c *gin.Context) {
	body := fmt.Sprintf(robotsBody, h.origin(c))
	c.Header("Cache-Control", "public, max-age=3600")
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(body))
}
