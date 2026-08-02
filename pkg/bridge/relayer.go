package bridge

import (
	"errors"
	"fmt"
	"sync"

	"github.com/yourusername/hogan-chain/pkg/l1_engine"
	"github.com/yourusername/hogan-chain/pkg/l2_business"
)

type BridgeRelayer struct {
	mu           sync.Mutex
	L1           *l1_engine.L1Ledger
	L2           *l2_business.L2Engine
	Ratio        float64 // 1 HGK = 5 HGXC
	LockedHGK    float64
	MintedHGXC   float64
}

func NewBridgeRelayer(l1 *l1_engine.L1Ledger, l2 *l2_business.L2Engine) *BridgeRelayer {
	return &BridgeRelayer{
		L1:         l1,
		L2:         l2,
		Ratio:      5.0, // 1:5 exchange ratio
		LockedHGK:  0.0,
		MintedHGXC: 0.0,
	}
}

// LockL1AndMintL2 takes HGK on L1 and issues equivalent HGXC on L2
func (b *BridgeRelayer) LockL1AndMintL2(user string, hgkAmount float64) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if hgkAmount <= 0 {
		return errors.New("bridge amount must be greater than zero")
	}

	// Verify L1 balance
	b.L1.Mu.Lock()
	if b.L1.Balances[user] < hgkAmount {
		b.L1.Mu.Unlock()
		return fmt.Errorf("insufficient L1 HGK balance for user %s", user)
	}

	// Lock on L1
	b.L1.Balances[user] -= hgkAmount
	b.L1.Balances["0x_L1_L2_BRIDGE_VAULT"] += hgkAmount
	b.L1.Mu.Unlock()

	// Mint equivalent on L2
	hgxcToMint := hgkAmount * b.Ratio

	b.L2.Mu.Lock()
	b.L2.Balances[user] += hgxcToMint
	b.L2.Balances["0x_BRIDGE_MINT_VAULT"] -= hgxcToMint
	b.L2.Mu.Unlock()

	b.LockedHGK += hgkAmount
	b.MintedHGXC += hgxcToMint

	fmt.Printf("[BRIDGE RELAYER] Locked %.2f HGK on L1 ---> Minted %.2f HGXC on L2 for %s\n",
		hgkAmount, hgxcToMint, user)

	return nil
}

// BurnL2AndUnlockL1 burns L2 HGXC and releases locked L1 HGK
func (b *BridgeRelayer) BurnL2AndUnlockL1(user string, hgxcAmount float64) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if hgxcAmount <= 0 {
		return errors.New("bridge amount must be greater than zero")
	}

	// Verify L2 balance
	b.L2.Mu.Lock()
	if b.L2.Balances[user] < hgxcAmount {
		b.L2.Mu.Unlock()
		return fmt.Errorf("insufficient L2 HGXC balance for user %s", user)
	}

	// Burn on L2
	b.L2.Balances[user] -= hgxcAmount
	b.L2.Balances["0x_BRIDGE_MINT_VAULT"] += hgxcAmount
	b.L2.Mu.Unlock()

	// Unlock on L1
	hgkToUnlock := hgxcAmount / b.Ratio

	b.L1.Mu.Lock()
	b.L1.Balances["0x_L1_L2_BRIDGE_VAULT"] -= hgkToUnlock
	b.L1.Balances[user] += hgkToUnlock
	b.L1.Mu.Unlock()

	b.LockedHGK -= hgkToUnlock
	b.MintedHGXC -= hgxcAmount

	fmt.Printf("[BRIDGE RELAYER] Burned %.2f HGXC on L2 ---> Unlocked %.2f HGK on L1 for %s\n",
		hgxcAmount, hgkToUnlock, user)

	return nil
}
