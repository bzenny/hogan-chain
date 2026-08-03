package identity_test

import (
	"github.com/yourusername/hogan-chain/internal/eventbus"
	"github.com/yourusername/hogan-chain/internal/identity"
	"github.com/yourusername/hogan-chain/internal/persistence"
	"path/filepath"
	"testing"
)

func TestManagerCanCreateTesterNotPrime(t *testing.T) {
	s, _ := persistence.OpenBolt(filepath.Join(t.TempDir(), "x.db"))
	defer s.Close()
	b := eventbus.New(s)
	svc := identity.NewService(s, b)
	_ = svc.SeedDefaults()
	if err := svc.Create("HGT_mgr", identity.User{ID: "t2", Role: identity.RoleTester}); err != nil {
		t.Fatal(err)
	}
	if err := svc.Create("HGT_mgr", identity.User{ID: "p2", Role: identity.RolePrime}); err == nil {
		t.Fatal("expected role ceiling")
	}
}
