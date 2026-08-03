package l1_engine_test

import (
	"github.com/yourusername/hogan-chain/internal/eventbus"
	"github.com/yourusername/hogan-chain/internal/persistence"
	"github.com/yourusername/hogan-chain/pkg/l1_engine"
	"path/filepath"
	"testing"
)

func TestLedgerPersistsTransfer(t *testing.T) {
	s, err := persistence.OpenBolt(filepath.Join(t.TempDir(), "x.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	b := eventbus.New(s)
	l, err := l1_engine.NewLedger(s, b, "T", 0, 100, map[string]int64{"a": 100})
	if err != nil {
		t.Fatal(err)
	}
	if err := l.Transfer("x", "a", "b", 25); err != nil {
		t.Fatal(err)
	}
	if got := l.Snapshot().Balances["b"]; got != 25 {
		t.Fatalf("got %d", got)
	}
}
