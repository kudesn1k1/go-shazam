package files

import (
	"context"
	"fmt"
	"time"

	"go-shazam/internal/core/db"
	"go-shazam/internal/queue"

	"github.com/hibiken/asynq"
	"go.uber.org/fx"
)

const (
	FileCleanupTaskType = "files:cleanup_temporary"
	cleanupBatchSize    = 500
)

// txRunner wraps db.Transactional so the cleanup handler can be unit-tested
// without constructing a real *sqlx.DB.
type txRunner func(ctx context.Context, fn func(ctx context.Context) error) error

type CleanupTaskHandler struct {
	repo  FilesRepositoryInterface
	s3    S3Client
	cfg   *Config
	runTx txRunner
}

func NewCleanupTaskHandler(repo FilesRepositoryInterface, s3 S3Client, cfg *Config, tm *db.TransactionManager) *CleanupTaskHandler {
	return &CleanupTaskHandler{
		repo: repo,
		s3:   s3,
		cfg:  cfg,
		runTx: func(ctx context.Context, fn func(ctx context.Context) error) error {
			_, err := db.Transactional(ctx, tm, func(txCtx context.Context) (struct{}, error) {
				return struct{}{}, fn(txCtx)
			})
			return err
		},
	}
}

func (h *CleanupTaskHandler) ProcessTask(ctx context.Context, _ *asynq.Task) error {
	cutoff := time.Now().UTC().Add(-h.cfg.CleanupOutdate)

	for {
		hashes, err := h.repo.ListExpired(ctx, cutoff, cleanupBatchSize)
		if err != nil {
			return fmt.Errorf("list expired: %w", err)
		}
		if len(hashes) == 0 {
			return nil
		}

		for _, hash := range hashes {
			if err := h.deleteOne(ctx, hash); err != nil {
				// Log and move on; a transient DB or S3 issue on one hash
				// shouldn't abort the whole sweep.
				fmt.Printf("[files-cleanup] delete %s: %v\n", hash, err)
				continue
			}
		}

		if len(hashes) < cleanupBatchSize {
			return nil
		}
	}
}

// deleteOne removes a single TEMPORARY file atomically across DB and S3.
// The DB row is deleted inside a transaction; the S3 object is deleted only
// after the DB delete succeeds. If S3 deletion fails, the transaction is
// rolled back and the row stays, so the next cleanup run retries cleanly.
func (h *CleanupTaskHandler) deleteOne(ctx context.Context, hash string) error {
	return h.runTx(ctx, func(txCtx context.Context) error {
		if err := h.repo.DeleteByHash(txCtx, hash); err != nil {
			return fmt.Errorf("db delete: %w", err)
		}
		if err := h.s3.DeleteObject(ctx, ObjectKey(hash)); err != nil {
			return fmt.Errorf("s3 delete: %w", err)
		}
		return nil
	})
}

func RegisterCleanupHandler(w queue.WorkerServer, h *CleanupTaskHandler) {
	w.RegisterServiceHandler(FileCleanupTaskType, h)
	fmt.Printf("[Queue] Registering handler for task type: %s\n", FileCleanupTaskType)
}

func RegisterCleanupSchedule(lc fx.Lifecycle, qcfg *queue.Config, fcfg *Config) error {
	provider := &staticPeriodicConfigProvider{
		configs: []*asynq.PeriodicTaskConfig{
			{
				Cronspec: fcfg.CleanupSchedule,
				Task:     asynq.NewTask(FileCleanupTaskType, nil),
			},
		},
	}

	mgr, err := asynq.NewPeriodicTaskManager(asynq.PeriodicTaskManagerOpts{
		RedisConnOpt: asynq.RedisClientOpt{
			Addr:     qcfg.Addr,
			Password: qcfg.Password,
			DB:       qcfg.DB,
		},
		PeriodicTaskConfigProvider: provider,
		SyncInterval:               5 * time.Minute,
	})
	if err != nil {
		return fmt.Errorf("new periodic manager: %w", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			fmt.Printf("[Queue] Starting periodic task manager (cleanup: %s)\n", fcfg.CleanupSchedule)
			return mgr.Start()
		},
		OnStop: func(context.Context) error {
			mgr.Shutdown()
			return nil
		},
	})

	return nil
}

type staticPeriodicConfigProvider struct {
	configs []*asynq.PeriodicTaskConfig
}

func (p *staticPeriodicConfigProvider) GetConfigs() ([]*asynq.PeriodicTaskConfig, error) {
	return p.configs, nil
}
