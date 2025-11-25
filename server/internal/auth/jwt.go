package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken           string
	RefreshToken          string
	AccessTokenExpiresAt  time.Time
	RefreshTokenExpiresAt time.Time
}

type JWTService struct {
	config *Config
}

func NewJWTService(config *Config) *JWTService {
	return &JWTService{config: config}
}

// GenerateTokenPair creates both access and refresh tokens
func (s *JWTService) GenerateTokenPair(userID uuid.UUID) (*TokenPair, error) {
	accessToken, accessExp, err := s.generateToken(userID, AccessToken, s.config.AccessTokenSecret, s.config.AccessTokenTTL)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshExp, err := s.generateToken(userID, RefreshToken, s.config.RefreshTokenSecret, s.config.RefreshTokenTTL)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessExp,
		RefreshTokenExpiresAt: refreshExp,
	}, nil
}

func (s *JWTService) generateToken(userID uuid.UUID, tokenType TokenType, secret string, ttl time.Duration) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(ttl)

	claims := &Claims{
		UserID:    userID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "go-shazam",
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, err
	}

	return signedToken, expiresAt, nil
}

// ValidateAccessToken validates an access token and returns claims
func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	return s.validateToken(tokenString, s.config.AccessTokenSecret, AccessToken)
}

// ValidateRefreshToken validates a refresh token and returns claims
func (s *JWTService) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return s.validateToken(tokenString, s.config.RefreshTokenSecret, RefreshToken)
}

// TODO: check
func (s *JWTService) validateToken(tokenString, secret string, expectedType TokenType) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.TokenType != expectedType {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GetAccessTokenTTLSeconds returns access token TTL in seconds
func (s *JWTService) GetAccessTokenTTLSeconds() int64 {
	return int64(s.config.AccessTokenTTL.Seconds())
}
