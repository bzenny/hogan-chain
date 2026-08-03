package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type TokenConfig struct {
	Symbol         string           `json:"symbol"`
	Decimals       uint8            `json:"decimals"`
	MaxSupplyUnits int64            `json:"max_supply_units"`
	Allocations    map[string]int64 `json:"allocations"`
}

type Genesis struct {
	ChainID              string                 `json:"chain_id"`
	BlockIntervalSeconds int                    `json:"block_interval_seconds"`
	Tokens               map[string]TokenConfig `json:"tokens"`
}

func LoadGenesis(path string) (*Genesis, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read genesis: %w", err)
	}
	var g Genesis
	if err := json.Unmarshal(raw, &g); err != nil {
		return nil, fmt.Errorf("decode genesis: %w", err)
	}
	if g.ChainID == "" {
		return nil, fmt.Errorf("chain_id is required")
	}
	for name, token := range g.Tokens {
		var total int64
		for _, amount := range token.Allocations {
			if amount < 0 {
				return nil, fmt.Errorf("%s allocation cannot be negative", name)
			}
			total += amount
		}
		if total != token.MaxSupplyUnits {
			return nil, fmt.Errorf("%s allocations %d do not equal max supply %d", name, total, token.MaxSupplyUnits)
		}
	}
	return &g, nil
}
