package model

import "time"

type SimulatorInfo struct {
	ID        int       `json:"id,omitempty"`
	Interval  int       `json:"interval,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}
