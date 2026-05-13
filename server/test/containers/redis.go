package containers

import (
	"context"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

// StartRedis boots an ephemeral Redis container and returns an asynq
// RedisClientOpt plus the raw host:port string.
func StartRedis(t *testing.T) (asynq.RedisClientOpt, string) {
	t.Helper()
	ctx := context.Background()

	c, err := tcredis.Run(ctx, "redis:7-alpine")
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Terminate(context.Background()) })

	endpoint, err := c.Endpoint(ctx, "")
	require.NoError(t, err)

	return asynq.RedisClientOpt{Addr: endpoint}, endpoint
}
