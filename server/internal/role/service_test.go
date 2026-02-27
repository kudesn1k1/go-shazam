package role

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a testify mock for RepositoryInterface.
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]RoleEntity, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]RoleEntity), args.Error(1)
}

func (m *MockRepository) FindByUserIDs(ctx context.Context, userIDs []uuid.UUID) ([]UserRoleRow, error) {
	args := m.Called(ctx, userIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]UserRoleRow), args.Error(1)
}

func (m *MockRepository) FindByName(ctx context.Context, name string) (*RoleEntity, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*RoleEntity), args.Error(1)
}

func (m *MockRepository) FindAll(ctx context.Context) ([]RoleEntity, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]RoleEntity), args.Error(1)
}

func (m *MockRepository) AssignRoles(ctx context.Context, userID uuid.UUID, roleIDs []int32) error {
	args := m.Called(ctx, userID, roleIDs)
	return args.Error(0)
}

func (m *MockRepository) RemoveAllForUser(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// serviceWithoutTx creates a Service with nil TransactionManager for tests that don't
// exercise UpdateUserRoles (which requires a real DB transaction).
func serviceWithoutTx(repo RepositoryInterface) *Service {
	return &Service{repo: repo, transactionManager: nil}
}

func TestRoleService_GetUserRoles_Success(t *testing.T) {
	repo := new(MockRepository)
	userID := uuid.New()

	repo.On("FindByUserID", mock.Anything, userID).Return([]RoleEntity{
		{ID: 1, Name: "admin"},
		{ID: 2, Name: "user"},
	}, nil)

	svc := serviceWithoutTx(repo)
	roles, err := svc.GetUserRoles(context.Background(), userID)

	assert.NoError(t, err)
	assert.Equal(t, []Role{RoleAdmin, RoleUser}, roles)
	repo.AssertExpectations(t)
}

func TestRoleService_GetUserRoles_Empty(t *testing.T) {
	repo := new(MockRepository)
	userID := uuid.New()

	repo.On("FindByUserID", mock.Anything, userID).Return([]RoleEntity{}, nil)

	svc := serviceWithoutTx(repo)
	roles, err := svc.GetUserRoles(context.Background(), userID)

	assert.NoError(t, err)
	assert.Empty(t, roles)
	repo.AssertExpectations(t)
}

func TestRoleService_GetUserRoles_RepositoryError(t *testing.T) {
	repo := new(MockRepository)
	userID := uuid.New()
	repoErr := errors.New("db error")

	repo.On("FindByUserID", mock.Anything, userID).Return(nil, repoErr)

	svc := serviceWithoutTx(repo)
	roles, err := svc.GetUserRoles(context.Background(), userID)

	assert.ErrorIs(t, err, repoErr)
	assert.Nil(t, roles)
	repo.AssertExpectations(t)
}

func TestRoleService_GetUsersRoles_BatchSuccess(t *testing.T) {
	repo := new(MockRepository)
	id1 := uuid.New()
	id2 := uuid.New()
	userIDs := []uuid.UUID{id1, id2}

	repo.On("FindByUserIDs", mock.Anything, userIDs).Return([]UserRoleRow{
		{UserID: id1, RoleID: 1, Name: "admin"},
		{UserID: id1, RoleID: 2, Name: "user"},
		{UserID: id2, RoleID: 2, Name: "user"},
	}, nil)

	svc := serviceWithoutTx(repo)
	rolesMap, err := svc.GetUsersRoles(context.Background(), userIDs)

	assert.NoError(t, err)
	assert.Equal(t, []Role{RoleAdmin, RoleUser}, rolesMap[id1])
	assert.Equal(t, []Role{RoleUser}, rolesMap[id2])
	repo.AssertExpectations(t)
}

func TestRoleService_GetUsersRoles_EmptyInput(t *testing.T) {
	repo := new(MockRepository)

	repo.On("FindByUserIDs", mock.Anything, []uuid.UUID{}).Return(nil, nil)

	svc := serviceWithoutTx(repo)
	rolesMap, err := svc.GetUsersRoles(context.Background(), []uuid.UUID{})

	assert.NoError(t, err)
	assert.Empty(t, rolesMap)
	repo.AssertExpectations(t)
}

func TestRoleService_AssignDefaultRole_Success(t *testing.T) {
	repo := new(MockRepository)
	userID := uuid.New()

	repo.On("FindByName", mock.Anything, RoleUser.String()).Return(&RoleEntity{ID: 2, Name: "user"}, nil)
	repo.On("AssignRoles", mock.Anything, userID, []int32{2}).Return(nil)

	svc := serviceWithoutTx(repo)
	err := svc.AssignDefaultRole(context.Background(), userID)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestRoleService_AssignDefaultRole_RoleNotFound(t *testing.T) {
	repo := new(MockRepository)
	userID := uuid.New()
	repoErr := errors.New("role not found")

	repo.On("FindByName", mock.Anything, RoleUser.String()).Return(nil, repoErr)

	svc := serviceWithoutTx(repo)
	err := svc.AssignDefaultRole(context.Background(), userID)

	assert.ErrorIs(t, err, repoErr)
	repo.AssertExpectations(t)
}

func TestRoleService_UpdateUserRoles_UnknownRole(t *testing.T) {
	repo := new(MockRepository)
	userID := uuid.New()

	repo.On("FindAll", mock.Anything).Return([]RoleEntity{
		{ID: 1, Name: "admin"},
		{ID: 2, Name: "user"},
	}, nil)

	svc := serviceWithoutTx(repo)
	err := svc.UpdateUserRoles(context.Background(), userID, []Role{"superadmin"})

	assert.ErrorIs(t, err, ErrUnknownRole)
	repo.AssertExpectations(t)
}

func TestRoleService_UpdateUserRoles_FindAllError(t *testing.T) {
	repo := new(MockRepository)
	userID := uuid.New()
	repoErr := errors.New("db error")

	repo.On("FindAll", mock.Anything).Return(nil, repoErr)

	svc := serviceWithoutTx(repo)
	err := svc.UpdateUserRoles(context.Background(), userID, []Role{RoleAdmin})

	assert.ErrorIs(t, err, repoErr)
	repo.AssertExpectations(t)
}

func TestRoleService_RolesToStrings(t *testing.T) {
	roles := []Role{RoleAdmin, RoleUser}
	strs := RolesToStrings(roles)
	assert.Equal(t, []string{"admin", "user"}, strs)
}

func TestRoleService_RoleString(t *testing.T) {
	assert.Equal(t, "admin", RoleAdmin.String())
	assert.Equal(t, "user", RoleUser.String())
}
