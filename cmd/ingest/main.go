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
	"syscall"
	"time"
)

func main() {

	portInfo := ":8080"

	var log *slog.Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()
	sigCtx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var queue chan message.Data = make(chan message.Data, 10)
	var db db.DB = db.NewMessageStore()

	simulatorManger := simulator.NewSimulatorManger(sigCtx, queue)
	simulatorManger.SpawnMultiple(3, time.Second*10)

	workerManger := worker.NewWorkerManager(sigCtx, queue, db)
	workerManger.SpawnMultiple(5)

	srv := server.NewAPIServer(sigCtx, db, log, portInfo)
	srv.StartServer()

	<-sigCtx.Done()
	simulatorManger.StopAll()
	close(queue)
	workerManger.StopAll()
	srv.StopServer()

	log.Info("exiting the program")
}
