# Realtime Ingestion (Go)
A lightweight concurrent data ingestion system written in Go.
It uses simulators to produce telemetry data, processes it through a worker pool, and exposes HTTP endpoints to query stored results.

## Overview

Simulators → Buffered Channel → Workers → In-Memory Store → HTTP API

* Simulators generate periodic device data usning goroutines.
* Workers consume data from a shared queue and store it.
* MessageStore is a thread-safe in-memory database.
* API Server exposes endpoints to query simulator data.
* Graceful shutdown using context + OS signals.
* Structured logging with slog.

## API Endpoints

* GET /simulators/  
Returns all running simulator IDs.

* GET /simulators/{simulator_id}/data  
Returns all stored data for a simulator.

## Configuration

| Variable  | Default | Description |
| ------------- |:-------------:|--- |
| SERVER_PORT   | 8080          | HTTP server port |
| NUM_SIMULATORS| 5             |Number of simulator goroutines|
| NUM_WORKERS   | 2 * NumCPU()  | Worker pool size |

## Run
```
go run ./cmd/main.go
```