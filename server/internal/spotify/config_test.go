package spotify

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_Success(t *testing.T) {
	os.Setenv("SPOTIFY_CLIENT_ID", "test-client-id")
	os.Setenv("SPOTIFY_CLIENT_SECRET", "test-client-secret")
	defer func() {
		os.Unsetenv("SPOTIFY_CLIENT_ID")
		os.Unsetenv("SPOTIFY_CLIENT_SECRET")
	}()

	config := LoadConfig()

	assert.Equal(t, "test-client-id", config.ClientID)
	assert.Equal(t, "test-client-secret", config.ClientSecret)
}

func TestLoadConfig_MissingClientID(t *testing.T) {
	os.Setenv("SPOTIFY_CLIENT_SECRET", "test-client-secret")
	defer func() {
		os.Unsetenv("SPOTIFY_CLIENT_SECRET")
	}()

	assert.Panics(t, func() {
		LoadConfig()
	})
}

func TestLoadConfig_MissingClientSecret(t *testing.T) {
	os.Setenv("SPOTIFY_CLIENT_ID", "test-client-id")
	defer func() {
		os.Unsetenv("SPOTIFY_CLIENT_ID")
	}()

	assert.Panics(t, func() {
		LoadConfig()
	})
}

func TestLoadConfig_EmptyValues(t *testing.T) {
	os.Setenv("SPOTIFY_CLIENT_ID", "")
	os.Setenv("SPOTIFY_CLIENT_SECRET", "")
	defer func() {
		os.Unsetenv("SPOTIFY_CLIENT_ID")
		os.Unsetenv("SPOTIFY_CLIENT_SECRET")
	}()

	assert.Panics(t, func() {
		LoadConfig()
	})
}
