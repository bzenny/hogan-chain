package l3_tenant

import (
	"fmt"
	"sync"
	"time"
)

type TenantStatus string

const (
	Active   TenantStatus = "ACTIVE"
	Expired  TenantStatus = "EXPIRED"
	Pending  TenantStatus = "PENDING"
)

type TenantLease struct {
	ProjectID      string
	OwnerAddress   string
	LockedHGXC     float64
	LeaseDuration  time.Duration
	ExpiryTime     time.Time
	Status         TenantStatus
	CustomSubState map[string]string
}

type L3Manager struct {
	mu            sync.Mutex
	ActiveLeases  map[string]*TenantLease
	LeaseCostRate float64 // HGXC cost per block cycle
}

func NewL3Manager() *L3Manager {
	return &L3Manager{
		ActiveLeases:  make(map[string]*TenantLease),
		LeaseCostRate: 100.0, // 100 HGXC per lease activation
	}
}

// RegisterTenantLease locks collateral to lease dedicated L3 execution throughput
func (m *L3Manager) RegisterTenantLease(projectID, owner string, durationHours int) *TenantLease {
	m.mu.Lock()
	defer m.mu.Unlock()

	lease := &TenantLease{
		ProjectID:      projectID,
		OwnerAddress:   owner,
		LockedHGXC:     m.LeaseCostRate,
		LeaseDuration:  time.Duration(durationHours) * time.Hour,
		ExpiryTime:     time.Now().Add(time.Duration(durationHours) * time.Hour),
		Status:         Active,
		CustomSubState: make(map[string]string),
	}

	m.ActiveLeases[projectID] = lease

	fmt.Printf("[L3 TENANT SPACE] Project '%s' registered by %s | Locked: %.2f HGXC | Status: %s\n",
		projectID, owner, lease.LockedHGXC, lease.Status)

	return lease
}

// SetSubStateValue allows an active tenant to store isolated project state data on L3
func (m *L3Manager) SetSubStateValue(projectID, key, value string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	lease, exists := m.ActiveLeases[projectID]
	if !exists || lease.Status != Active {
		fmt.Printf("[L3 REJECTED] Active lease not found for project '%s'\n", projectID)
		return false
	}

	lease.CustomSubState[key] = value
	fmt.Printf("[L3 STATE UPDATE] Project '%s' set [%s = %s]\n", projectID, key, value)
	return true
}
