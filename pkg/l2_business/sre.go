package l2_business

import (
	"errors"
	"fmt"

	"github.com/yourusername/hogan-chain/pkg/l1_engine"
)

type RWAExecutionEngine struct {
	L1Registry *l1_engine.SPVRegistry
	Minting    *l1_engine.MintingService
}

func NewRWAExecutionEngine(registry *l1_engine.SPVRegistry, minting *l1_engine.MintingService) *RWAExecutionEngine {
	return &RWAExecutionEngine{L1Registry: registry, Minting: minting}
}
func (e *RWAExecutionEngine) SimulateFractionalization(assetID string, requested l1_engine.CentAmount) (l1_engine.CentAmount, error) {
	asset, err := e.L1Registry.GetAsset(assetID)
	if err != nil {
		return 0, err
	}
	if asset.Status != l1_engine.AssetActive {
		return 0, errors.New("asset is not active")
	}
	net := asset.DeterminedValue - asset.RecognizedLiabilities
	ceiling := (net * l1_engine.CentAmount(asset.AuthorizedTokenizationBps)) / 10000
	existing, err := e.Minting.TotalActiveExposure(assetID)
	if err != nil {
		return 0, err
	}
	available := ceiling - existing
	if requested <= 0 || requested > available {
		return available, fmt.Errorf("requested exposure exceeds available ceiling")
	}
	return available, nil
}

type SREInput struct {
	ADI               float64
	EUR               float64
	SaturationPenalty float64
}

func ScoreSRE(in SREInput) (float64, bool) {
	score := in.ADI*.5 + in.EUR*.5 - in.SaturationPenalty
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}
	return score, score >= .78
}
