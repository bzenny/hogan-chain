package l4_sandbox

import (
	"errors"
	"github.com/yourusername/hogan-chain/internal/eventbus"
	"github.com/yourusername/hogan-chain/internal/persistence"
	"sort"
	"time"
)

type ProjectStatus string

const (
	ProjectDraft     ProjectStatus = "DRAFT"
	ProjectTesting   ProjectStatus = "TESTING"
	ProjectSubmitted ProjectStatus = "SUBMITTED"
	ProjectArchived  ProjectStatus = "ARCHIVED"
)

type TestAccount struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"owner_id"`
	THGK      int64     `json:"thgk_units"`
	THGXC     int64     `json:"thgxc_units"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
type Project struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Type        string        `json:"type"`
	Description string        `json:"description"`
	OwnerID     string        `json:"owner_id"`
	Status      ProjectStatus `json:"status"`
	Version     string        `json:"version"`
	CreatedAt   time.Time     `json:"created_at"`
	ExpiresAt   time.Time     `json:"expires_at"`
}
type TestRun struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Scenario  string    `json:"scenario"`
	Expected  string    `json:"expected"`
	Actual    string    `json:"actual"`
	Passed    bool      `json:"passed"`
	CreatedAt time.Time `json:"created_at"`
}
type Manager struct {
	store persistence.Store
	bus   *eventbus.Bus
}

func NewManager(store persistence.Store, bus *eventbus.Bus) *Manager {
	return &Manager{store: store, bus: bus}
}
func (m *Manager) CreateAccount(actor string, a TestAccount) error {
	if a.ID == "" {
		return errors.New("account id required")
	}
	a.OwnerID = actor
	a.THGK = 100000
	a.THGXC = 500000
	if a.ExpiresAt.IsZero() {
		a.ExpiresAt = time.Now().Add(24 * time.Hour)
	}
	a.CreatedAt = time.Now().UTC()
	if err := m.store.Put("test_accounts", a.ID, a); err != nil {
		return err
	}
	return m.bus.Publish(persistence.EventRecord{Type: "TestAccountCreated", ActorID: actor, Layer: "L4", Resource: a.ID, Message: "sandbox account funded"})
}
func (m *Manager) CreateProject(actor string, p Project) error {
	if p.ID == "" || p.Name == "" {
		return errors.New("project id and name required")
	}
	p.OwnerID = actor
	p.Status = ProjectTesting
	p.CreatedAt = time.Now().UTC()
	if p.ExpiresAt.IsZero() {
		p.ExpiresAt = time.Now().Add(30 * 24 * time.Hour)
	}
	if err := m.store.Put("test_projects", p.ID, p); err != nil {
		return err
	}
	return m.bus.Publish(persistence.EventRecord{Type: "TestProjectCreated", ActorID: actor, Layer: "L4", Resource: p.ID, Message: p.Name})
}
func (m *Manager) RecordRun(actor string, r TestRun) error {
	if r.ID == "" || r.ProjectID == "" {
		return errors.New("run and project ids required")
	}
	r.CreatedAt = time.Now().UTC()
	if err := m.store.Put("test_runs", r.ID, r); err != nil {
		return err
	}
	return m.bus.Publish(persistence.EventRecord{Type: "TestRunRecorded", ActorID: actor, Layer: "L4", Resource: r.ID, Message: r.Scenario})
}
func (m *Manager) SubmitProject(actor, id string) error {
	var p Project
	ok, err := m.store.Get("test_projects", id, &p)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("project not found")
	}
	p.Status = ProjectSubmitted
	if err := m.store.Put("test_projects", id, p); err != nil {
		return err
	}
	return m.bus.Publish(persistence.EventRecord{Type: "TestProjectSubmitted", ActorID: actor, Layer: "L4", Resource: id, Message: "submitted to HGT manager"})
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
func (m *Manager) ListProjects() ([]Project, error) {
	out, err := list[Project](m.store, "test_projects")
	if err == nil {
		sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	}
	return out, err
}
func (m *Manager) ListAccounts() ([]TestAccount, error) {
	return list[TestAccount](m.store, "test_accounts")
}
