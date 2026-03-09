package worker

import (
	"context"
	"log/slog"
	"os"
	"realtime-ingestion/internal/db"
	"realtime-ingestion/internal/model"
	"sort"
	"sync"
	"time"
)

type WorkerManager struct {
	ctx      context.Context
	cancel   context.CancelFunc
	workers  map[int]*Worker
	log      *slog.Logger
	in       chan model.Data
	nextID   int
	mu       sync.RWMutex
	wg       sync.WaitGroup
	stopping bool
	db       db.DB
}

func NewWorkerManager(parent context.Context, in chan model.Data, db db.DB) *WorkerManager {
	ctx, cancel := context.WithCancel(parent)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil)).With(slog.String("service", "worker"))
	manager := &WorkerManager{
		ctx:     ctx,
		cancel:  cancel,
		log:     log,
		in:      in,
		workers: make(map[int]*Worker),
		db:      db,
	}
	return manager
}

func (wm *WorkerManager) SpawnMultiple(n int) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if wm.stopping {
		return
	}

	for i := 1; i <= n; i++ {
		wm.nextID += 1
		ctx, cancel := context.WithCancel(wm.ctx)
		log := wm.log.With(slog.Int("worker-id", wm.nextID))
		worker := &Worker{
			id:        wm.nextID,
			ctx:       ctx,
			cancel:    cancel,
			in:        wm.in,
			log:       log,
			createdAt: time.Now(),
			db:        wm.db,
		}
		wm.workers[wm.nextID] = worker
		wm.wg.Add(1)
		go func(worker *Worker) {
			defer wm.wg.Done()
			worker.run()
		}(worker)
	}
}

func (wm *WorkerManager) SpawnOne() {
	wm.SpawnMultiple(1)
}

func (wm *WorkerManager) StopAll() {
	wm.mu.Lock()
	wm.cancel()
	wm.stopping = true
	wm.mu.Unlock()

	wm.wg.Wait()

	wm.mu.Lock()
	wm.workers = make(map[int]*Worker)
	wm.mu.Unlock()
}

func (wm *WorkerManager) StopOne() {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	for id := range wm.workers {
		worker := wm.workers[id]
		worker.Stop()
		delete(wm.workers, id)
		break
	}
}
func (wm *WorkerManager) GetAllWorkerInfo() []model.WorkerInfo {
	wm.mu.RLock()
	snapWorkers := make([]Worker, 0, len(wm.workers))
	for id := range wm.workers {
		snapWorkers = append(snapWorkers, *wm.workers[id])
	}

	wm.mu.RUnlock()

	sort.Slice(snapWorkers, func(i, j int) bool {
		return snapWorkers[i].id < snapWorkers[j].id
	})

	workerInfo := make([]model.WorkerInfo, len(snapWorkers))
	for i := range snapWorkers {
		workerInfo[i] = model.WorkerInfo{
			ID:        snapWorkers[i].id,
			CreatedAt: snapWorkers[i].createdAt,
		}
	}

	return workerInfo
}
