package model

import "time"

type WorkerInfo struct {
	ID           int       `json:"id,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	SuccessCount int64     `json:"success_count,omitempty"`
	FailCount    int64     `json:"fail_count,omitempty"`
}
