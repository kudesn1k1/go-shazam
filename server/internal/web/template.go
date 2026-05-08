package web

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"regexp"
	"strings"
)

// Template wraps the SPA's built index.html and lets us swap the <head>'s SEO
// block per request. The original index.html (post-Vite-build) carries the
// hashed asset references the SPA needs; we keep <body> and the script/link
// tags untouched and only rewrite the meta block.
type Template struct {
	prefix []byte // everything up to and including <head>
	suffix []byte // </head> onward
}

// headRegex matches the existing meta tags we want to replace. We strip them
// out wholesale and let renderHeadBlock emit fresh ones; this avoids having to
// thread per-tag state through the template.
//
// The set we strip:
//   - <title>...</title>
//   - <meta name="description"|"robots" .../>
//   - <link rel="canonical" .../>
//   - any <meta property="og:...".../>
var (
	titleRegex      = regexp.MustCompile(`(?is)<title[^>]*>.*?</title>\s*`)
	descRegex       = regexp.MustCompile(`(?is)<meta\s+name=["']description["'][^>]*/?>\s*`)
	robotsRegex     = regexp.MustCompile(`(?is)<meta\s+name=["']robots["'][^>]*/?>\s*`)
	canonicalRegex  = regexp.MustCompile(`(?is)<link\s+rel=["']canonical["'][^>]*/?>\s*`)
	ogRegex         = regexp.MustCompile(`(?is)<meta\s+property=["']og:[^"']+["'][^>]*/?>\s*`)
	headCloseRegex  = regexp.MustCompile(`(?i)</head>`)
)

// LoadTemplate parses the SPA's built index.html out of the embedded FS and
// produces a Template ready to render. Called once at server startup.
func LoadTemplate(embedded fs.FS) (*Template, error) {
	data, err := fs.ReadFile(embedded, "dist/index.html")
	if err != nil {
		return nil, fmt.Errorf("read embedded index.html: %w", err)
	}

	stripped := titleRegex.ReplaceAll(data, nil)
	stripped = descRegex.ReplaceAll(stripped, nil)
	stripped = robotsRegex.ReplaceAll(stripped, nil)
	stripped = canonicalRegex.ReplaceAll(stripped, nil)
	stripped = ogRegex.ReplaceAll(stripped, nil)

	// Find the closing </head> and split there. Everything before is the
	// "prefix" we'll prepend the new head block to; everything from </head>
	// onward is the "suffix" we re-append untouched.
	loc := headCloseRegex.FindIndex(stripped)
	if loc == nil {
		return nil, errors.New("index.html has no </head>")
	}

	return &Template{
		prefix: append([]byte(nil), stripped[:loc[0]]...),
		suffix: append([]byte(nil), stripped[loc[0]:]...),
	}, nil
}

// Render produces the final HTML for one request, splicing the per-route meta
// block in just before </head>.
func (t *Template) Render(meta PageMeta) []byte {
	headBlock := meta.renderHeadBlock()

	// Prefix already ends inside <head>; we want a clean indented insertion.
	var buf bytes.Buffer
	buf.Grow(len(t.prefix) + len(headBlock) + len(t.suffix))
	buf.Write(t.prefix)
	if !bytes.HasSuffix(t.prefix, []byte("\n")) {
		buf.WriteByte('\n')
	}
	buf.WriteString("    ")
	// Indent every emitted line by 4 spaces for readability when viewed.
	buf.WriteString(strings.ReplaceAll(strings.TrimRight(headBlock, "\n"), "\n", "\n    "))
	buf.WriteByte('\n')
	buf.WriteString("  ")
	buf.Write(t.suffix)
	return buf.Bytes()
}
