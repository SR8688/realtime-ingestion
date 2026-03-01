package simulator

import (
	"context"
	"log/slog"
	"os"
	"realtime-ingestion/internal/message"
	"sync"
	"time"
)

type SimulatorManager struct {
	ctx        context.Context
	cancel     context.CancelFunc
	simulators map[int]*Simulator
	log        *slog.Logger
	out        chan message.Data
	nextID     int
	mu         sync.RWMutex
	wg         sync.WaitGroup
	stopping   bool
}

func NewSimulatorManger(parent context.Context, out chan message.Data) *SimulatorManager {
	ctx, cancel := context.WithCancel(parent)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil)).With(slog.String("service", "simulator"))
	manager := &SimulatorManager{
		ctx:        ctx,
		cancel:     cancel,
		log:        log,
		out:        out,
		simulators: make(map[int]*Simulator),
	}
	return manager
}
func (sm *SimulatorManager) SpawnMultiple(n int, interval time.Duration) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.stopping {
		return
	}

	for i := 1; i <= n; i++ {
		sm.nextID += 1
		ctx, cancel := context.WithCancel(sm.ctx)
		simulator := &Simulator{
			id:        sm.nextID,
			ctx:       ctx,
			cancel:    cancel,
			out:       sm.out,
			interval:  interval,
			log:       sm.log,
			createdAt: time.Now(),
		}
		sm.simulators[sm.nextID] = simulator
		sm.wg.Add(1)
		go func(simulator *Simulator) {
			defer sm.wg.Done()
			simulator.produceData()
		}(simulator)
	}
}

func (sm *SimulatorManager) SpawnOne(interval time.Duration) {
	sm.SpawnMultiple(1, interval)
}
func (sm *SimulatorManager) StopAll() {
	sm.mu.Lock()
	sm.cancel()
	sm.stopping = true
	sm.mu.Unlock()

	sm.wg.Wait()

	sm.mu.Lock()
	sm.simulators = make(map[int]*Simulator)
	sm.mu.Unlock()
}

func (sm *SimulatorManager) StopOne() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	for id := range sm.simulators {
		simulator := sm.simulators[id]
		simulator.Stop()
		delete(sm.simulators, id)
		break
	}
}
