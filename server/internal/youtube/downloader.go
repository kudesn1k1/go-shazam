package youtube

import (
	"compress/gzip"
	"context"
	"fmt"
	"go-shazam/internal/logger"
	"go-shazam/internal/song"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/buger/jsonparser"
)

type YoutubeSongDownloader struct {
	config *Config
	http   *http.Client
}

func NewYoutubeSongDownloader(c *Config) song.SongDownloader {
	return &YoutubeSongDownloader{
		config: c,
		http:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *YoutubeSongDownloader) DownloadSong(ctx context.Context, data *song.SongMetadata, dir string) (*song.DownloadedSong, error) {
	logger := logger.FromContext(ctx)

	logger.Info("Downloading song", "title", data.Title, "artist", data.Artist)
	searchResults, err := s.searchVideoByParsing(ctx, fmt.Sprintf("%s %s", data.Title, data.Artist), 5)
	if err != nil {
		logger.Error("Failed to download song", "title", data.Title, "artist", data.Artist, "error", err)
		return nil, err
	}

	logger.Info("Found search results", "count", len(searchResults))
	for _, result := range searchResults {
		logger.Info("Downloading song", "title", result.Title, "duration", result.Duration, "id", result.ID)
	}

	return nil, nil
}

func (s *YoutubeSongDownloader) searchVideoByParsing(ctx context.Context, query string, limit int) (results []*SearchResult, err error) {
	logger := logger.FromContext(ctx)
	logger.Info("Parsing YouTube search page", "query", query)

	encodedQuery := url.QueryEscape(query)
	searchURL := fmt.Sprintf("https://www.youtube.com/results?search_query=%s", encodedQuery)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Accept-Language", "en")

	resp, err := s.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch search page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gz.Close()
		reader = gz
	}

	buffer, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	body := string(buffer)
	splitScript := strings.Split(body, `window["ytInitialData"] = `)
	if len(splitScript) != 2 {
		splitScript = strings.Split(body, `var ytInitialData = `)
	}

	if len(splitScript) != 2 {
		return nil, fmt.Errorf("ytInitialData not found")
	}
	splitScript = strings.Split(splitScript[1], `window["ytInitialPlayerResponse"] = null;`)
	jsonData := []byte(splitScript[0])

	index := 0
	var contents []byte

	for {
		contents, _, _, _ = jsonparser.Get(jsonData, "contents", "twoColumnSearchResultsRenderer", "primaryContents", "sectionListRenderer", "contents", fmt.Sprintf("[%d]", index), "itemSectionRenderer", "contents")
		_, _, _, err = jsonparser.Get(contents, "[0]", "carouselAdRenderer")

		if err == nil {
			index++
		} else {
			break
		}
	}

	_, err = jsonparser.ArrayEach(contents, func(value []byte, t jsonparser.ValueType, i int, err error) {
		if err != nil {
			return
		}

		if limit > 0 && len(results) >= limit {
			return
		}

		id, err := jsonparser.GetString(value, "videoRenderer", "videoId")
		if err != nil {
			return
		}

		title, err := jsonparser.GetString(value, "videoRenderer", "title", "runs", "[0]", "text")
		if err != nil {
			return
		}

		durationData, err := jsonparser.GetString(value, "videoRenderer", "lengthText", "simpleText")
		duration, err := parseYouTubeDuration(durationData)
		if err != nil {
			return
		}

		results = append(results, &SearchResult{
			Title:    title,
			Duration: duration,
			ID:       id,
		})
	})

	if err != nil {
		return results, err
	}

	return results, nil
}

func parseYouTubeDuration(duration string) (int, error) {
	parts := strings.Split(duration, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return 0, fmt.Errorf("invalid duration format: %s", duration)
	}

	var hours, minutes, seconds int
	if len(parts) == 3 {
		hours, _ = strconv.Atoi(parts[0])
		minutes, _ = strconv.Atoi(parts[1])
		seconds, _ = strconv.Atoi(parts[2])
	} else {
		minutes, _ = strconv.Atoi(parts[0])
		seconds, _ = strconv.Atoi(parts[1])
	}

	return hours*3600 + minutes*60 + seconds, nil
}
