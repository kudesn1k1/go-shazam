package web

import (
	"encoding/json"
	"fmt"
	"strings"

	"go-shazam/internal/song"
)

// musicRecordingJSONLD builds a schema.org MusicRecording document for a song.
// Google uses this to render music-rich results; other crawlers use it as
// structured metadata.
//
// We deliberately keep the schema minimal — only fields we actually have
// reliable data for. Adding fields with placeholder/empty values can hurt
// SEO more than it helps.
func musicRecordingJSONLD(s *song.SongResponse, pageURL string) string {
	doc := map[string]interface{}{
		"@context": "https://schema.org",
		"@type":    "MusicRecording",
		"name":     s.Title,
		"byArtist": map[string]interface{}{
			"@type": "MusicGroup",
			"name":  s.Artist,
		},
		"duration": isoDuration(s.Duration),
		"url":      pageURL,
	}

	// Marshal then escape `</` so a stray closing-script-tag in the data can't
	// break out of the surrounding <script> block. This is the standard
	// defense for inline JSON-LD.
	raw, err := json.Marshal(doc)
	if err != nil {
		return ""
	}
	return strings.ReplaceAll(string(raw), "</", `<\/`)
}

// isoDuration converts milliseconds to ISO 8601 duration (e.g. 195000 → "PT3M15S").
func isoDuration(ms int) string {
	totalSec := ms / 1000
	min := totalSec / 60
	sec := totalSec % 60
	if min == 0 {
		return fmt.Sprintf("PT%dS", sec)
	}
	if sec == 0 {
		return fmt.Sprintf("PT%dM", min)
	}
	return fmt.Sprintf("PT%dM%dS", min, sec)
}
