package l3_domain

import (
	"errors"
	"github.com/yourusername/hogan-chain/internal/eventbus"
	"github.com/yourusername/hogan-chain/internal/persistence"
	"github.com/yourusername/hogan-chain/pkg/l1_engine"
	"sort"
	"time"
)

type Status string

const (
	DomainDraft  Status = "DRAFT"
	DomainActive Status = "ACTIVE"
	DomainPaused Status = "PAUSED"
	DomainClosed Status = "CLOSED"
)

type AssetReference struct {
	AssetID          string    `json:"asset_id"`
	SPVID            string    `json:"spv_id"`
	ValuationVersion uint32    `json:"valuation_version"`
	ReferencedAt     time.Time `json:"referenced_at"`
}
type TenantAccess struct {
	TenantID    string    `json:"tenant_id"`
	ContractID  string    `json:"contract_id"`
	Permissions []string  `json:"permissions"`
	GrantedAt   time.Time `json:"granted_at"`
}
type Domain struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Type            string            `json:"type"`
	SubsidiaryID    string            `json:"subsidiary_id"`
	OwnerID         string            `json:"owner_id"`
	Status          Status            `json:"status"`
	ActiveState     map[string]string `json:"active_state"`
	AssetReferences []AssetReference  `json:"asset_references"`
	Access          []TenantAccess    `json:"access"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}
type Manager struct {
	store    persistence.Store
	bus      *eventbus.Bus
	registry *l1_engine.SPVRegistry
}

func NewManager(store persistence.Store, bus *eventbus.Bus, registry *l1_engine.SPVRegistry) *Manager {
	return &Manager{store: store, bus: bus, registry: registry}
}
func (m *Manager) Create(actor string, d Domain) error {
	if d.ID == "" || d.Name == "" {
		return errors.New("domain id and name required")
	}
	d.OwnerID = actor
	d.Status = DomainActive
	d.ActiveState = map[string]string{}
	d.CreatedAt = time.Now().UTC()
	d.UpdatedAt = d.CreatedAt
	if err := m.store.Put("domains", d.ID, d); err != nil {
		return err
	}
	return m.bus.Publish(persistence.EventRecord{Type: "DomainCreated", ActorID: actor, Layer: "L3", Resource: d.ID, Message: d.Name})
}
func (m *Manager) Get(id string) (*Domain, error) {
	var d Domain
	ok, err := m.store.Get("domains", id, &d)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("domain not found")
	}
	return &d, nil
}
func (m *Manager) AddAssetReference(actor, domainID, assetID string) error {
	d, err := m.Get(domainID)
	if err != nil {
		return err
	}
	a, err := m.registry.GetAsset(assetID)
	if err != nil {
		return err
	}
	d.AssetReferences = append(d.AssetReferences, AssetReference{AssetID: a.AssetID, SPVID: a.SPVID, ValuationVersion: a.ValuationVersion, ReferencedAt: time.Now().UTC()})
	d.UpdatedAt = time.Now().UTC()
	if err := m.store.Put("domains", d.ID, d); err != nil {
		return err
	}
	return m.bus.Publish(persistence.EventRecord{Type: "DomainAssetReferenced", ActorID: actor, Layer: "L3", Resource: d.ID, Message: assetID})
}
func (m *Manager) UpdateState(actor, domainID, key, value string) error {
	d, err := m.Get(domainID)
	if err != nil {
		return err
	}
	if d.Status != DomainActive {
		return errors.New("domain not active")
	}
	if d.ActiveState == nil {
		d.ActiveState = map[string]string{}
	}
	d.ActiveState[key] = value
	d.UpdatedAt = time.Now().UTC()
	if err := m.store.Put("domains", d.ID, d); err != nil {
		return err
	}
	return m.bus.Publish(persistence.EventRecord{Type: "DomainStateUpdated", ActorID: actor, Layer: "L3", Resource: d.ID, Message: key + " updated"})
}
func (m *Manager) GrantAccess(actor, domainID, tenantID, contractID string, permissions []string) error {
	d, err := m.Get(domainID)
	if err != nil {
		return err
	}
	d.Access = append(d.Access, TenantAccess{TenantID: tenantID, ContractID: contractID, Permissions: permissions, GrantedAt: time.Now().UTC()})
	d.UpdatedAt = time.Now().UTC()
	if err := m.store.Put("domains", d.ID, d); err != nil {
		return err
	}
	return m.bus.Publish(persistence.EventRecord{Type: "TenantAccessGranted", ActorID: actor, Layer: "L3", Resource: d.ID, Message: tenantID})
}
func (m *Manager) List() ([]Domain, error) {
	items, err := m.store.List("domains", func() any { return &Domain{} })
	if err != nil {
		return nil, err
	}
	out := make([]Domain, 0, len(items))
	for _, x := range items {
		out = append(out, *(x.(*Domain)))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
