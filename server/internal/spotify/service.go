package spotify

import (
	"context"
	"errors"
	"fmt"
	"go-shazam/internal/logger"
	"go-shazam/internal/song"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// ErrSpotifyUnavailable wraps transient failures (network, 5xx, exhausted retries)
// so callers can decide whether to reschedule the work or surface a soft error.
var ErrSpotifyUnavailable = errors.New("spotify temporarily unavailable")

// rateLimitKey is the shared key all replicas use; redis_rate scopes counts
// per-key, so every server/worker instance contends for the same budget.
const rateLimitKey = "ratelimit:spotify:metadata"

type SpotifySongMetadataSource struct {
	oauth   *clientcredentials.Config
	cfg     *Config
	limiter *redis_rate.Limiter

	mu     sync.Mutex
	token  *oauth2.Token
	client *spotify.Client
}

// NewSpotifySongMetadataSource intentionally does NOT contact Spotify at construction.
// Boot must succeed even if Spotify is unreachable; the OAuth token is acquired
// lazily on first use, so the worker can come up and start processing other queues.
//
// The rate limiter is Redis-backed (GCRA via redis_rate) so the budget is
// shared across every server/worker replica — horizontal scaling doesn't
// multiply the effective Spotify request rate.
func NewSpotifySongMetadataSource(spotifyConfig *Config, rdb *redis.Client) song.SongMetadataSource {
	oauth := &clientcredentials.Config{
		ClientID:     spotifyConfig.ClientID,
		ClientSecret: spotifyConfig.ClientSecret,
		TokenURL:     spotifyauth.TokenURL,
	}

	return &SpotifySongMetadataSource{
		oauth:   oauth,
		cfg:     spotifyConfig,
		limiter: redis_rate.NewLimiter(rdb),
	}
}

func (s *SpotifySongMetadataSource) GetSongMetadata(ctx context.Context, id string) (*song.SongMetadata, error) {
	if err := s.waitRateLimit(ctx); err != nil {
		return nil, err
	}

	var lastErr error
	for attempt := 0; attempt <= s.cfg.MaxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		metadata, err := s.fetchOnce(ctx, id)
		if err == nil {
			return metadata, nil
		}
		lastErr = err

		if !isTransient(err) {
			return nil, err
		}

		if attempt < s.cfg.MaxRetries {
			backoff := s.cfg.RetryBackoff * (1 << attempt)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}
	}
	return nil, fmt.Errorf("%w: %v", ErrSpotifyUnavailable, lastErr)
}

// waitRateLimit blocks until the Redis-backed rate limiter has headroom for
// one more request, or the caller's ctx is cancelled.
//
// Fail-open behavior on Redis errors: if Redis is unreachable, we let the
// request through rather than blocking all Spotify lookups. Spotify's own
// rate limiter (handled via the SDK's WithRetry on 429) is the backstop.
func (s *SpotifySongMetadataSource) waitRateLimit(ctx context.Context) error {
	log := logger.FromContext(ctx)
	limit := redis_rate.Limit{
		Rate:   int(s.cfg.RateLimitRPS),
		Burst:  s.cfg.RateLimitBurst,
		Period: time.Second,
	}

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		res, err := s.limiter.Allow(ctx, rateLimitKey, limit)
		if err != nil {
			log.Warn("redis rate-limit unavailable, failing open", "error", err)
			return nil
		}
		if res.Allowed > 0 {
			return nil
		}

		// Bucket is empty — sleep until the next slot opens, then re-check.
		// RetryAfter is bounded by the limit's Period, so this terminates.
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(res.RetryAfter):
		}
	}
}

func (s *SpotifySongMetadataSource) fetchOnce(ctx context.Context, id string) (*song.SongMetadata, error) {
	if err := s.ensureToken(ctx); err != nil {
		return nil, err
	}

	callCtx, cancel := context.WithTimeout(ctx, s.cfg.RequestTimeout)
	defer cancel()

	track, err := s.client.GetTrack(callCtx, spotify.ID(id))
	if err != nil {
		return nil, fmt.Errorf("get track: %w", err)
	}
	if len(track.Artists) == 0 {
		return nil, errors.New("track has no artists")
	}

	return &song.SongMetadata{
		Title:      track.Name,
		Artist:     track.Artists[0].Name,
		DurationMs: int(track.Duration),
	}, nil
}

func (s *SpotifySongMetadataSource) ExtractSourceID(link string) (string, error) {
	return s.extractIDFromLink(link)
}

func (s *SpotifySongMetadataSource) extractIDFromLink(link string) (string, error) {
	u, err := url.Parse(link)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	if !s.isValidHost(u) {
		return "", errors.New("invalid Spotify link: unsupported host")
	}

	pathParts := strings.Split(u.Path, "/")
	for i, part := range pathParts {
		if part == "track" && i+1 < len(pathParts) {
			trackID := pathParts[i+1]
			if index := strings.Index(trackID, "?"); index != -1 {
				trackID = trackID[:index]
			}
			return trackID, nil
		}
	}

	return "", errors.New("invalid Spotify link: track ID not found")
}

func (s *SpotifySongMetadataSource) isValidHost(u *url.URL) bool {
	validHosts := map[string]bool{
		"open.spotify.com":     true,
		"www.open.spotify.com": true,
		"spotify.com":          true,
		"www.spotify.com":      true,
	}
	return validHosts[u.Host]
}

func (s *SpotifySongMetadataSource) ensureToken(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.token != nil && s.token.Valid() && s.client != nil {
		return nil
	}

	tokenCtx, cancel := context.WithTimeout(ctx, s.cfg.RequestTimeout)
	defer cancel()

	token, err := s.oauth.Token(tokenCtx)
	if err != nil {
		return fmt.Errorf("acquire spotify token: %w", err)
	}

	s.token = token
	httpClient := spotifyauth.New().Client(ctx, token)
	s.client = spotify.New(httpClient, spotify.WithRetry(true))
	return nil
}

// isTransient classifies errors for the retry loop. 5xx and unrecognized errors
// (network, EOF, etc.) are retried; 4xx and explicit caller-cancellation are not.
// A wrapped DeadlineExceeded is treated as transient because it's typically the
// per-call timeout firing, not the caller's outer ctx — and if it IS the outer
// ctx, the next loop iteration exits on ctx.Err() before retrying.
func isTransient(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return false
	}

	var spErr spotify.Error
	if errors.As(err, &spErr) {
		return spErr.Status >= 500 && spErr.Status < 600
	}

	return true
}
