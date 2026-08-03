package l3_tenant

import (
	"errors"
	"github.com/yourusername/hogan-chain/internal/eventbus"
	"github.com/yourusername/hogan-chain/internal/persistence"
	"sort"
	"time"
)

type TenantStatus string

const (
	TenantActive     TenantStatus = "ACTIVE"
	TenantSuspended  TenantStatus = "SUSPENDED"
	TenantTerminated TenantStatus = "TERMINATED"
)

type ContractStatus string

const (
	ContractDraft      ContractStatus = "DRAFT"
	ContractActive     ContractStatus = "ACTIVE"
	ContractSuspended  ContractStatus = "SUSPENDED"
	ContractCompleted  ContractStatus = "COMPLETED"
	ContractDefaulted  ContractStatus = "DEFAULTED"
	ContractTerminated ContractStatus = "TERMINATED"
)

type Tenant struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Controller string       `json:"controller"`
	Status     TenantStatus `json:"status"`
	CreatedBy  string       `json:"created_by"`
	CreatedAt  time.Time    `json:"created_at"`
}
type Obligation struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Metric      string    `json:"metric"`
	TargetValue int64     `json:"target_value"`
	Unit        string    `json:"unit"`
	DueAt       time.Time `json:"due_at"`
	Status      string    `json:"status"`
}
type Contract struct {
	ID              string         `json:"id"`
	TenantID        string         `json:"tenant_id"`
	DomainID        string         `json:"domain_id"`
	Type            string         `json:"type"`
	Status          ContractStatus `json:"status"`
	Rights          []string       `json:"rights"`
	Obligations     []Obligation   `json:"obligations"`
	PaymentTerms    string         `json:"payment_terms"`
	CollateralToken string         `json:"collateral_token"`
	CollateralUnits int64          `json:"collateral_units"`
	TermsHash       string         `json:"terms_hash"`
	EffectiveAt     time.Time      `json:"effective_at"`
	ExpiresAt       time.Time      `json:"expires_at"`
	CreatedBy       string         `json:"created_by"`
	CreatedAt       time.Time      `json:"created_at"`
}
type Manager struct {
	store persistence.Store
	bus   *eventbus.Bus
}

func NewManager(store persistence.Store, bus *eventbus.Bus) *Manager {
	return &Manager{store: store, bus: bus}
}
func (m *Manager) Register(actor string, t Tenant) error {
	if t.ID == "" || t.Name == "" {
		return errors.New("tenant id and name required")
	}
	t.Status = TenantActive
	t.CreatedBy = actor
	t.CreatedAt = time.Now().UTC()
	if err := m.store.Put("tenants", t.ID, t); err != nil {
		return err
	}
	return m.bus.Publish(persistence.EventRecord{Type: "TenantRegistered", ActorID: actor, Layer: "L3", Resource: t.ID, Message: t.Name})
}
func (m *Manager) CreateContract(actor string, c Contract) error {
	if c.ID == "" || c.TenantID == "" || c.DomainID == "" {
		return errors.New("contract, tenant and domain ids required")
	}
	var t Tenant
	if ok, err := m.store.Get("tenants", c.TenantID, &t); err != nil {
		return err
	} else if !ok {
		return errors.New("tenant not found")
	}
	c.Status = ContractDraft
	c.CreatedBy = actor
	c.CreatedAt = time.Now().UTC()
	if err := m.store.Put("contracts", c.ID, c); err != nil {
		return err
	}
	return m.bus.Publish(persistence.EventRecord{Type: "ContractCreated", ActorID: actor, Layer: "L3", Resource: c.ID, Message: c.Type})
}
func (m *Manager) ActivateContract(actor, id string) error {
	var c Contract
	ok, err := m.store.Get("contracts", id, &c)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("contract not found")
	}
	if c.Status != ContractDraft {
		return errors.New("contract not in draft")
	}
	c.Status = ContractActive
	if c.EffectiveAt.IsZero() {
		c.EffectiveAt = time.Now().UTC()
	}
	if err := m.store.Put("contracts", id, c); err != nil {
		return err
	}
	return m.bus.Publish(persistence.EventRecord{Type: "ContractActivated", ActorID: actor, Layer: "L3", Resource: id, Message: "tenant contract activated", Data: map[string]any{"domain_id": c.DomainID, "tenant_id": c.TenantID, "rights": c.Rights}})
}
func list[T any](s persistence.Store, b string) ([]T, error) {
	items, err := s.List(b, func() any { return new(T) })
	if err != nil {
		return nil, err
	}
	out := make([]T, 0, len(items))
	for _, x := range items {
		out = append(out, *(x.(*T)))
	}
	return out, nil
}
func (m *Manager) ListTenants() ([]Tenant, error) {
	out, err := list[Tenant](m.store, "tenants")
	if err == nil {
		sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	}
	return out, err
}
func (m *Manager) ListContracts() ([]Contract, error) { return list[Contract](m.store, "contracts") }
