package spotify

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpotifySongMetadataSource_ExtractSourceID_ValidSpotifyLink(t *testing.T) {
	source := &SpotifySongMetadataSource{}

	id, err := source.ExtractSourceID("https://open.spotify.com/track/4iV5W9uYEdYUVa79Axb7Rh")

	assert.NoError(t, err)
	assert.Equal(t, "4iV5W9uYEdYUVa79Axb7Rh", id)
}

func TestSpotifySongMetadataSource_ExtractSourceID_SpotifyLinkWithQueryParams(t *testing.T) {
	source := &SpotifySongMetadataSource{}

	id, err := source.ExtractSourceID("https://open.spotify.com/track/4iV5W9uYEdYUVa79Axb7Rh?si=abcd1234")

	assert.NoError(t, err)
	assert.Equal(t, "4iV5W9uYEdYUVa79Axb7Rh", id)
}

func TestSpotifySongMetadataSource_ExtractSourceID_InvalidHost(t *testing.T) {
	source := &SpotifySongMetadataSource{}

	id, err := source.ExtractSourceID("https://example.com/track/123")

	assert.Error(t, err)
	assert.Empty(t, id)
	assert.Contains(t, err.Error(), "invalid Spotify link")
}

func TestSpotifySongMetadataSource_ExtractSourceID_NoTrackID(t *testing.T) {
	source := &SpotifySongMetadataSource{}

	id, err := source.ExtractSourceID("https://open.spotify.com/album/123")

	assert.Error(t, err)
	assert.Empty(t, id)
	assert.Contains(t, err.Error(), "track ID not found")
}

func TestSpotifySongMetadataSource_isValidHost_ValidHosts(t *testing.T) {
	source := &SpotifySongMetadataSource{}

	assert.True(t, source.isValidHost(&url.URL{Host: "open.spotify.com"}))
	assert.True(t, source.isValidHost(&url.URL{Host: "www.spotify.com"}))
	assert.True(t, source.isValidHost(&url.URL{Host: "spotify.com"}))
	assert.True(t, source.isValidHost(&url.URL{Host: "www.open.spotify.com"}))
}

func TestSpotifySongMetadataSource_isValidHost_InvalidHost(t *testing.T) {
	source := &SpotifySongMetadataSource{}

	assert.False(t, source.isValidHost(&url.URL{Host: "example.com"}))
	assert.False(t, source.isValidHost(&url.URL{Host: "youtube.com"}))
}
