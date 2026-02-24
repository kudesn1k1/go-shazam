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
	roleRepository RoleRepositoryInterface
	cryptoService  *CryptoService
	jwtService     *auth.JWTService
}

func NewUserService(
	userRepository UserRepositoryInterface,
	roleRepository RoleRepositoryInterface,
	cryptoService *CryptoService,
	jwtService *auth.JWTService,
) *UserService {
	return &UserService{
		userRepository: userRepository,
		roleRepository: roleRepository,
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

	//TODO add transaction
	defaultRole, err := s.roleRepository.FindRoleByName(ctx, "user")
	if err != nil {
		return nil, err
	}

	if err := s.roleRepository.AssignRole(ctx, userID, defaultRole.ID); err != nil {
		return nil, err
	}

	tokenPair, err := s.jwtService.GenerateTokenPair(userID, []string{"user"})
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

	roles, err := s.roleRepository.FindRolesByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	roleNames := make([]string, len(roles))
	for i, r := range roles {
		roleNames[i] = r.Name
	}

	tokenPair, err := s.jwtService.GenerateTokenPair(user.ID, roleNames)
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

	roles, err := s.roleRepository.FindRolesByUserID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	roleNames := make([]string, len(roles))
	for i, r := range roles {
		roleNames[i] = r.Name
	}

	tokenPair, err := s.jwtService.GenerateTokenPair(claims.UserID, roleNames)
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

	roles, err := s.roleRepository.FindRolesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	roleNames := make([]string, len(roles))
	for i, r := range roles {
		roleNames[i] = r.Name
	}

	return &UserResponse{
		ID:        user.ID.String(),
		Email:     email,
		Roles:     roleNames,
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

	responses := make([]UserResponse, len(users))
	for i, u := range users {
		email, err := s.cryptoService.DecryptEmail(u.Email)
		if err != nil {
			return nil, 0, err
		}
		//TODO: Fix N+1 query
		roles, err := s.roleRepository.FindRolesByUserID(ctx, u.ID)
		if err != nil {
			return nil, 0, err
		}

		roleNames := make([]string, len(roles))
		for j, r := range roles {
			roleNames[j] = r.Name
		}

		responses[i] = UserResponse{
			ID:        u.ID.String(),
			Email:     email,
			Roles:     roleNames,
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

	roles, err := s.roleRepository.FindRolesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	roleNames := make([]string, len(roles))
	for i, r := range roles {
		roleNames[i] = r.Name
	}

	return &UserResponse{
		ID:        user.ID.String(),
		Email:     email,
		Roles:     roleNames,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *UserService) UpdateUserRoles(ctx context.Context, userID uuid.UUID, roleNames []string) error {
	allRoles, err := s.roleRepository.FindAll(ctx)
	if err != nil {
		return err
	}

	roleMap := make(map[string]int, len(allRoles))
	for _, r := range allRoles {
		roleMap[r.Name] = r.ID
	}

	if err := s.roleRepository.RemoveAllUserRoles(ctx, userID); err != nil {
		return err
	}

	for _, name := range roleNames {
		roleID, ok := roleMap[name]
		if !ok {
			continue
		}
		//TODO: Fix N+1 query using bulk insert
		if err := s.roleRepository.AssignRole(ctx, userID, roleID); err != nil {
			return err
		}
	}

	return nil
}
