package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"realtime-ingestion/internal/db"
	"strconv"
	"time"
)

type APIServer struct {
	ctx  context.Context
	db   db.DB
	log  *slog.Logger
	port string
	srv  *http.Server
}

func NewAPIServer(ctx context.Context, db db.DB, log *slog.Logger, port string) *APIServer {
	server := &APIServer{
		ctx:  ctx,
		db:   db,
		log:  log.With(slog.String("service", "api")),
		port: port,
	}
	return server
}

func (a *APIServer) StartServer() {
	router := http.NewServeMux()

	simulatorRouter := http.NewServeMux()
	simulatorRouter.HandleFunc("GET /", a.getAllSimulatorIDs)
	simulatorRouter.HandleFunc("GET /{simulator_id}/data", a.getAllDataForSimulator)

	router.Handle("/simulators/", http.StripPrefix("/simulators", simulatorRouter)) //important how we do the stripping with one less slash

	srv := &http.Server{
		Addr:         a.port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	a.srv = srv

	go func() {
		info := fmt.Sprintf("starting http server at %s", a.port)
		a.log.Info(info)
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.log.Error("http server error", "err", err)
		}
	}()
}
func (a *APIServer) StopServer() {
	if a.srv == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.srv.Shutdown(ctx); err != nil {
		a.log.Error("error shutting down server", "err", err)
	}
}

func (a *APIServer) getAllSimulatorIDs(w http.ResponseWriter, r *http.Request) {
	ids, err := a.db.GetAllSimulatorIDs(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ids)
}

func (a *APIServer) getAllDataForSimulator(w http.ResponseWriter, r *http.Request) {
	simulatorID := r.PathValue("simulator_id")
	id, err := strconv.Atoi(simulatorID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data, err := a.db.GetAllDataForSimulator(r.Context(), int(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
