package l1_engine

import (
	"errors"
	"sort"
	"time"

	"github.com/yourusername/hogan-chain/internal/eventbus"
	"github.com/yourusername/hogan-chain/internal/persistence"
)

type InstrumentStatus string

const (
	InstrumentActive    InstrumentStatus = "ACTIVE"
	InstrumentSuspended InstrumentStatus = "SUSPENDED"
	InstrumentRetired   InstrumentStatus = "RETIRED"
)

type Instrument struct {
	ID               string           `json:"id"`
	AssetID          string           `json:"asset_id"`
	SPVID            string           `json:"spv_id"`
	ValuationVersion uint32           `json:"valuation_version"`
	ApprovedExposure CentAmount       `json:"approved_exposure_cents"`
	IssuedExposure   CentAmount       `json:"issued_exposure_cents"`
	Symbol           string           `json:"symbol"`
	Status           InstrumentStatus `json:"status"`
	ApprovedBy       string           `json:"approved_by"`
	CreatedAt        time.Time        `json:"created_at"`
}
type MintingService struct {
	store    persistence.Store
	bus      *eventbus.Bus
	registry *SPVRegistry
}

func NewMintingService(store persistence.Store, bus *eventbus.Bus, registry *SPVRegistry) *MintingService {
	return &MintingService{store: store, bus: bus, registry: registry}
}
func (m *MintingService) TotalActiveExposure(assetID string) (CentAmount, error) {
	items, err := m.store.List("instruments", func() any { return &Instrument{} })
	if err != nil {
		return 0, err
	}
	var total CentAmount
	for _, x := range items {
		i := x.(*Instrument)
		if i.AssetID == assetID && i.Status != InstrumentRetired {
			total += i.IssuedExposure
		}
	}
	return total, nil
}
func (m *MintingService) Issue(actor string, i Instrument) error {
	a, err := m.registry.GetAsset(i.AssetID)
	if err != nil {
		return err
	}
	if a.Status != AssetActive {
		return errors.New("asset is not active")
	}
	if i.ValuationVersion != a.ValuationVersion {
		return errors.New("proposal valuation version is stale")
	}
	net := a.DeterminedValue - a.RecognizedLiabilities
	ceiling := (net * CentAmount(a.AuthorizedTokenizationBps)) / 10000
	existing, err := m.TotalActiveExposure(a.AssetID)
	if err != nil {
		return err
	}
	if i.ApprovedExposure <= 0 || existing+i.ApprovedExposure > ceiling {
		return errors.New("issuance exceeds available tokenization capacity")
	}
	var prior Instrument
	if ok, _ := m.store.Get("instruments", i.ID, &prior); ok {
		return errors.New("instrument already exists")
	}
	i.SPVID = a.SPVID
	i.IssuedExposure = i.ApprovedExposure
	i.Status = InstrumentActive
	i.ApprovedBy = actor
	i.CreatedAt = time.Now().UTC()
	if err := m.store.Put("instruments", i.ID, i); err != nil {
		return err
	}
	return m.bus.Publish(persistence.EventRecord{Type: "InstrumentIssued", ActorID: actor, Layer: "L1", Resource: i.ID, Message: "asset-backed instrument issued"})
}
func (m *MintingService) List() ([]Instrument, error) {
	items, err := m.store.List("instruments", func() any { return &Instrument{} })
	if err != nil {
		return nil, err
	}
	out := make([]Instrument, 0, len(items))
	for _, x := range items {
		out = append(out, *(x.(*Instrument)))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}
