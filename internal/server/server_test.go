package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"realtime-ingestion/internal/model"
	"realtime-ingestion/internal/simulator"
	"realtime-ingestion/internal/worker"
	"reflect"
	"strings"
	"testing"
)

type fakeDB struct {
	ids []int
	err error
}

func (f *fakeDB) GetAllSimulatorIDs(ctx context.Context) ([]int, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.ids, nil
}
func (f *fakeDB) CreateMessage(ctx context.Context, data model.Data) error {
	return nil
}
func (f *fakeDB) GetAllDataForSimulator(ctx context.Context, simulatorID int) ([]model.Data, error) {
	return nil, nil
}

func TestAPIServer_getAllSimulatorIDs(t *testing.T) {
	tests := []struct {
		name       string
		db         *fakeDB
		wantStatus int
		wantIDs    []int
		wantHeader string
	}{
		{
			name:       "returns ids as json",
			db:         &fakeDB{ids: []int{1, 2, 3}},
			wantStatus: http.StatusOK,
			wantIDs:    []int{1, 2, 3},
			wantHeader: "application/json",
		},
		{
			name:       "db error returns 500",
			db:         &fakeDB{err: errors.New("db failed")},
			wantStatus: http.StatusInternalServerError,
			wantIDs:    nil,
		},
	}
	for _, tt := range tests {
		log := slog.New(slog.NewTextHandler(io.Discard, nil))
		out := make(chan model.Data)
		simulatorManager := simulator.NewSimulatorManger(context.Background(), out)
		workerManager := worker.NewWorkerManager(context.Background(), out, tt.db)
		t.Run(tt.name, func(t *testing.T) {

			a := NewAPIServer(context.Background(), tt.db, log, ":0", simulatorManager, workerManager)
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/simulators/", nil)
			a.getAllSimulatorInfo(rr, req)
			if rr.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d. body=%q", rr.Code, tt.wantStatus, rr.Body.String())
			}

			if tt.wantStatus != http.StatusOK {
				return
			}
			ct := rr.Header().Get("Content-Type")
			if !strings.HasPrefix(ct, tt.wantHeader) {
				t.Fatalf("header = %#v, want %#v", ct, tt.wantHeader)
			}
			var got []int
			if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
				t.Fatalf("invalid json: %v, body=%q", err, rr.Body.String())
			}

			if !reflect.DeepEqual(got, tt.wantIDs) {
				t.Fatalf("ids = %#v, want %#v", got, tt.wantIDs)
			}
		})
	}
}
