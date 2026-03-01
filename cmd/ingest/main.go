package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"realtime-ingestion/internal/db"
	"realtime-ingestion/internal/message"
	"realtime-ingestion/internal/server"
	"realtime-ingestion/internal/simulator"
	"realtime-ingestion/internal/worker"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

var (
	port          = ":8080"
	numSimulators = 5
	numWorkers    = runtime.NumCPU() * 2
	interval      = time.Second * 10
	queueSize     = 10
)

func main() {
	var log *slog.Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	err := godotenv.Load(".env")
	if err != nil {
		log.Warn("could not load .env", "err", err)
	}

	p, ok := os.LookupEnv("SERVER_PORT")
	if ok && p != "" {
		if p[0] == ':' {
			port = p
		} else {
			port = ":" + p
		}
	}

	val, ok := os.LookupEnv("NUM_SIMULATORS")
	if ok {
		n, err := strconv.Atoi(val)
		if err == nil && n > 0 {
			numSimulators = n
		}
	}

	val, ok = os.LookupEnv("NUM_WORKERS")
	if ok {
		n, err := strconv.Atoi(val)
		if err == nil && n > 0 {
			numWorkers = n
		}
	}

	log.Info("service configuration",
		"port", port,
		"num_simulators", numSimulators,
		"num_workers", numWorkers,
		"queue_size", queueSize,
	)

	ctx := context.Background()
	sigCtx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var queue chan message.Data = make(chan message.Data, queueSize)
	var store db.DB = db.NewMessageStore()

	simulatorManger := simulator.NewSimulatorManger(sigCtx, queue)
	simulatorManger.SpawnMultiple(numSimulators, interval)

	workerManger := worker.NewWorkerManager(sigCtx, queue, store)
	workerManger.SpawnMultiple(numWorkers)

	srv := server.NewAPIServer(sigCtx, store, log, port)
	srv.StartServer()

	<-sigCtx.Done()
	log.Info("shutdown signal received")
	simulatorManger.StopAll()
	close(queue)
	workerManger.StopAll()
	srv.StopServer()

	log.Info("exiting the program")
}
