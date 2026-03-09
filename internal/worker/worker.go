package worker

import (
	"context"
	"log/slog"
	"realtime-ingestion/internal/db"
	"realtime-ingestion/internal/model"
	"sync/atomic"
	"time"
)

type Worker struct {
	id           int
	ctx          context.Context
	cancel       context.CancelFunc
	in           chan model.Data
	log          *slog.Logger
	createdAt    time.Time
	db           db.DB
	successCount atomic.Int64 //will be sent periodically to observability tools
	failCount    atomic.Int64
}

func (w *Worker) processWork(data model.Data) {
	err := w.db.CreateMessage(w.ctx, data)
	if err != nil {
		w.log.Warn("failed to create message", slog.String("error", err.Error()))
		w.failCount.Add(1)
		return
	}
	w.successCount.Add(1)
}

func (w *Worker) run() {
	for {
		select {
		case data, ok := <-w.in:
			if !ok {
				w.log.Info("worker exiting")
				return
			}
			w.processWork(data)
		case <-w.ctx.Done():
			w.log.Info("worker shutting down")
			return
		}
	}
}
func (w *Worker) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
}
