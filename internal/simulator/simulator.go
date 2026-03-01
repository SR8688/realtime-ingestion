package simulator

import (
	"context"
	"log/slog"
	"realtime-ingestion/internal/message"
	"time"
)

const (
	simulatorID = "simID"
)

type Simulator struct {
	id        int
	ctx       context.Context
	size      int
	out       chan message.Data
	interval  time.Duration
	log       *slog.Logger
	createdAt time.Time
}

func (s *Simulator) Start() {
	go s.produceData()
}

func (s *Simulator) produceData() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	defer s.log.Info("simulator stopped", slog.Int(simulatorID, s.id))
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			select {
			case s.out <- message.Data{}:
			case <-s.ctx.Done():
				return
			}
		}
	}
}
