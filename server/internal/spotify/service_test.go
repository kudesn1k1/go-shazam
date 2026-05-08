package spotify

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zmb3/spotify/v2"
)

// newDeadRedis returns a redis.Client pointed at a port nothing's listening on.
// Used to verify the rate-limiter fails-open: with Redis unreachable, calls
// must still proceed (cancellation is short-circuited via ctx.Err() before
// the limiter is consulted, so the cancellation test below works regardless).
func newDeadRedis() *redis.Client {
	return redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
}

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

// Boot must not contact Spotify — the constructor must succeed even with bogus
// credentials, so the worker can come up when Spotify is unreachable.
func TestNewSpotifySongMetadataSource_DoesNotContactSpotify(t *testing.T) {
	cfg := &Config{
		ClientID:       "definitely-invalid",
		ClientSecret:   "definitely-invalid",
		RequestTimeout: 1 * time.Second,
		MaxRetries:     0,
		RetryBackoff:   10 * time.Millisecond,
		RateLimitRPS:   10,
		RateLimitBurst: 10,
	}

	src := NewSpotifySongMetadataSource(cfg, newDeadRedis())

	require.NotNil(t, src)
	_, ok := src.(*SpotifySongMetadataSource)
	assert.True(t, ok)
}

func TestIsTransient_5xxRetried(t *testing.T) {
	for _, status := range []int{500, 502, 503, 504} {
		err := spotify.Error{Status: status, Message: "boom"}
		assert.True(t, isTransient(err), "status %d should be transient", status)
	}
}

func TestIsTransient_4xxNotRetried(t *testing.T) {
	for _, status := range []int{400, 401, 403, 404} {
		err := spotify.Error{Status: status, Message: "nope"}
		assert.False(t, isTransient(err), "status %d should not be transient", status)
	}
}

func TestIsTransient_NetworkErrorRetried(t *testing.T) {
	err := errors.New("connection refused")
	assert.True(t, isTransient(err))
}

func TestIsTransient_CallerCancellationNotRetried(t *testing.T) {
	assert.False(t, isTransient(context.Canceled))
	assert.False(t, isTransient(fmt.Errorf("wrapped: %w", context.Canceled)))
}

func TestIsTransient_NilNotTransient(t *testing.T) {
	assert.False(t, isTransient(nil))
}

// Caller cancellation must short-circuit the retry loop without hitting the
// network or sleeping for the full backoff window.
func TestGetSongMetadata_RespectsCallerCancellation(t *testing.T) {
	cfg := &Config{
		ClientID:       "bogus",
		ClientSecret:   "bogus",
		RequestTimeout: 1 * time.Second,
		MaxRetries:     5,
		RetryBackoff:   1 * time.Second,
		RateLimitRPS:   100,
		RateLimitBurst: 10,
	}
	src := NewSpotifySongMetadataSource(cfg, newDeadRedis()).(*SpotifySongMetadataSource)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	_, err := src.GetSongMetadata(ctx, "4iV5W9uYEdYUVa79Axb7Rh")
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled), "expected context.Canceled, got %v", err)
	assert.Less(t, elapsed, 200*time.Millisecond, "cancelled call should return immediately, took %s", elapsed)
}
