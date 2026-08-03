package l1_engine

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/yourusername/hogan-chain/internal/eventbus"
	"github.com/yourusername/hogan-chain/internal/persistence"
)

type AssetStatus string

const (
	AssetActive     AssetStatus = "ACTIVE"
	AssetRestricted AssetStatus = "RESTRICTED"
	AssetRetired    AssetStatus = "RETIRED"
)

type CentAmount int64
type ValuationVersion struct {
	Version               uint32     `json:"version"`
	DeterminedValue       CentAmount `json:"determined_value_cents"`
	RecognizedLiabilities CentAmount `json:"recognized_liabilities_cents"`
	ApprovedBy            string     `json:"approved_by"`
	CreatedAt             time.Time  `json:"created_at"`
}
type RWAAssetRecord struct {
	AssetID                   string             `json:"asset_id"`
	SPVID                     string             `json:"spv_id"`
	AssetClass                string             `json:"asset_class"`
	Jurisdiction              string             `json:"jurisdiction"`
	Currency                  string             `json:"currency"`
	DeterminedValue           CentAmount         `json:"determined_value_cents"`
	RecognizedLiabilities     CentAmount         `json:"recognized_liabilities_cents"`
	AuthorizedTokenizationBps uint16             `json:"authorized_tokenization_bps"`
	ValuationVersion          uint32             `json:"valuation_version"`
	ControllerAddress         string             `json:"controller_address"`
	Status                    AssetStatus        `json:"status"`
	History                   []ValuationVersion `json:"history"`
	LastUpdated               time.Time          `json:"last_updated"`
}
type SPVRegistry struct {
	store persistence.Store
	bus   *eventbus.Bus
}

func NewSPVRegistry(store persistence.Store, bus *eventbus.Bus) *SPVRegistry {
	return &SPVRegistry{store: store, bus: bus}
}
func (r *SPVRegistry) RegisterAsset(actor string, record RWAAssetRecord) error {
	if record.AssetID == "" || record.SPVID == "" {
		return errors.New("asset id and spv id required")
	}
	if record.DeterminedValue <= 0 {
		return errors.New("determined value must be positive")
	}
	if record.RecognizedLiabilities < 0 || record.RecognizedLiabilities > record.DeterminedValue {
		return errors.New("invalid liabilities")
	}
	if record.AuthorizedTokenizationBps > 10000 {
		return errors.New("tokenization bps cannot exceed 10000")
	}
	var existing RWAAssetRecord
	if ok, _ := r.store.Get("assets", record.AssetID, &existing); ok {
		return fmt.Errorf("asset %s already exists", record.AssetID)
	}
	record.ValuationVersion = 1
	record.Status = AssetActive
	record.Currency = defaultString(record.Currency, "USD")
	record.LastUpdated = time.Now().UTC()
	record.History = []ValuationVersion{{Version: 1, DeterminedValue: record.DeterminedValue, RecognizedLiabilities: record.RecognizedLiabilities, ApprovedBy: actor, CreatedAt: record.LastUpdated}}
	if err := r.store.Put("assets", record.AssetID, record); err != nil {
		return err
	}
	return r.bus.Publish(persistence.EventRecord{Type: "AssetRegistered", ActorID: actor, Layer: "L1", Resource: record.AssetID, Message: "authoritative OCA/RWA asset anchored"})
}
func defaultString(v, d string) string {
	if v == "" {
		return d
	}
	return v
}
func (r *SPVRegistry) GetAsset(id string) (*RWAAssetRecord, error) {
	var a RWAAssetRecord
	ok, err := r.store.Get("assets", id, &a)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("asset not found")
	}
	return &a, nil
}
func (r *SPVRegistry) ListAssets() ([]RWAAssetRecord, error) {
	items, err := r.store.List("assets", func() any { return &RWAAssetRecord{} })
	if err != nil {
		return nil, err
	}
	out := make([]RWAAssetRecord, 0, len(items))
	for _, x := range items {
		out = append(out, *(x.(*RWAAssetRecord)))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].AssetID < out[j].AssetID })
	return out, nil
}
func (r *SPVRegistry) UpdateValuation(actor, assetID string, newValue, newLiabilities CentAmount) error {
	a, err := r.GetAsset(assetID)
	if err != nil {
		return err
	}
	if a.Status == AssetRetired {
		return errors.New("retired asset cannot be revalued")
	}
	if newValue <= 0 || newLiabilities < 0 || newLiabilities > newValue {
		return errors.New("invalid valuation")
	}
	a.ValuationVersion++
	a.DeterminedValue = newValue
	a.RecognizedLiabilities = newLiabilities
	a.LastUpdated = time.Now().UTC()
	a.History = append(a.History, ValuationVersion{Version: a.ValuationVersion, DeterminedValue: newValue, RecognizedLiabilities: newLiabilities, ApprovedBy: actor, CreatedAt: a.LastUpdated})
	if err := r.store.Put("assets", a.AssetID, a); err != nil {
		return err
	}
	return r.bus.Publish(persistence.EventRecord{Type: "ValuationUpdated", ActorID: actor, Layer: "L1", Resource: assetID, Message: fmt.Sprintf("valuation version %d approved", a.ValuationVersion)})
}
