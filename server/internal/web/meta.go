package web

import (
	"fmt"
	"html"
	"strings"

	"go-shazam/internal/song"
)

// PageMeta is the per-route SEO payload that gets injected into the HTML <head>
// before the SPA boots. Everything here is what social bots and the slower
// crawlers will see; the SPA can additionally update these client-side via
// @unhead/vue once it mounts, but those updates aren't visible to bots.
type PageMeta struct {
	Title       string
	Description string
	CanonicalURL string
	OGType      string
	OGImage     string
	Robots      string // "index, follow" | "noindex, nofollow" | "noindex"
	JSONLD      string // pre-serialized JSON, optional
}

const (
	defaultTitle       = "Go Shazam — Identify the music around you"
	defaultDescription = "Tap to identify any song playing around you. Browse our growing catalog of recognized tracks and explore by artist, title, or recency."
	siteName           = "Go Shazam"
)

// HomeMeta returns the meta block for the landing page.
func HomeMeta(origin string) PageMeta {
	return PageMeta{
		Title:        defaultTitle,
		Description:  defaultDescription,
		CanonicalURL: origin + "/",
		OGType:       "website",
		Robots:       "index, follow",
	}
}

// CatalogMeta returns the meta block for the public catalog index.
func CatalogMeta(origin string) PageMeta {
	return PageMeta{
		Title:        "Music catalog — Go Shazam",
		Description:  "Browse the songs Go Shazam already recognizes. Search by title or artist, sort by recency or duration.",
		CanonicalURL: origin + "/catalog",
		OGType:       "website",
		Robots:       "index, follow",
	}
}

// SongMeta returns the meta block for a specific song detail page, including
// JSON-LD MusicRecording structured data so Google can render rich results.
func SongMeta(origin string, s *song.SongResponse) PageMeta {
	canonical := fmt.Sprintf("%s/catalog/%s", origin, s.ID)
	title := fmt.Sprintf("%s by %s — Go Shazam", s.Title, s.Artist)
	description := fmt.Sprintf(`Listen to "%s" by %s. Part of the Go Shazam catalog.`, s.Title, s.Artist)

	return PageMeta{
		Title:        title,
		Description:  description,
		CanonicalURL: canonical,
		OGType:       "music.song",
		Robots:       "index, follow",
		JSONLD:       musicRecordingJSONLD(s, canonical),
	}
}

// PrivateMeta is used for noindex routes (profile, my-songs, admin pages, 404).
// The SPA still renders normally; we just tell crawlers to skip this URL.
func PrivateMeta(origin, path string) PageMeta {
	return PageMeta{
		Title:        "Go Shazam",
		Description:  defaultDescription,
		CanonicalURL: origin + path,
		OGType:       "website",
		Robots:       "noindex, nofollow",
	}
}

// renderHeadBlock produces the lines that get injected into the <head> of the
// HTML template. Every dynamic value is HTML-escaped; the JSON-LD is emitted
// inside a <script> tag where HTML escaping would corrupt it, so we use a
// JSON-safe escape (handled in jsonld.go).
func (m PageMeta) renderHeadBlock() string {
	var b strings.Builder

	fmt.Fprintf(&b, "<title>%s</title>\n", html.EscapeString(m.Title))
	fmt.Fprintf(&b, `    <meta name="description" content="%s" />`+"\n", html.EscapeString(m.Description))

	if m.Robots != "" {
		fmt.Fprintf(&b, `    <meta name="robots" content="%s" />`+"\n", html.EscapeString(m.Robots))
	}
	if m.CanonicalURL != "" {
		fmt.Fprintf(&b, `    <link rel="canonical" href="%s" />`+"\n", html.EscapeString(m.CanonicalURL))
	}

	fmt.Fprintf(&b, `    <meta property="og:site_name" content="%s" />`+"\n", html.EscapeString(siteName))
	fmt.Fprintf(&b, `    <meta property="og:title" content="%s" />`+"\n", html.EscapeString(m.Title))
	fmt.Fprintf(&b, `    <meta property="og:description" content="%s" />`+"\n", html.EscapeString(m.Description))
	if m.OGType != "" {
		fmt.Fprintf(&b, `    <meta property="og:type" content="%s" />`+"\n", html.EscapeString(m.OGType))
	}
	if m.CanonicalURL != "" {
		fmt.Fprintf(&b, `    <meta property="og:url" content="%s" />`+"\n", html.EscapeString(m.CanonicalURL))
	}
	if m.OGImage != "" {
		fmt.Fprintf(&b, `    <meta property="og:image" content="%s" />`+"\n", html.EscapeString(m.OGImage))
	}

	if m.JSONLD != "" {
		fmt.Fprintf(&b, `    <script type="application/ld+json">%s</script>`+"\n", m.JSONLD)
	}

	return b.String()
}
