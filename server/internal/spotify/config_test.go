package spotify

import (
	"os"
	"testing"
	"time"

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
	assert.Equal(t, 10*time.Second, config.RequestTimeout)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 500*time.Millisecond, config.RetryBackoff)
	assert.Equal(t, 10.0, config.RateLimitRPS)
	assert.Equal(t, 20, config.RateLimitBurst)
}

func TestLoadConfig_OverridesViaEnv(t *testing.T) {
	os.Setenv("SPOTIFY_CLIENT_ID", "id")
	os.Setenv("SPOTIFY_CLIENT_SECRET", "secret")
	os.Setenv("SPOTIFY_REQUEST_TIMEOUT", "3s")
	os.Setenv("SPOTIFY_MAX_RETRIES", "5")
	defer func() {
		os.Unsetenv("SPOTIFY_CLIENT_ID")
		os.Unsetenv("SPOTIFY_CLIENT_SECRET")
		os.Unsetenv("SPOTIFY_REQUEST_TIMEOUT")
		os.Unsetenv("SPOTIFY_MAX_RETRIES")
	}()

	config := LoadConfig()

	assert.Equal(t, 3*time.Second, config.RequestTimeout)
	assert.Equal(t, 5, config.MaxRetries)
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
