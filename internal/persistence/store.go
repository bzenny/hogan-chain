package persistence

import "time"

type Store interface {
	Put(bucket, key string, value any) error
	Get(bucket, key string, out any) (bool, error)
	List(bucket string, factory func() any) ([]any, error)
	Delete(bucket, key string) error
	AppendEvent(event any) error
	Snapshot(destination string) error
	Close() error
}

type EventRecord struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	ActorID   string         `json:"actor_id"`
	Layer     string         `json:"layer"`
	Resource  string         `json:"resource"`
	Message   string         `json:"message"`
	Data      map[string]any `json:"data,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}
