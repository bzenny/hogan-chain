package l1_engine

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/yourusername/hogan-chain/internal/eventbus"
	"github.com/yourusername/hogan-chain/internal/persistence"
)

type Units int64
type LedgerState struct {
	Symbol      string           `json:"symbol"`
	Decimals    uint8            `json:"decimals"`
	MaxSupply   Units            `json:"max_supply"`
	TotalSupply Units            `json:"total_supply"`
	Burned      Units            `json:"burned"`
	BlockHeight uint64           `json:"block_height"`
	Balances    map[string]Units `json:"balances"`
	UpdatedAt   time.Time        `json:"updated_at"`
}
type Ledger struct {
	mu    sync.RWMutex
	store persistence.Store
	bus   *eventbus.Bus
	state LedgerState
}

func NewLedger(store persistence.Store, bus *eventbus.Bus, symbol string, decimals uint8, max Units, alloc map[string]int64) (*Ledger, error) {
	l := &Ledger{store: store, bus: bus}
	var existing LedgerState
	if ok, err := store.Get("ledgers", symbol, &existing); err != nil {
		return nil, err
	} else if ok {
		l.state = existing
		return l, nil
	}
	bal := map[string]Units{}
	var total Units
	for k, v := range alloc {
		bal[k] = Units(v)
		total += Units(v)
	}
	if total != max {
		return nil, fmt.Errorf("allocation mismatch for %s", symbol)
	}
	l.state = LedgerState{Symbol: symbol, Decimals: decimals, MaxSupply: max, TotalSupply: max, Balances: bal, UpdatedAt: time.Now().UTC()}
	return l, l.persist()
}
func (l *Ledger) persist() error { return l.store.Put("ledgers", l.state.Symbol, l.state) }
func (l *Ledger) Snapshot() LedgerState {
	l.mu.RLock()
	defer l.mu.RUnlock()
	cp := l.state
	cp.Balances = map[string]Units{}
	for k, v := range l.state.Balances {
		cp.Balances[k] = v
	}
	return cp
}
func (l *Ledger) Transfer(actor, from, to string, amount Units) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.state.Balances[from] < amount {
		return errors.New("insufficient balance")
	}
	l.state.Balances[from] -= amount
	l.state.Balances[to] += amount
	l.state.UpdatedAt = time.Now().UTC()
	if err := l.persist(); err != nil {
		return err
	}
	return l.bus.Publish(persistence.EventRecord{Type: "TokenTransferred", ActorID: actor, Layer: "L1/L2", Resource: l.state.Symbol, Message: fmt.Sprintf("%d units transferred", amount), Data: map[string]any{"from": from, "to": to, "amount": amount}})
}
func (l *Ledger) Burn(actor, from string, amount Units) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.state.Balances[from] < amount {
		return errors.New("insufficient balance")
	}
	l.state.Balances[from] -= amount
	l.state.TotalSupply -= amount
	l.state.Burned += amount
	l.state.UpdatedAt = time.Now().UTC()
	if err := l.persist(); err != nil {
		return err
	}
	return l.bus.Publish(persistence.EventRecord{Type: "FeeBurned", ActorID: actor, Layer: "L1", Resource: l.state.Symbol, Message: fmt.Sprintf("%d units burned", amount)})
}
func (l *Ledger) MineBlock(actor string, txCount int) (uint64, error) {
	l.mu.Lock()
	l.state.BlockHeight++
	l.state.UpdatedAt = time.Now().UTC()
	height := l.state.BlockHeight
	err := l.persist()
	l.mu.Unlock()
	if err != nil {
		return 0, err
	}
	err = l.bus.Publish(persistence.EventRecord{Type: "BlockMined", ActorID: actor, Layer: "L1", Resource: fmt.Sprintf("block-%d", height), Message: fmt.Sprintf("block mined with %d transactions", txCount)})
	return height, err
}
