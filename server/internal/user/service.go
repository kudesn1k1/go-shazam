package user

import (
	"context"
	"errors"
	"time"

	"go-shazam/internal/auth"
	"go-shazam/internal/core/db"
	"go-shazam/internal/role"

	"github.com/google/uuid"
)

var (
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type UserService struct {
	userRepository     UserRepositoryInterface
	roleService        *role.Service
	cryptoService      *CryptoService
	jwtService         *auth.JWTService
	transactionManager *db.TransactionManager
}

func NewUserService(
	userRepository UserRepositoryInterface,
	roleService *role.Service,
	cryptoService *CryptoService,
	jwtService *auth.JWTService,
	transactionManager *db.TransactionManager,
) *UserService {
	return &UserService{
		userRepository:     userRepository,
		roleService:        roleService,
		cryptoService:      cryptoService,
		jwtService:         jwtService,
		transactionManager: transactionManager,
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

	_, err = db.Transactional(ctx, s.transactionManager, func(txCtx context.Context) (struct{}, error) {
		if err := s.userRepository.Create(txCtx, user); err != nil {
			return struct{}{}, err
		}
		return struct{}{}, s.roleService.AssignDefaultRole(txCtx, userID)
	})
	if err != nil {
		return nil, err
	}

	tokenPair, err := s.jwtService.GenerateTokenPair(userID, []string{role.RoleUser.String()})
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

	roles, err := s.roleService.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	tokenPair, err := s.jwtService.GenerateTokenPair(user.ID, role.RolesToStrings(roles))
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

	roles, err := s.roleService.GetUserRoles(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	tokenPair, err := s.jwtService.GenerateTokenPair(claims.UserID, role.RolesToStrings(roles))
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

	roles, err := s.roleService.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &UserResponse{
		ID:        user.ID.String(),
		Email:     email,
		Roles:     role.RolesToStrings(roles),
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *UserService) GetAllUsers(ctx context.Context, page, limit int) ([]UserResponse, int, error) {
	offset := (page - 1) * limit
	users, err := s.userRepository.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.userRepository.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	userIDs := make([]uuid.UUID, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}

	rolesMap, err := s.roleService.GetUsersRoles(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]UserResponse, len(users))
	for i, u := range users {
		email, err := s.cryptoService.DecryptEmail(u.Email)
		if err != nil {
			return nil, 0, err
		}

		responses[i] = UserResponse{
			ID:        u.ID.String(),
			Email:     email,
			Roles:     role.RolesToStrings(rolesMap[u.ID]),
			CreatedAt: u.CreatedAt.Format(time.RFC3339),
		}
	}

	return responses, total, nil
}

func (s *UserService) GetUserByID(ctx context.Context, userID uuid.UUID) (*UserResponse, error) {
	user, err := s.userRepository.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	email, err := s.cryptoService.DecryptEmail(user.Email)
	if err != nil {
		return nil, err
	}

	roles, err := s.roleService.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &UserResponse{
		ID:        user.ID.String(),
		Email:     email,
		Roles:     role.RolesToStrings(roles),
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *UserService) UpdateUserRoles(ctx context.Context, userID uuid.UUID, roleNames []string) error {
	roles := make([]role.Role, len(roleNames))
	for i, name := range roleNames {
		roles[i] = role.Role(name)
	}
	return s.roleService.UpdateUserRoles(ctx, userID, roles)
}
