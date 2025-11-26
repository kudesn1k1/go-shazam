package spotify

import (
	"context"
	"errors"
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

func (s *SpotifySongMetadataSource) GetSongsMetadata(ctx context.Context, id string) (*song.SongMetadata, error) {
	id, err := s.ExtractIdFromLink(id)
	if err != nil {
		return nil, err
	}

	err = s.ensureToken(ctx)
	if err != nil {
		return nil, err
	}

	track, err := s.client.GetTrack(ctx, spotify.ID(id))
	if err != nil {
		return nil, err
	}

	return &song.SongMetadata{
		Title:      track.Name,
		Artist:     track.Artists[0].Name,
		DurationMs: int(track.Duration),
	}, nil
}

func (s *SpotifySongMetadataSource) ExtractIdFromLink(link string) (string, error) {
	u, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	if !s.isValidLink(u) {
		return "", errors.New("invalid link")
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

	return "", errors.New("invalid link")
}

func (s *SpotifySongMetadataSource) isValidLink(u *url.URL) bool {
	return u.Host == "open.spotify.com" || u.Host == "www.open.spotify.com" || u.Host == "spotify.com" || u.Host == "www.spotify.com"
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
