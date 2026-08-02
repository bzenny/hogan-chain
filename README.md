# Hogan Chain & Half-Gallon Tech Architecture Simulator

A 4-tier, independent (non-forked) blockchain simulation engine written in Go. Built to model state transitions, dual-token economics, cross-layer relaying, and domain-specific workloads—specifically **Simil-Rarity AI/Quantum loops** and **Dual-Ledger Real-World Asset (RWA) verification**.

---

## 🏗 Ecosystem Architecture

The engine implements a 4-tier design mapped directly to the **Three-Plane Model of Cognitive & Emergent Systems**:
+-----------------------------------------------------------------------------------+
|                        HOGAN CHAIN 4-TIER ARCHITECTURE                            |
+-----------------------------------------------------------------------------------+
|  L1: HOGAN CHAIN (HGK)    -->  ACTUAL PLANE (Deterministic Settlement & Fee Burn) |
|  L2: HALF-GALLON TECH     -->  PROBABLE PLANE (AI, Quantum, DeFi & RWA Dual-Ledger)|
|  L3: TENANT NETWORKS      -->  LEASING SPACE (Sub-Projects, SPVs, Eco-Hubs)       |
|  L4: DEV SANDBOX          -->  INFINITE PLANE (High-Entropy Testing & Faucets)    |
+-----------------------------------------------------------------------------------+


### Layer Roles & Economic Dynamics

* **Layer 1: Hogan Chain (`HGK`) — Max Cap: 100,000 HGK**
  * **Role:** Reserve asset, L1 gas engine, and settlement anchor (Actual Plane).
  * **Fee Mechanism:** EIP-1559 style transaction fee burning (0.1% per block), creating measurable deflationary pressure in a tight-supply ecosystem.
* **Layer 2: Half-Gallon Tech (`HGXC`) — Max Cap: 500,000 HGXC**
  * **Role:** Operational utility token ($1 \text{ HGK} : 5 \text{ HGXC}$ relative ratio).
  * **Execution Engine:** Powers metered workloads including AI inference passes, Quantum circuit composition, DeFi AMM swaps, and Dual-Ledger RWA verification proofs.
* **Layer 3: Tenant & Project Space**
  * **Role:** Dedicated execution and sub-state storage for leased sub-projects (e.g., waste recovery hubs like TerraCycle/TCSi, real estate SPVs). Leased by locking collateral.
* **Layer 4: Ephemeral Dev Sandbox**
  * **Role:** High-entropy testing space (Infinite Plane). Faucet test credits expire periodically to prevent test noise from cluttering L1–L3 state history.

---

## 📊 Genesis Tokenomics Allocation (`config/genesis.json`)

| Allocation Vault | HGK (L1 Engine) | HGXC (L2 Utility) | Operational Function |
| :--- | :--- | :--- | :--- |
| **Founder / Compute Treasury** | 40,000 (40%) | 200,000 (40%) | Strategic reserve; funds L2 compute nodes and R&D. |
| **L1 $\leftrightarrow$ L2 Bridge Vault** | 35,000 (35%) | 175,000 (35%) | Cross-chain collateral pool for atomic lock-and-mint transfers. |
| **L3 Tenant Leasing Reserve** | 15,000 (15%) | 75,000 (15%) | Collateral buffer for incoming sub-project onboarding. |
| **L4 Sandbox Faucet** | 10,000 (10%) | 50,000 (10%) | Ephemeral test credit distribution pool. |
| **Total Genesis Supply** | **100,000 HGK** | **500,000 HGXC** | **Hard-capped in genesis configuration.** |

---

## 📂 Codebase Layout

```text
hogan-chain/
├── config/
│   └── genesis.json       <-- Hardcoded supply caps (100k HGK / 500k HGXC) & initial vaults
├── pkg/
│   ├── l1_engine/
│   │   └── ledger.go      <-- L1 HGK state, block mining, and fee-burning mechanics
│   ├── l2_business/
│   │   ├── engine.go      <-- L2 HGXC execution: AI, Quantum, DeFi & Dual-Ledger RWA
│   │   └── sre.go         <-- Simil-Rarity Engine: ADI, EUR, and SRSI emergence logic
│   ├── l3_tenant/
│   │   └── lease.go       <-- Tenant sub-state rentals & lease locking
│   ├── l4_sandbox/
│   │   └── faucet.go      <-- Ephemeral dev space & temporary test credits
│   └── bridge/
│       └── relayer.go     <-- L1 <-> L2 cross-layer lock-and-mint relayer
├── cmd/
│   └── node/
│       └── main.go        <-- Entry point: boots network, processes jobs, renders CLI & Web dashboard
├── web/
│   └── templates/
│       └── index.html     <-- Multi-role Web UI (Master Dev, Prime User, Test User)
├── .gitignore             <-- Ignores binaries, logs, and OS files
├── go.mod                 <-- Go module definition
└── README.md              <-- System documentation & run guide
