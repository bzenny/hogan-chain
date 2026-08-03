package l2_business

import (
	"fmt"

	"github.com/yourusername/hogan-chain/pkg/l1_engine"
)

type RWAExecutionEngine struct {
	L1Registry *l1_engine.SPVRegistry
}

func NewRWAExecutionEngine(registry *l1_engine.SPVRegistry) *RWAExecutionEngine {
	return &RWAExecutionEngine{
		L1Registry: registry,
	}
}

// SimulateFractionalization verifies L1 limits before L2/L3 execution
func (e *RWAExecutionEngine) SimulateFractionalization(assetID string, requestedFractionCents l1_engine.CentAmount) bool {
	asset, exists := e.L1Registry.Assets[assetID]
	if !exists {
		fmt.Printf("[L2 RWA REJECTED] Asset %s does not exist on L1\n", assetID)
		return false
	}

	// Calculate max tokenizable value from L1 hard ceiling
	netValue := asset.DeterminedValue - asset.RecognizedLiabilities
	maxTokenizable := (netValue * l1_engine.CentAmount(asset.AuthorizedTokenizationBps)) / 10000

	if requestedFractionCents > maxTokenizable {
		fmt.Printf("[L2 RWA REJECTED] Requested $%.2f exceeds L1 Authorized Tokenization Ceiling ($%.2f)\n",
			float64(requestedFractionCents)/100.0, float64(maxTokenizable)/100.0)
		return false
	}

	fmt.Printf("[L2 RWA APPROVED] Asset '%s' referenced L1 v%d | Fraction $%.2f within L1 cap ($%.2f)\n",
		assetID, asset.ValuationVersion, float64(requestedFractionCents)/100.0, float64(maxTokenizable)/100.0)

	return true
}
