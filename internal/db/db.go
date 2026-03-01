package db

import (
	"context"
	"fmt"
	"realtime-ingestion/internal/message"
	"sync"
)

type DB interface {
	CreateMessage(ctx context.Context, data message.Data) error
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
