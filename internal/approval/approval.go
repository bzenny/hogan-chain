package approval

import (
	"errors"
	"sort"
	"time"

	"github.com/yourusername/hogan-chain/internal/eventbus"
	"github.com/yourusername/hogan-chain/internal/persistence"
)

type Status string

const (
	Draft       Status = "DRAFT"
	Submitted   Status = "SUBMITTED"
	UnderReview Status = "UNDER_REVIEW"
	Approved    Status = "APPROVED"
	Rejected    Status = "REJECTED"
	Cancelled   Status = "CANCELLED"
	Executed    Status = "EXECUTED"
)

type Proposal struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Title        string         `json:"title"`
	RequestedBy  string         `json:"requested_by"`
	RequiredRole string         `json:"required_role"`
	ResourceID   string         `json:"resource_id"`
	Payload      map[string]any `json:"payload"`
	Status       Status         `json:"status"`
	DecisionNote string         `json:"decision_note,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}
type Service struct {
	store persistence.Store
	bus   *eventbus.Bus
}

func NewService(store persistence.Store, bus *eventbus.Bus) *Service {
	return &Service{store: store, bus: bus}
}
func (s *Service) Create(p Proposal) error {
	if p.ID == "" || p.Type == "" || p.RequestedBy == "" {
		return errors.New("proposal id, type, and requester required")
	}
	p.Status = Submitted
	p.CreatedAt = time.Now().UTC()
	p.UpdatedAt = p.CreatedAt
	if err := s.store.Put("proposals", p.ID, p); err != nil {
		return err
	}
	return s.bus.Publish(persistence.EventRecord{Type: "ProposalSubmitted", ActorID: p.RequestedBy, Layer: "SYSTEM", Resource: p.ID, Message: p.Title})
}
func (s *Service) Get(id string) (*Proposal, error) {
	var p Proposal
	ok, err := s.store.Get("proposals", id, &p)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("proposal not found")
	}
	return &p, nil
}
func (s *Service) Decide(id, actor string, approve bool, note string) error {
	p, err := s.Get(id)
	if err != nil {
		return err
	}
	if p.Status != Submitted && p.Status != UnderReview {
		return errors.New("proposal is not reviewable")
	}
	if approve {
		p.Status = Approved
	} else {
		p.Status = Rejected
	}
	p.DecisionNote = note
	p.UpdatedAt = time.Now().UTC()
	if err := s.store.Put("proposals", p.ID, p); err != nil {
		return err
	}
	typ := "ProposalRejected"
	if approve {
		typ = "ProposalApproved"
	}
	return s.bus.Publish(persistence.EventRecord{Type: typ, ActorID: actor, Layer: "SYSTEM", Resource: p.ID, Message: note})
}
func (s *Service) MarkExecuted(id, actor string) error {
	p, err := s.Get(id)
	if err != nil {
		return err
	}
	if p.Status != Approved {
		return errors.New("proposal is not approved")
	}
	p.Status = Executed
	p.UpdatedAt = time.Now().UTC()
	if err := s.store.Put("proposals", id, p); err != nil {
		return err
	}
	return s.bus.Publish(persistence.EventRecord{Type: "ProposalExecuted", ActorID: actor, Layer: "SYSTEM", Resource: id, Message: p.Title})
}
func (s *Service) List() ([]Proposal, error) {
	items, err := s.store.List("proposals", func() any { return &Proposal{} })
	if err != nil {
		return nil, err
	}
	out := make([]Proposal, 0, len(items))
	for _, x := range items {
		out = append(out, *(x.(*Proposal)))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}
