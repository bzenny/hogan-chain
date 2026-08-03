package app

import (
	"fmt"
	"github.com/yourusername/hogan-chain/internal/approval"
	"github.com/yourusername/hogan-chain/internal/config"
	"github.com/yourusername/hogan-chain/internal/eventbus"
	"github.com/yourusername/hogan-chain/internal/identity"
	"github.com/yourusername/hogan-chain/internal/persistence"
	"github.com/yourusername/hogan-chain/pkg/bridge"
	"github.com/yourusername/hogan-chain/pkg/l1_engine"
	"github.com/yourusername/hogan-chain/pkg/l2_business"
	"github.com/yourusername/hogan-chain/pkg/l3_domain"
	"github.com/yourusername/hogan-chain/pkg/l3_tenant"
	"github.com/yourusername/hogan-chain/pkg/l4_sandbox"
	"path/filepath"
	"time"
)

type Application struct {
	Store     persistence.Store
	Events    *eventbus.Bus
	Identity  *identity.Service
	Approvals *approval.Service
	HGK       *l1_engine.Ledger
	HGXC      *l1_engine.Ledger
	Assets    *l1_engine.SPVRegistry
	Minting   *l1_engine.MintingService
	L2        *l2_business.Engine
	RWA       *l2_business.RWAExecutionEngine
	Domains   *l3_domain.Manager
	Tenants   *l3_tenant.Manager
	Sandbox   *l4_sandbox.Manager
	Bridge    *bridge.Service
	Genesis   *config.Genesis
}

func Build(genesisPath, dbPath string) (*Application, error) {
	g, err := config.LoadGenesis(genesisPath)
	if err != nil {
		return nil, err
	}
	store, err := persistence.OpenBolt(dbPath)
	if err != nil {
		return nil, err
	}
	bus := eventbus.New(store)
	ids := identity.NewService(store, bus)
	if err := ids.SeedDefaults(); err != nil {
		store.Close()
		return nil, err
	}
	hgkc := g.Tokens["hgk"]
	hgxc := g.Tokens["hgxc"]
	hgk, err := l1_engine.NewLedger(store, bus, hgkc.Symbol, hgkc.Decimals, l1_engine.Units(hgkc.MaxSupplyUnits), hgkc.Allocations)
	if err != nil {
		return nil, err
	}
	hgc, err := l1_engine.NewLedger(store, bus, hgxc.Symbol, hgxc.Decimals, l1_engine.Units(hgxc.MaxSupplyUnits), hgxc.Allocations)
	if err != nil {
		return nil, err
	}
	assets := l1_engine.NewSPVRegistry(store, bus)
	mint := l1_engine.NewMintingService(store, bus, assets)
	br, err := bridge.New(store, bus, hgk, hgc)
	if err != nil {
		return nil, err
	}
	a := &Application{Store: store, Events: bus, Identity: ids, Approvals: approval.NewService(store, bus), HGK: hgk, HGXC: hgc, Assets: assets, Minting: mint, L2: l2_business.NewEngine(store, bus, hgc), RWA: l2_business.NewRWAExecutionEngine(assets, mint), Domains: l3_domain.NewManager(store, bus, assets), Tenants: l3_tenant.NewManager(store, bus), Sandbox: l4_sandbox.NewManager(store, bus), Bridge: br, Genesis: g}
	a.registerCoordination()
	return a, nil
}
func (a *Application) registerCoordination() {
	a.Events.Subscribe("ContractActivated", func(e persistence.EventRecord) {
		domain, _ := e.Data["domain_id"].(string)
		tenant, _ := e.Data["tenant_id"].(string)
		rights := []string{}
		if raw, ok := e.Data["rights"].([]string); ok {
			rights = raw
		}
		_ = a.Domains.GrantAccess(e.ActorID, domain, tenant, e.Resource, rights)
	})
}
func (a *Application) Snapshot(actor string) (string, error) {
	name := fmt.Sprintf("hogan-%s.db", time.Now().UTC().Format("20060102-150405"))
	path := filepath.Join("data", "snapshots", name)
	if err := a.Store.Snapshot(path); err != nil {
		return "", err
	}
	_ = a.Events.Publish(persistence.EventRecord{Type: "SnapshotCreated", ActorID: actor, Layer: "SYSTEM", Resource: path, Message: "database snapshot created"})
	return path, nil
}
func (a *Application) Close() error { return a.Store.Close() }
