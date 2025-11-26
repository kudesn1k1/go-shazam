package spotify

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpotifySongMetadataSource_extractIdFromLink_ValidSpotifyLink(t *testing.T) {
	source := &SpotifySongMetadataSource{}

	id, err := source.ExtractIdFromLink("https://open.spotify.com/track/4iV5W9uYEdYUVa79Axb7Rh")

	assert.NoError(t, err)
	assert.Equal(t, "4iV5W9uYEdYUVa79Axb7Rh", id)
}

func TestSpotifySongMetadataSource_extractIdFromLink_SpotifyLinkWithQueryParams(t *testing.T) {
	source := &SpotifySongMetadataSource{}

	id, err := source.ExtractIdFromLink("https://open.spotify.com/track/4iV5W9uYEdYUVa79Axb7Rh?si=abcd1234")

	assert.NoError(t, err)
	assert.Equal(t, "4iV5W9uYEdYUVa79Axb7Rh", id)
}

func TestSpotifySongMetadataSource_extractIdFromLink_InvalidLink(t *testing.T) {
	source := &SpotifySongMetadataSource{}

	id, err := source.ExtractIdFromLink("https://example.com/track/123")

	assert.Error(t, err)
	assert.Empty(t, id)
	assert.Contains(t, err.Error(), "invalid link")
}

func TestSpotifySongMetadataSource_isValidLink_ValidHosts(t *testing.T) {
	source := &SpotifySongMetadataSource{}

	assert.True(t, source.isValidLink(&url.URL{Host: "open.spotify.com"}))
	assert.True(t, source.isValidLink(&url.URL{Host: "www.spotify.com"}))
	assert.True(t, source.isValidLink(&url.URL{Host: "spotify.com"}))
	assert.True(t, source.isValidLink(&url.URL{Host: "www.spotify.com"}))
}

func TestSpotifySongMetadataSource_isValidLink_InvalidHost(t *testing.T) {
	source := &SpotifySongMetadataSource{}

	assert.False(t, source.isValidLink(&url.URL{Host: "example.com"}))
	assert.False(t, source.isValidLink(&url.URL{Host: "youtube.com"}))
}
