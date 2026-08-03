package l1_engine

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type AssetStatus string

const (
	AssetActive     AssetStatus = "ACTIVE"
	AssetRestricted AssetStatus = "RESTRICTED"
	AssetRetired    AssetStatus = "RETIRED"
)

// CentAmount uses int64 to avoid float precision issues across state transitions
type CentAmount int64

type RWAAssetRecord struct {
	AssetID                   string      `json:"asset_id"`
	SPVID                     string      `json:"spv_id"`
	AssetClass                string      `json:"asset_class"`
	Jurisdiction              string      `json:"jurisdiction"`
	DeterminedValue           CentAmount  `json:"determined_value_cents"`
	RecognizedLiabilities     CentAmount  `json:"recognized_liabilities_cents"`
	AuthorizedTokenizationBps uint16      `json:"authorized_tokenization_bps"` // Basis points (e.g., 6000 = 60%)
	ValuationVersion          uint32      `json:"valuation_version"`
	ControllerAddress         string      `json:"controller_address"`
	Status                    AssetStatus `json:"status"`
	LastUpdated               int64       `json:"last_updated"`
}

type SPVRegistry struct {
	mu     sync.RWMutex
	Assets map[string]*RWAAssetRecord
}

func NewSPVRegistry() *SPVRegistry {
	return &SPVRegistry{
		Assets: make(map[string]*RWAAssetRecord),
	}
}

// RegisterAsset anchors a new SPV or RWA on Layer 1
func (r *SPVRegistry) RegisterAsset(record *RWAAssetRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.Assets[record.AssetID]; exists {
		return fmt.Errorf("asset %s already registered on L1", record.AssetID)
	}

	record.ValuationVersion = 1
	record.Status = AssetActive
	record.LastUpdated = time.Now().Unix()
	r.Assets[record.AssetID] = record

	fmt.Printf("[L1 SPV REGISTRY] Anchored Asset '%s' (SPV: %s) | Value: $%.2f | Ceiling: %d bps\n",
		record.AssetID, record.SPVID, float64(record.DeterminedValue)/100.0, record.AuthorizedTokenizationBps)

	return nil
}

// UpdateValuation increments the versioned valuation on L1
func (r *SPVRegistry) UpdateValuation(assetID string, newValue, newLiabilities CentAmount) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	asset, exists := r.Assets[assetID]
	if !exists {
		return errors.New("asset not found on L1")
	}

	asset.DeterminedValue = newValue
	asset.RecognizedLiabilities = newLiabilities
	asset.ValuationVersion++
	asset.LastUpdated = time.Now().Unix()

	fmt.Printf("[L1 SPV REGISTRY] Valuation Updated '%s' -> Version %d | Net Value: $%.2f\n",
		assetID, asset.ValuationVersion, float64(newValue-newLiabilities)/100.0)

	return nil
}
