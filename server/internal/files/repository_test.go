package files

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestObjectKey(t *testing.T) {
	require.Equal(t, "files/deadbeef", ObjectKey("deadbeef"))
}

// The FilesRepository itself is exercised via service tests with a mock;
// the repository is a thin SQL wrapper and is verified through manual
// `docker-compose up`-based smoke tests rather than a real-DB test harness.
