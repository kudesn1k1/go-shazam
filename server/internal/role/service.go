package role

import (
	"context"
	"errors"
	"go-shazam/internal/core/db"

	"github.com/google/uuid"
)

var ErrUnknownRole = errors.New("unknown role")

type Service struct {
	repo               RepositoryInterface
	transactionManager *db.TransactionManager
}

func NewService(repo RepositoryInterface, tm *db.TransactionManager) *Service {
	return &Service{repo: repo, transactionManager: tm}
}

func (s *Service) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]Role, error) {
	entities, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return entitiesToRoles(entities), nil
}

// GetUsersRoles returns a map of userID -> []Role fetched in a single query.
func (s *Service) GetUsersRoles(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID][]Role, error) {
	rows, err := s.repo.FindByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID][]Role, len(userIDs))
	for _, row := range rows {
		result[row.UserID] = append(result[row.UserID], Role(row.Name))
	}
	return result, nil
}

func (s *Service) AssignDefaultRole(ctx context.Context, userID uuid.UUID) error {
	defaultRole, err := s.repo.FindByName(ctx, RoleUser.String())
	if err != nil {
		return err
	}
	return s.repo.AssignRoles(ctx, userID, []int32{int32(defaultRole.ID)})
}

// UpdateUserRoles replaces all roles for a user atomically.
func (s *Service) UpdateUserRoles(ctx context.Context, userID uuid.UUID, roles []Role) error {
	allRoles, err := s.repo.FindAll(ctx)
	if err != nil {
		return err
	}

	roleMap := make(map[Role]int32, len(allRoles))
	for _, r := range allRoles {
		roleMap[Role(r.Name)] = int32(r.ID)
	}

	roleIDs := make([]int32, 0, len(roles))
	for _, name := range roles {
		id, ok := roleMap[name]
		if !ok {
			return ErrUnknownRole
		}
		roleIDs = append(roleIDs, id)
	}

	_, err = db.Transactional(ctx, s.transactionManager, func(txCtx context.Context) (struct{}, error) {
		if err := s.repo.RemoveAllForUser(txCtx, userID); err != nil {
			return struct{}{}, err
		}
		return struct{}{}, s.repo.AssignRoles(txCtx, userID, roleIDs)
	})
	return err
}

func entitiesToRoles(entities []RoleEntity) []Role {
	roles := make([]Role, len(entities))
	for i, e := range entities {
		roles[i] = Role(e.Name)
	}
	return roles
}

func RolesToStrings(roles []Role) []string {
	strs := make([]string, len(roles))
	for i, r := range roles {
		strs[i] = r.String()
	}
	return strs
}
