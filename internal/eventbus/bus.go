package eventbus

import (
	"fmt"
	"sync"
	"time"

	"github.com/yourusername/hogan-chain/internal/persistence"
)

type Handler func(persistence.EventRecord)

type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
	store    persistence.Store
}

func New(store persistence.Store) *Bus { return &Bus{handlers: map[string][]Handler{}, store: store} }
func (b *Bus) Subscribe(eventType string, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], h)
}
func (b *Bus) Publish(e persistence.EventRecord) error {
	if e.ID == "" {
		e.ID = fmt.Sprintf("evt-%d", time.Now().UnixNano())
	}
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now().UTC()
	}
	if err := b.store.AppendEvent(e); err != nil {
		return err
	}
	b.mu.RLock()
	hs := append([]Handler(nil), b.handlers[e.Type]...)
	b.mu.RUnlock()
	for _, h := range hs {
		h(e)
	}
	return nil
}
