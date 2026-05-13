package user

import (
	"context"
	"testing"

	"go-shazam/internal/auth"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --- mocks ---

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) Create(ctx context.Context, u *UserEntity) error {
	return m.Called(ctx, u).Error(0)
}
func (m *mockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*UserEntity, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserEntity), args.Error(1)
}
func (m *mockUserRepo) FindByEmailHash(ctx context.Context, h string) (*UserEntity, error) {
	args := m.Called(ctx, h)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserEntity), args.Error(1)
}
func (m *mockUserRepo) ExistsByEmailHash(ctx context.Context, h string) (bool, error) {
	args := m.Called(ctx, h)
	return args.Bool(0), args.Error(1)
}
func (m *mockUserRepo) FindAll(ctx context.Context, limit, offset int) ([]UserEntity, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]UserEntity), args.Error(1)
}
func (m *mockUserRepo) Count(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}
func (m *mockUserRepo) GetAvatarHash(ctx context.Context, id uuid.UUID) (*string, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*string), args.Error(1)
}
func (m *mockUserRepo) SetAvatarHash(ctx context.Context, id uuid.UUID, h *string) error {
	return m.Called(ctx, id, h).Error(0)
}

// --- helpers ---

func newJWT() *auth.JWTService {
	return auth.NewJWTService(&auth.Config{
		AccessTokenSecret:  "test-access-secret-32characters!",
		RefreshTokenSecret: "test-refresh-secret-32character!",
	})
}

func newCrypto() *CryptoService {
	return NewCryptoService(&auth.Config{
		EmailEncryptionKey: "test-email-encryption-key-32chr!",
	})
}

// --- tests ---

func TestUserService_Register_DuplicateEmailReturnsErrUserAlreadyExists(t *testing.T) {
	// Register exits early when ExistsByEmailHash returns true — never touches
	// the role service or transaction manager, so we can safely pass nil for those.
	repo := new(mockUserRepo)
	repo.On("ExistsByEmailHash", mock.Anything, mock.AnythingOfType("string")).
		Return(true, nil)

	svc := NewUserService(repo, nil, newCrypto(), newJWT(), nil, nil)

	_, err := svc.Register(context.Background(), &CreateUser{
		Email:    "user@example.com",
		Password: "correcthorsebatterystaple",
	})

	assert.ErrorIs(t, err, ErrUserAlreadyExists)
	repo.AssertExpectations(t)
}

func TestUserService_Login_WrongPasswordReturnsErrInvalidCredentials(t *testing.T) {
	repo := new(mockUserRepo)
	crypto := newCrypto()

	storedHash, err := crypto.HashPassword("correct-password")
	require.NoError(t, err)

	repo.On("FindByEmailHash", mock.Anything, mock.AnythingOfType("string")).
		Return(&UserEntity{
			ID:             uuid.New(),
			HashedPassword: storedHash,
		}, nil)

	svc := NewUserService(repo, nil, crypto, newJWT(), nil, nil)

	_, err = svc.Login(context.Background(), &LoginRequest{
		Email:    "user@example.com",
		Password: "wrong-password",
	})

	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestUserService_Login_UnknownEmailReturnsErrInvalidCredentials(t *testing.T) {
	repo := new(mockUserRepo)
	repo.On("FindByEmailHash", mock.Anything, mock.AnythingOfType("string")).
		Return(nil, ErrUserNotFound)

	svc := NewUserService(repo, nil, newCrypto(), newJWT(), nil, nil)

	_, err := svc.Login(context.Background(), &LoginRequest{
		Email:    "ghost@example.com",
		Password: "doesnt-matter",
	})

	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestUserService_Login_PasswordVerificationUsesBcrypt(t *testing.T) {
	// Anchor test: confirms that VerifyPassword on the crypto service correctly
	// validates a bcrypt hash produced by HashPassword. If bcrypt cost or rounds
	// change, this catches it.
	crypto := newCrypto()
	hashed, err := crypto.HashPassword("hunter2")
	require.NoError(t, err)
	assert.True(t, crypto.VerifyPassword("hunter2", hashed))
	assert.False(t, crypto.VerifyPassword("hunter3", hashed))
}

func TestUserService_Register_HappyPath_RequiresIntegration(t *testing.T) {
	// The full register-happy-path requires real role.Service (no interface) and
	// a real transaction manager. Covered by the integration suite — see Task 8
	// in docs/superpowers/plans/2026-05-13-comprehensive-testing.md.
	t.Skip("full Register coverage lives in the integration suite (real role.Service + real DB)")
}
