package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestConfig() *Config {
	return &Config{
		AccessTokenSecret:  "test-access-secret-32characters!",
		RefreshTokenSecret: "test-refresh-secret-32character!",
		AccessTokenTTL:     15 * time.Minute,
		RefreshTokenTTL:    7 * 24 * time.Hour,
	}
}

func TestJWTService_GenerateTokenPair_ProducesValidAccessToken(t *testing.T) {
	svc := NewJWTService(newTestConfig())
	userID := uuid.New()

	pair, err := svc.GenerateTokenPair(userID, []string{"user", "admin"})
	require.NoError(t, err)
	require.NotEmpty(t, pair.AccessToken)
	require.NotEmpty(t, pair.RefreshToken)
	assert.True(t, pair.AccessTokenExpiresAt.After(time.Now()))

	claims, err := svc.ValidateAccessToken(pair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, AccessToken, claims.TokenType)
	assert.ElementsMatch(t, []string{"user", "admin"}, claims.Roles)
}

func TestJWTService_ValidateAccessToken_ExpiredTokenReturnsErrExpired(t *testing.T) {
	cfg := newTestConfig()
	cfg.AccessTokenTTL = -1 * time.Minute // already expired
	svc := NewJWTService(cfg)

	pair, err := svc.GenerateTokenPair(uuid.New(), []string{"user"})
	require.NoError(t, err)

	_, err = svc.ValidateAccessToken(pair.AccessToken)
	assert.ErrorIs(t, err, ErrExpiredToken)
}

func TestJWTService_ValidateAccessToken_WrongSignatureReturnsErrInvalid(t *testing.T) {
	svc := NewJWTService(newTestConfig())
	pair, err := svc.GenerateTokenPair(uuid.New(), []string{"user"})
	require.NoError(t, err)

	// Tamper with last 5 chars of the token
	tampered := pair.AccessToken[:len(pair.AccessToken)-5] + "AAAAA"
	_, err = svc.ValidateAccessToken(tampered)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestJWTService_ValidateAccessToken_WrongTokenTypeRejected(t *testing.T) {
	svc := NewJWTService(newTestConfig())
	pair, err := svc.GenerateTokenPair(uuid.New(), []string{"user"})
	require.NoError(t, err)

	// Refresh token presented to ValidateAccessToken — must fail.
	_, err = svc.ValidateAccessToken(pair.RefreshToken)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestJWTService_ValidateAccessToken_RejectsNonHMACAlgorithm(t *testing.T) {
	svc := NewJWTService(newTestConfig())
	token := jwt.NewWithClaims(jwt.SigningMethodNone, &Claims{
		UserID:    uuid.New(),
		TokenType: AccessToken,
		Roles:     []string{"user"},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
	})
	signed, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	_, err = svc.ValidateAccessToken(signed)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestJWTService_GetAccessTokenTTLSeconds_MatchesConfig(t *testing.T) {
	cfg := newTestConfig()
	svc := NewJWTService(cfg)
	assert.Equal(t, int64(15*60), svc.GetAccessTokenTTLSeconds())
}
