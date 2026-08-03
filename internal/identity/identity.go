package identity

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/yourusername/hogan-chain/internal/eventbus"
	"github.com/yourusername/hogan-chain/internal/persistence"
)

type Role string

const (
	RolePrime              Role = "HOGANCHAIN_PRIME"
	RoleManager            Role = "HGT_MANAGER"
	RoleSubsidiaryDirector Role = "SUBSIDIARY_DIRECTOR"
	RoleProgramManager     Role = "PROGRAM_MANAGER"
	RoleDomainOperator     Role = "DOMAIN_OPERATOR"
	RoleTester             Role = "TESTER"
)

type Status string

const (
	StatusActive    Status = "ACTIVE"
	StatusSuspended Status = "SUSPENDED"
)

type Scope struct {
	Type       string `json:"type"`
	ResourceID string `json:"resource_id"`
}
type User struct {
	ID          string     `json:"id"`
	DisplayName string     `json:"display_name"`
	Role        Role       `json:"role"`
	Status      Status     `json:"status"`
	CreatedBy   string     `json:"created_by"`
	ReportsTo   string     `json:"reports_to"`
	Permissions []string   `json:"permissions"`
	Scopes      []Scope    `json:"scopes"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

type Service struct {
	store persistence.Store
	bus   *eventbus.Bus
}

func NewService(store persistence.Store, bus *eventbus.Bus) *Service {
	return &Service{store: store, bus: bus}
}

var rolePermissions = map[Role][]string{
	RolePrime:              {"*"},
	RoleManager:            {"task.*", "program.*", "subsidiary.*", "domain.*", "tenant.*", "contract.*", "sandbox.*", "proposal.create", "user.delegate"},
	RoleSubsidiaryDirector: {"task.*", "program.*", "domain.*", "tenant.*", "contract.*", "sandbox.project.review", "user.delegate"},
	RoleProgramManager:     {"task.*", "program.manage", "sandbox.project.review", "tester.create"},
	RoleDomainOperator:     {"domain.state.update", "domain.telemetry.record", "domain.workflow.manage", "tenant.performance.record"},
	RoleTester:             {"sandbox.*", "test_account.*", "proposal.submit_to_manager"},
}

func Defaults() []User {
	now := time.Now().UTC()
	return []User{
		{ID: "HoganChain_prime", DisplayName: "HoganChain Prime", Role: RolePrime, Status: StatusActive, Permissions: rolePermissions[RolePrime], CreatedAt: now},
		{ID: "HGT_mgr", DisplayName: "Half-Gallon Tech Manager", Role: RoleManager, Status: StatusActive, CreatedBy: "HoganChain_prime", ReportsTo: "HoganChain_prime", Permissions: rolePermissions[RoleManager], CreatedAt: now},
		{ID: "test_1a", DisplayName: "Tester 1A", Role: RoleTester, Status: StatusActive, CreatedBy: "HGT_mgr", ReportsTo: "HGT_mgr", Permissions: rolePermissions[RoleTester], Scopes: []Scope{{Type: "sandbox", ResourceID: "default"}}, CreatedAt: now},
	}
}
func (s *Service) SeedDefaults() error {
	for _, u := range Defaults() {
		var existing User
		ok, err := s.store.Get("users", u.ID, &existing)
		if err != nil {
			return err
		}
		if !ok {
			if err := s.store.Put("users", u.ID, u); err != nil {
				return err
			}
		}
	}
	return nil
}
func (s *Service) Get(id string) (*User, error) {
	var u User
	ok, err := s.store.Get("users", id, &u)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("user not found")
	}
	return &u, nil
}
func (s *Service) List() ([]User, error) {
	items, err := s.store.List("users", func() any { return &User{} })
	if err != nil {
		return nil, err
	}
	out := make([]User, 0, len(items))
	for _, it := range items {
		out = append(out, *(it.(*User)))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
func matches(granted, required string) bool {
	if granted == "*" || granted == required {
		return true
	}
	if len(granted) > 2 && granted[len(granted)-2:] == ".*" {
		p := granted[:len(granted)-1]
		return len(required) >= len(p) && required[:len(p)] == p
	}
	return false
}
func (s *Service) Authorize(userID, permission string) error {
	u, err := s.Get(userID)
	if err != nil {
		return err
	}
	if u.Status != StatusActive {
		return errors.New("identity is not active")
	}
	if u.ExpiresAt != nil && time.Now().After(*u.ExpiresAt) {
		return errors.New("identity expired")
	}
	for _, p := range u.Permissions {
		if matches(p, permission) {
			return nil
		}
	}
	return fmt.Errorf("%s lacks %s", userID, permission)
}
func canCreate(parent, child Role) bool {
	if parent == RolePrime {
		return true
	}
	if parent == RoleManager {
		return child == RoleSubsidiaryDirector || child == RoleProgramManager || child == RoleDomainOperator || child == RoleTester
	}
	if parent == RoleSubsidiaryDirector {
		return child == RoleProgramManager || child == RoleDomainOperator || child == RoleTester
	}
	if parent == RoleProgramManager {
		return child == RoleTester
	}
	return false
}
func (s *Service) Create(actorID string, u User) error {
	actor, err := s.Get(actorID)
	if err != nil {
		return err
	}
	if !canCreate(actor.Role, u.Role) {
		return fmt.Errorf("%s cannot create role %s", actor.Role, u.Role)
	}
	var existing User
	if ok, _ := s.store.Get("users", u.ID, &existing); ok {
		return errors.New("user already exists")
	}
	u.CreatedBy = actorID
	u.ReportsTo = actorID
	u.Status = StatusActive
	u.CreatedAt = time.Now().UTC()
	if len(u.Permissions) == 0 {
		u.Permissions = append([]string(nil), rolePermissions[u.Role]...)
	}
	if err := s.store.Put("users", u.ID, u); err != nil {
		return err
	}
	return s.bus.Publish(persistence.EventRecord{Type: "UserCreated", ActorID: actorID, Layer: "SYSTEM", Resource: u.ID, Message: "delegated identity created"})
}
