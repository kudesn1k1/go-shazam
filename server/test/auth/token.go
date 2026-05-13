// Package authtest provides helpers for issuing JWTs in integration tests.
// Do not use outside of test code — these helpers panic on signing failure.
package authtest

import (
	"testing"

	"go-shazam/internal/auth"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// MustIssueToken signs and returns an access token for use in integration
// test HTTP requests. Panics if signing fails.
func MustIssueToken(t *testing.T, svc *auth.JWTService, userID uuid.UUID, roles ...string) string {
	t.Helper()
	if len(roles) == 0 {
		roles = []string{"user"}
	}
	pair, err := svc.GenerateTokenPair(userID, roles)
	require.NoError(t, err)
	return pair.AccessToken
}

// MustIssueBearer returns "Bearer <token>" for direct Header.Set.
func MustIssueBearer(t *testing.T, svc *auth.JWTService, userID uuid.UUID, roles ...string) string {
	return "Bearer " + MustIssueToken(t, svc, userID, roles...)
}
