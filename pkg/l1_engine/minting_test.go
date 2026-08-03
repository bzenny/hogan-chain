package l1_engine_test

import (
	"github.com/yourusername/hogan-chain/internal/eventbus"
	"github.com/yourusername/hogan-chain/internal/persistence"
	"github.com/yourusername/hogan-chain/pkg/l1_engine"
	"path/filepath"
	"testing"
)

func TestMintCeiling(t *testing.T) {
	s, _ := persistence.OpenBolt(filepath.Join(t.TempDir(), "x.db"))
	defer s.Close()
	b := eventbus.New(s)
	r := l1_engine.NewSPVRegistry(s, b)
	if err := r.RegisterAsset("prime", l1_engine.RWAAssetRecord{AssetID: "A", SPVID: "S", DeterminedValue: 10000, AuthorizedTokenizationBps: 5000}); err != nil {
		t.Fatal(err)
	}
	m := l1_engine.NewMintingService(s, b, r)
	if err := m.Issue("prime", l1_engine.Instrument{ID: "I", AssetID: "A", ValuationVersion: 1, ApprovedExposure: 6000}); err == nil {
		t.Fatal("expected ceiling rejection")
	}
}
