package l2_business

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/yourusername/hogan-chain/internal/eventbus"
	"github.com/yourusername/hogan-chain/internal/persistence"
	"github.com/yourusername/hogan-chain/pkg/l1_engine"
)

type WorkloadType string

const (
	AIInference    WorkloadType = "AI_INFERENCE"
	MultiAgent     WorkloadType = "MULTI_AGENT"
	QuantumCircuit WorkloadType = "QUANTUM_CIRCUIT"
	DeFiSimulation WorkloadType = "DEFI_SIMULATION"
	RWAAnalysis    WorkloadType = "RWA_ANALYSIS"
	SREAnalysis    WorkloadType = "SRE_ANALYSIS"
)

type TaskStatus string

const (
	TaskDraft           TaskStatus = "DRAFT"
	TaskQueued          TaskStatus = "QUEUED"
	TaskActive          TaskStatus = "ACTIVE"
	TaskWaitingApproval TaskStatus = "WAITING_APPROVAL"
	TaskCompleted       TaskStatus = "COMPLETED"
	TaskFailed          TaskStatus = "FAILED"
	TaskCancelled       TaskStatus = "CANCELLED"
)

type Task struct {
	ID           string          `json:"id"`
	Title        string          `json:"title"`
	Description  string          `json:"description"`
	Type         WorkloadType    `json:"type"`
	Priority     string          `json:"priority"`
	SubsidiaryID string          `json:"subsidiary_id,omitempty"`
	ProgramID    string          `json:"program_id,omitempty"`
	DomainID     string          `json:"domain_id,omitempty"`
	AssetID      string          `json:"asset_id,omitempty"`
	Budget       l1_engine.Units `json:"budget_hgxc_units"`
	Status       TaskStatus      `json:"status"`
	CreatedBy    string          `json:"created_by"`
	Result       string          `json:"result,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}
type Subsidiary struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Purpose    string          `json:"purpose"`
	Status     string          `json:"status"`
	DirectorID string          `json:"director_id,omitempty"`
	Budget     l1_engine.Units `json:"budget_hgxc_units"`
	CreatedBy  string          `json:"created_by"`
	CreatedAt  time.Time       `json:"created_at"`
}
type Program struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Runtime      WorkloadType    `json:"runtime"`
	SubsidiaryID string          `json:"subsidiary_id"`
	Budget       l1_engine.Units `json:"budget_hgxc_units"`
	Status       string          `json:"status"`
	CreatedBy    string          `json:"created_by"`
	CreatedAt    time.Time       `json:"created_at"`
}
type Engine struct {
	store  persistence.Store
	bus    *eventbus.Bus
	ledger *l1_engine.Ledger
	fees   map[WorkloadType]l1_engine.Units
}

func NewEngine(store persistence.Store, bus *eventbus.Bus, ledger *l1_engine.Ledger) *Engine {
	return &Engine{store: store, bus: bus, ledger: ledger, fees: map[WorkloadType]l1_engine.Units{AIInference: 500, MultiAgent: 2500, QuantumCircuit: 25000, DeFiSimulation: 100, RWAAnalysis: 1000, SREAnalysis: 1500}}
}
func (e *Engine) CreateTask(actor string, t Task) error {
	if t.ID == "" || t.Title == "" {
		return errors.New("task id and title required")
	}
	var prior Task
	if ok, _ := e.store.Get("tasks", t.ID, &prior); ok {
		return errors.New("task exists")
	}
	t.CreatedBy = actor
	t.Status = TaskQueued
	t.CreatedAt = time.Now().UTC()
	t.UpdatedAt = t.CreatedAt
	if err := e.store.Put("tasks", t.ID, t); err != nil {
		return err
	}
	return e.bus.Publish(persistence.EventRecord{Type: "TaskCreated", ActorID: actor, Layer: "L2", Resource: t.ID, Message: t.Title})
}
func (e *Engine) ExecuteTask(actor, id, payer string) error {
	var t Task
	ok, err := e.store.Get("tasks", id, &t)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("task not found")
	}
	cost := e.fees[t.Type]
	if cost == 0 {
		return errors.New("unsupported workload")
	}
	if err := e.ledger.Transfer(actor, payer, "HGT_COMPUTE_TREASURY", cost); err != nil {
		return err
	}
	t.Status = TaskCompleted
	t.Result = fmt.Sprintf("%s completed; cost %d HGXC units", t.Type, cost)
	t.UpdatedAt = time.Now().UTC()
	if err := e.store.Put("tasks", id, t); err != nil {
		return err
	}
	return e.bus.Publish(persistence.EventRecord{Type: "TaskCompleted", ActorID: actor, Layer: "L2", Resource: id, Message: t.Result})
}
func (e *Engine) CreateSubsidiary(actor string, s Subsidiary) error {
	if s.ID == "" || s.Name == "" {
		return errors.New("subsidiary id and name required")
	}
	s.CreatedBy = actor
	s.Status = "INTERNAL"
	s.CreatedAt = time.Now().UTC()
	if err := e.store.Put("subsidiaries", s.ID, s); err != nil {
		return err
	}
	return e.bus.Publish(persistence.EventRecord{Type: "SubsidiaryCreated", ActorID: actor, Layer: "L2", Resource: s.ID, Message: s.Name})
}
func (e *Engine) CreateProgram(actor string, p Program) error {
	if p.ID == "" || p.Name == "" {
		return errors.New("program id and name required")
	}
	p.CreatedBy = actor
	p.Status = "ACTIVE"
	p.CreatedAt = time.Now().UTC()
	if err := e.store.Put("programs", p.ID, p); err != nil {
		return err
	}
	return e.bus.Publish(persistence.EventRecord{Type: "ProgramCreated", ActorID: actor, Layer: "L2", Resource: p.ID, Message: p.Name})
}
func listTyped[T any](store persistence.Store, bucket string) ([]T, error) {
	items, err := store.List(bucket, func() any { return new(T) })
	if err != nil {
		return nil, err
	}
	out := make([]T, 0, len(items))
	for _, x := range items {
		out = append(out, *(x.(*T)))
	}
	return out, nil
}
func (e *Engine) ListTasks() ([]Task, error) {
	out, err := listTyped[Task](e.store, "tasks")
	if err == nil {
		sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	}
	return out, err
}
func (e *Engine) ListSubsidiaries() ([]Subsidiary, error) {
	return listTyped[Subsidiary](e.store, "subsidiaries")
}
func (e *Engine) ListPrograms() ([]Program, error) { return listTyped[Program](e.store, "programs") }
