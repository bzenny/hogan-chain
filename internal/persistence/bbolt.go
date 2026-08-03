package persistence

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type diskState struct {
	Buckets map[string]map[string]json.RawMessage `json:"buckets"`
	Events  []json.RawMessage                     `json:"events"`
}

// BoltStore retains the original service name while using a dependency-free,
// atomic JSON persistence file. The interface can later be backed by bbolt
// without changing the application services.
type BoltStore struct {
	mu    sync.RWMutex
	path  string
	state diskState
}

func OpenBolt(path string) (*BoltStore, error) {
	if path == "" {
		return nil, errors.New("database path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	s := &BoltStore{path: path, state: diskState{Buckets: map[string]map[string]json.RawMessage{}, Events: []json.RawMessage{}}}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, s.flushLocked()
		}
		return nil, err
	}
	if len(raw) == 0 {
		return s, nil
	}
	if err := json.Unmarshal(raw, &s.state); err != nil {
		return nil, err
	}
	if s.state.Buckets == nil {
		s.state.Buckets = map[string]map[string]json.RawMessage{}
	}
	return s, nil
}

func (s *BoltStore) flushLocked() error {
	raw, err := json.MarshalIndent(s.state, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *BoltStore) Put(bucket, key string, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state.Buckets[bucket] == nil {
		s.state.Buckets[bucket] = map[string]json.RawMessage{}
	}
	s.state.Buckets[bucket][key] = append(json.RawMessage(nil), raw...)
	return s.flushLocked()
}

func (s *BoltStore) Get(bucket, key string, out any) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b := s.state.Buckets[bucket]
	if b == nil {
		return false, nil
	}
	raw, ok := b[key]
	if !ok {
		return false, nil
	}
	return true, json.Unmarshal(raw, out)
}

func (s *BoltStore) List(bucket string, factory func() any) ([]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b := s.state.Buckets[bucket]
	result := make([]any, 0, len(b))
	for _, raw := range b {
		item := factory()
		if err := json.Unmarshal(raw, item); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *BoltStore) Delete(bucket, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state.Buckets[bucket] != nil {
		delete(s.state.Buckets[bucket], key)
	}
	return s.flushLocked()
}

func (s *BoltStore) AppendEvent(event any) error {
	raw, err := json.Marshal(event)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Events = append(s.state.Events, append(json.RawMessage(nil), raw...))
	return s.flushLocked()
}

func (s *BoltStore) Snapshot(destination string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(s.state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(destination, raw, 0o600)
}

func (s *BoltStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.flushLocked()
}
