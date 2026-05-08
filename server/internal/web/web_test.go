package web

import (
	"encoding/json"
	"strings"
	"testing"
	"testing/fstest"

	"go-shazam/internal/song"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleIndexHTML = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <link rel="icon" type="image/svg+xml" href="/vite.svg" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Old Title</title>
    <meta name="description" content="Old description" />
    <meta property="og:title" content="Old OG" />
    <meta property="og:type" content="website" />
  </head>
  <body>
    <div id="app"></div>
    <script type="module" src="/assets/index-abc.js"></script>
  </body>
</html>`

func newTestTemplate(t *testing.T) *Template {
	t.Helper()
	fakeFS := fstest.MapFS{
		"dist/index.html": &fstest.MapFile{Data: []byte(sampleIndexHTML)},
	}
	tpl, err := LoadTemplate(fakeFS)
	require.NoError(t, err)
	return tpl
}

func TestLoadTemplate_StripsExistingMetaTags(t *testing.T) {
	tpl := newTestTemplate(t)
	out := string(tpl.Render(HomeMeta("https://example.com")))

	assert.NotContains(t, out, "Old Title", "old <title> must be removed")
	assert.NotContains(t, out, "Old description", "old description must be removed")
	assert.NotContains(t, out, "Old OG", "old OG title must be removed")
	// Non-meta content must survive untouched.
	assert.Contains(t, out, `<div id="app"></div>`)
	assert.Contains(t, out, `/assets/index-abc.js`)
	assert.Contains(t, out, `/vite.svg`)
}

func TestRender_HomeMetaInjected(t *testing.T) {
	tpl := newTestTemplate(t)
	out := string(tpl.Render(HomeMeta("https://example.com")))

	assert.Contains(t, out, "<title>Go Shazam — Identify the music around you</title>")
	assert.Contains(t, out, `<link rel="canonical" href="https://example.com/" />`)
	assert.Contains(t, out, `<meta property="og:type" content="website" />`)
	assert.Contains(t, out, `<meta name="robots" content="index, follow" />`)
}

func TestRender_PrivateMetaSetsNoindex(t *testing.T) {
	tpl := newTestTemplate(t)
	out := string(tpl.Render(PrivateMeta("https://example.com", "/profile")))

	assert.Contains(t, out, `<meta name="robots" content="noindex, nofollow" />`)
	assert.Contains(t, out, `<link rel="canonical" href="https://example.com/profile" />`)
}

func TestRender_SongMetaIncludesJSONLD(t *testing.T) {
	tpl := newTestTemplate(t)
	resp := &song.SongResponse{
		ID:       "11111111-1111-1111-1111-111111111111",
		Title:    "Test Track",
		Artist:   "Test Artist",
		Duration: 195000,
	}
	out := string(tpl.Render(SongMeta("https://example.com", resp)))

	assert.Contains(t, out, "Test Track by Test Artist — Go Shazam")
	assert.Contains(t, out, `https://example.com/catalog/11111111-1111-1111-1111-111111111111`)
	assert.Contains(t, out, `<meta property="og:type" content="music.song" />`)
	assert.Contains(t, out, `<script type="application/ld+json">`)
	assert.Contains(t, out, `"@type":"MusicRecording"`)
	assert.Contains(t, out, `"name":"Test Track"`)
	assert.Contains(t, out, `"PT3M15S"`)
}

func TestRender_HTMLEscapesUserContent(t *testing.T) {
	tpl := newTestTemplate(t)
	resp := &song.SongResponse{
		ID:       "22222222-2222-2222-2222-222222222222",
		Title:    `Evil <script>alert(1)</script>`,
		Artist:   `"; DROP TABLE songs; --`,
		Duration: 1000,
	}
	out := string(tpl.Render(SongMeta("https://example.com", resp)))

	// The raw </script> sequence must NOT appear inside the HTML attributes:
	// only its escaped form is acceptable. (JSON-LD body escapes </ to <\/.)
	assert.NotContains(t, out, `<script>alert(1)</script>`)
	// The escaped form is what should land in the title attribute.
	assert.Contains(t, out, `&lt;script&gt;alert(1)&lt;/script&gt;`)
}

func TestJSONLD_ValidJSON(t *testing.T) {
	resp := &song.SongResponse{
		ID:       "33333333-3333-3333-3333-333333333333",
		Title:    "Song",
		Artist:   "Artist",
		Duration: 60000,
	}
	raw := musicRecordingJSONLD(resp, "https://example.com/catalog/33333333-3333-3333-3333-333333333333")

	// Strip the </ -> <\/ escaping back out for parsing — the on-page form is
	// JSON-with-backslash-escapes, but those are valid JSON.
	var doc map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(strings.ReplaceAll(raw, `<\/`, "</")), &doc))

	assert.Equal(t, "https://schema.org", doc["@context"])
	assert.Equal(t, "MusicRecording", doc["@type"])
	assert.Equal(t, "Song", doc["name"])
	assert.Equal(t, "PT1M", doc["duration"])
}

func TestISODuration(t *testing.T) {
	cases := map[int]string{
		0:      "PT0S",
		1000:   "PT1S",
		60000:  "PT1M",
		61000:  "PT1M1S",
		195000: "PT3M15S",
	}
	for ms, want := range cases {
		assert.Equal(t, want, isoDuration(ms), "ms=%d", ms)
	}
}
