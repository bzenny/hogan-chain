package bridge

import (
	"errors"
	"github.com/yourusername/hogan-chain/internal/eventbus"
	"github.com/yourusername/hogan-chain/internal/persistence"
	"github.com/yourusername/hogan-chain/pkg/l1_engine"
	"sync"
)

type State struct {
	LockedHGK    l1_engine.Units `json:"locked_hgk"`
	ReleasedHGXC l1_engine.Units `json:"released_hgxc"`
	Ratio        int64           `json:"ratio"`
}
type Service struct {
	mu    sync.Mutex
	store persistence.Store
	bus   *eventbus.Bus
	l1    *l1_engine.Ledger
	l2    *l1_engine.Ledger
	state State
}

func New(store persistence.Store, bus *eventbus.Bus, l1, l2 *l1_engine.Ledger) (*Service, error) {
	s := &Service{store: store, bus: bus, l1: l1, l2: l2, state: State{Ratio: 5}}
	var st State
	if ok, err := store.Get("system", "bridge", &st); err != nil {
		return nil, err
	} else if ok {
		s.state = st
	} else {
		if err := store.Put("system", "bridge", s.state); err != nil {
			return nil, err
		}
	}
	return s, nil
}
func (s *Service) LockAndRelease(actor, user string, hgk l1_engine.Units) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if hgk <= 0 {
		return errors.New("amount must be positive")
	}
	hgxc := hgk * l1_engine.Units(s.state.Ratio)
	if err := s.l1.Transfer(actor, user, "L1_L2_BRIDGE_VAULT", hgk); err != nil {
		return err
	}
	if err := s.l2.Transfer(actor, "L2_BRIDGE_RESERVE", user, hgxc); err != nil {
		return err
	}
	s.state.LockedHGK += hgk
	s.state.ReleasedHGXC += hgxc
	if err := s.store.Put("system", "bridge", s.state); err != nil {
		return err
	}
	return s.bus.Publish(persistence.EventRecord{Type: "BridgeReleased", ActorID: actor, Layer: "L1/L2", Resource: user, Message: "HGK locked and HGXC released"})
}
func (s *Service) Snapshot() State { s.mu.Lock(); defer s.mu.Unlock(); return s.state }
