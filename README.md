# Hogan Chain & Half-Gallon Tech System Manager

A persistent, single-user systems laboratory written in Go. It models a four-tier blockchain-style ecosystem without pretending to be a production public chain.

## State hierarchy

- **L1 — Actual Plane:** HGK ledger, authoritative SPV/OCA/RWA records, valuation history, tokenization ceilings, approvals, and asset-backed instruments.
- **L2 — Probable Plane:** HGXC workloads, tasks, programs, subsidiaries, AI/SRE/quantum/RWA analysis, and proposals.
- **L3 — Operating Plane:** domains hold active operating state; tenants hold contracts, rights, obligations, payments, collateral, and performance requirements.
- **L4 — Infinite Plane:** test identities, disposable accounts, dApp experiments, scenarios, and persistent test results.

## Default identities

- `HoganChain_prime` — system manager and L1 authority.
- `HGT_mgr` — Half-Gallon Tech director for L2/L3 programs, subsidiaries, domains, tenants, and delegated users.
- `test_1a` — L4 tester for dApps, test accounts, and sandbox programs.

The dashboard identity selector is intentionally lightweight. Server-side permission and delegation rules still apply, and every mutation is persisted and journaled.

## Run

```bash
go mod tidy
go test ./...
go run ./cmd/node
```

Open `http://localhost:8080`.

Use another port or data file when needed:

```bash
PORT=8088 DB_PATH=data/dev.db go run ./cmd/node
```

## Persistence

`config/genesis.json` initializes token state once. After initialization, `data/hogan-chain.db` is authoritative. Snapshots are written to `data/snapshots/`.

## Minting rules

1. HGK and HGXC are issued at genesis only.
2. The bridge locks HGK and releases existing HGXC reserve; it does not inflate supply.
3. Asset-backed instruments are separate instruments tied to an L1 asset and valuation version.
4. Existing exposure plus requested exposure cannot exceed the L1 tokenization ceiling.
5. Stale valuation references are rejected.

## HTMX rules

- Go owns validation, authorization, persistence, and HTML fragments.
- HTMX performs partial-page mutations and refreshes.
- Destructive or authoritative actions require confirmation.
- Minimal JavaScript is used only for presentation behavior.

## Development note

Replace `github.com/yourusername/hogan-chain` in `go.mod` and imports with your final GitHub repository path before publishing.
