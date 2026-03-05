package db

import (
	"context"
	"fmt"
	"realtime-ingestion/internal/message"
	"sort"
	"sync"
)

type DB interface {
	CreateMessage(ctx context.Context, data message.Data) error
	GetAllSimulatorIDs(ctx context.Context) ([]int, error)
	GetAllDataForSimulator(ctx context.Context, simulatorID int) ([]message.Data, error)
}

type MessageStore struct {
	data     map[int][]message.Data
	mu       sync.RWMutex
	uniqueID map[int]map[string]struct{}
}

func NewMessageStore() *MessageStore {
	store := &MessageStore{
		data:     make(map[int][]message.Data),
		uniqueID: make(map[int]map[string]struct{}),
	}
	return store
}

func (ms *MessageStore) CreateMessage(ctx context.Context, data message.Data) error {

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.uniqueID[data.DeviceID] == nil {
		ms.uniqueID[data.DeviceID] = make(map[string]struct{})
	}
	_, exists := ms.uniqueID[data.DeviceID][data.ID]
	if exists {
		return fmt.Errorf("message with id %s already exists", data.ID)
	}

	ms.uniqueID[data.DeviceID][data.ID] = struct{}{}
	ms.data[data.DeviceID] = append(ms.data[data.DeviceID], data)

	return nil

}
func (ms *MessageStore) GetAllSimulatorIDs(ctx context.Context) ([]int, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	simulatorIDs := make([]int, len(ms.data))
	i := 0
	for id := range ms.data {
		simulatorIDs[i] = id
		i++
	}

	sort.Ints(simulatorIDs)

	return simulatorIDs, nil
}

func (ms *MessageStore) GetAllDataForSimulator(ctx context.Context, simulatorID int) ([]message.Data, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	ms.mu.RLock()
	defer ms.mu.RUnlock()
	data, exists := ms.data[simulatorID]
	if !exists {
		return nil, fmt.Errorf("simulator with id - %d does not exist", simulatorID)
	}

	result := make([]message.Data, len(data))
	copy(result, data)

	return result, nil
}
