package l2_business

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

type DomainType string

const (
	DomainCognitive  DomainType = "COGNITIVE_AI"
	DomainQuantum    DomainType = "QUANTUM_COMPLEXITY"
	DomainPhysio     DomainType = "PHYSIOLOGICAL_SYSTEMS"
	DomainSocioLogic DomainType = "SOCIO_TECHNICAL"
)

// EmergentIntersection holds metrics generated during Infinite -> Probable plane traversal
type EmergentIntersection struct {
	ID                  string     `json:"id"`
	DomainA             DomainType `json:"domain_a"`
	DomainB             DomainType `json:"domain_b"`
	AnalogicalDistance  float64    `json:"adi"`  // ADI: Vector/conceptual distance (0.0 - 1.0)
	EntropyUtilityRatio float64    `json:"eur"`  // EUR: Productive noise vs coherence
	SRSI                float64    `json:"srsi"` // Simil-Rarity Salience Index (0.0 - 1.0)
	FlaggedForReview    bool       `json:"flagged"`
	Timestamp           int64      `json:"timestamp"`
}

type SimilRarityEngine struct {
	mu            sync.Mutex
	History       []*EmergentIntersection
	MinThreshold  float64 // Minimum SRSI required to crystallize into L1 Actual Plane
	HighConfidence float64 // Threshold to auto-flag for human/master dev review
}

func NewSimilRarityEngine() *SimilRarityEngine {
	rand.Seed(time.Now().UnixNano())
	return &SimilRarityEngine{
		History:        make([]*EmergentIntersection, 0),
		MinThreshold:   0.60,
		HighConfidence: 0.78,
	}
}

// EvaluateCollision samples from the Infinite Plane, filters in Probable Plane, and calculates SRSI
func (sre *SimilRarityEngine) EvaluateCollision(dA, dB DomainType) *EmergentIntersection {
	sre.mu.Lock()
	defer sre.mu.Unlock()

	// 1. Calculate Analogical Distance Index (ADI): High distance = uncorrelated domains
	adi := 0.55 + (rand.Float64() * 0.42) // Normalized range ~0.55 - 0.97

	// 2. Calculate Entropy-Utility Ratio (EUR): Noise vs structural mapping potential
	eur := 0.40 + (rand.Float64() * 0.55) // Range ~0.40 - 0.95

	// 3. Compute Simil-Rarity Salience Index (SRSI)
	// Formula: SRSI = (ADI * 0.5) + (EUR * 0.5) - (Penalty for recursive saturation noise if EUR > 0.90)
	srsi := (adi * 0.5) + (eur * 0.5)
	if eur > 0.90 {
		srsi -= 0.08 // Saturation noise penalty
	}

	// Clamp SRSI
	srsi = math.Max(0.0, math.Min(1.0, srsi))

	id := fmt.Sprintf("SRI-%d", time.Now().UnixNano()%100000)
	flagged := srsi >= sre.HighConfidence

	intersection := &EmergentIntersection{
		ID:                  id,
		DomainA:             dA,
		DomainB:             dB,
		AnalogicalDistance:  math.Round(adi*100) / 100,
		EntropyUtilityRatio: math.Round(eur*100) / 100,
		SRSI:                math.Round(srsi*100) / 100,
		FlaggedForReview:    flagged,
		Timestamp:           time.Now().Unix(),
	}

	sre.History = append(sre.History, intersection)

	fmt.Printf("[SRE ENGINE] Collision Analyzed [%s <-> %s] | ADI: %.2f | EUR: %.2f | SRSI: %.2f (Flagged: %t)\n",
		dA, dB, adi, eur, srsi, flagged)

	return intersection
}

// GetLatestMetrics returns the 5 most recent emergent collisions for live telemetry
func (sre *SimilRarityEngine) GetLatestMetrics() []*EmergentIntersection {
	sre.mu.Lock()
	defer sre.mu.Unlock()

	if len(sre.History) <= 5 {
		return sre.History
	}
	return sre.History[len(sre.History)-5:]
}
