package l1_engine

import (
	"sync"
)

type L1Ledger struct {
	mu           sync.Mutex
	Symbol       string
	MaxSupply    float64
	TotalSupply  float64
	BurnedTokens float64
	Balances     map[string]float64
	BlockHeight  uint64
}

func NewL1Ledger() *L1Ledger {
	return &L1Ledger{
		Symbol:      "HGK",
		MaxSupply:   100000.0,
		TotalSupply: 100000.0,
		Balances: map[string]float64{
			"0x_FOUNDER_TREASURY":    40000.0,
			"0x_L1_L2_BRIDGE_VAULT":  35000.0,
			"0x_L3_TENANT_RESERVE":   15000.0,
			"0x_SYSTEM_SAFETY_BUFFER": 10000.0,
		},
		BlockHeight: 0,
	}
}

func (l *L1Ledger) MineBlock(txCount int) (uint64, float64) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.BlockHeight++
	burnAmount := float64(txCount) * 0.05
	if l.Balances["0x_SYSTEM_SAFETY_BUFFER"] >= burnAmount {
		l.Balances["0x_SYSTEM_SAFETY_BUFFER"] -= burnAmount
		l.BurnedTokens += burnAmount
		l.TotalSupply -= burnAmount
	}

	return l.BlockHeight, l.BurnedTokens
}
