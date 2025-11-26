package spotify

import (
	"context"
	"errors"
	"fmt"
	"go-shazam/internal/song"
	"net/url"
	"strings"
	"sync"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type SpotifySongMetadataSource struct {
	config *clientcredentials.Config
	token  *oauth2.Token
	client *spotify.Client
	mu     sync.Mutex
}

func NewSpotifySongMetadataSource(spotifyConfig *Config) (song.SongMetadataSource, error) {
	config := &clientcredentials.Config{
		ClientID:     spotifyConfig.ClientID,
		ClientSecret: spotifyConfig.ClientSecret,
		TokenURL:     spotifyauth.TokenURL,
	}

	ctx := context.Background()

	token, err := config.Token(ctx)
	if err != nil {
		return nil, err
	}

	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient, spotify.WithRetry(true))

	return &SpotifySongMetadataSource{
		config: config,
		token:  token,
		client: client,
	}, nil
}

func (s *SpotifySongMetadataSource) GetSongMetadata(ctx context.Context, id string) (*song.SongMetadata, error) {
	if err := s.ensureToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure token: %w", err)
	}

	track, err := s.client.GetTrack(ctx, spotify.ID(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get track from Spotify: %w", err)
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

	if s.token.Valid() {
		return nil
	}

	token, err := s.config.Token(ctx)
	if err != nil {
		return err
	}

	s.token = token
	httpClient := spotifyauth.New().Client(ctx, token)
	s.client = spotify.New(httpClient, spotify.WithRetry(true))

	return nil
}
