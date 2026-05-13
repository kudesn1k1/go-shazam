package main

import (
	"log"
	"net/http"

	"go-shazam/internal/app"
)

func main() {
	// Lightweight healthcheck listener — independent of asynq. If this
	// goroutine is alive, the process is alive. The contract is process
	// liveness, not queue liveness.
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})
		if err := http.ListenAndServe(":5001", mux); err != nil {
			log.Printf("healthz listener exited: %v", err)
		}
	}()

	app.NewWorkerApp().Run()
}
