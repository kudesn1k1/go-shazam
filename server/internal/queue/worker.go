package queue

import (
	"context"

	"github.com/hibiken/asynq"
)

type TaskHandler func(ctx context.Context, t *asynq.Task) error

type WorkerServer interface {
	RegisterHandler(pattern string, handler TaskHandler)
	RegisterServiceHandler(pattern string, handler asynq.Handler)
	Run() error
	Start() error
	Stop()
}

type workerServer struct {
	server *asynq.Server
	mux    *asynq.ServeMux
}

func NewWorkerServer(cfg *Config) WorkerServer {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
		},
		asynq.Config{
			Concurrency: 10,
		},
	)

	mux := asynq.NewServeMux()

	return &workerServer{
		server: srv,
		mux:    mux,
	}
}

func (w *workerServer) RegisterHandler(pattern string, handler TaskHandler) {
	w.mux.HandleFunc(pattern, asynq.HandlerFunc(handler))
}

func (w *workerServer) RegisterServiceHandler(pattern string, handler asynq.Handler) {
	w.mux.Handle(pattern, handler)
}

func (w *workerServer) Run() error {
	return w.server.Run(w.mux)
}

func (w *workerServer) Start() error {
	return w.server.Start(w.mux)
}

func (w *workerServer) Stop() {
	w.server.Stop()
	w.server.Shutdown()
}
