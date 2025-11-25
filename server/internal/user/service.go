package user

import (
	"context"
	"errors"
	"time"

	"go-shazam/internal/auth"

	"github.com/google/uuid"
)

var (
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type UserService struct {
	userRepository UserRepositoryInterface
	cryptoService  *CryptoService
	jwtService     *auth.JWTService
}

func NewUserService(
	userRepository UserRepositoryInterface,
	cryptoService *CryptoService,
	jwtService *auth.JWTService,
) *UserService {
	return &UserService{
		userRepository: userRepository,
		cryptoService:  cryptoService,
		jwtService:     jwtService,
	}
}

func (s *UserService) Register(ctx context.Context, dto *CreateUser) (*InternalTokenResponse, error) {
	emailHash := s.cryptoService.HashEmail(dto.Email)

	exists, err := s.userRepository.ExistsByEmailHash(ctx, emailHash)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserAlreadyExists
	}

	encryptedEmail, err := s.cryptoService.EncryptEmail(dto.Email)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := s.cryptoService.HashPassword(dto.Password)
	if err != nil {
		return nil, err
	}

	userID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := &UserEntity{
		ID:             userID,
		Email:          encryptedEmail,
		EmailHash:      emailHash,
		HashedPassword: hashedPassword,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.userRepository.Create(ctx, user); err != nil {
		return nil, err
	}

	tokenPair, err := s.jwtService.GenerateTokenPair(userID)
	if err != nil {
		return nil, err
	}

	return &InternalTokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    s.jwtService.GetAccessTokenTTLSeconds(),
	}, nil
}

func (s *UserService) Login(ctx context.Context, dto *LoginRequest) (*InternalTokenResponse, error) {
	emailHash := s.cryptoService.HashEmail(dto.Email)

	user, err := s.userRepository.FindByEmailHash(ctx, emailHash)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !s.cryptoService.VerifyPassword(dto.Password, user.HashedPassword) {
		return nil, ErrInvalidCredentials
	}

	tokenPair, err := s.jwtService.GenerateTokenPair(user.ID)
	if err != nil {
		return nil, err
	}

	return &InternalTokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    s.jwtService.GetAccessTokenTTLSeconds(),
	}, nil
}

func (s *UserService) RefreshTokens(ctx context.Context, refreshToken string) (*InternalTokenResponse, error) {
	claims, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	_, err = s.userRepository.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	tokenPair, err := s.jwtService.GenerateTokenPair(claims.UserID)
	if err != nil {
		return nil, err
	}

	return &InternalTokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    s.jwtService.GetAccessTokenTTLSeconds(),
	}, nil
}

func (s *UserService) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*UserResponse, error) {
	user, err := s.userRepository.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	email, err := s.cryptoService.DecryptEmail(user.Email)
	if err != nil {
		return nil, err
	}

	return &UserResponse{
		ID:    user.ID.String(),
		Email: email,
	}, nil
}
