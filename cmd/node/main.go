package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/yourusername/hogan-chain/pkg/bridge"
	"github.com/yourusername/hogan-chain/pkg/l1_engine"
	"github.com/yourusername/hogan-chain/pkg/l2_business"
	"github.com/yourusername/hogan-chain/pkg/l3_tenant"
)

var (
	l1     *l1_engine.L1Ledger
	l2     *l2_business.L2Engine
	relayer *bridge.BridgeRelayer
	l3     *l3_tenant.L3Manager
)

func main() {
	fmt.Println("==================================================")
	fmt.Println("   HOGAN CHAIN & HALF-GALLON TECH NODE BOOTING   ")
	fmt.Println("==================================================")

	// Initialize System Components
	l1 = l1_engine.NewL1Ledger()
	l2 = l2_business.NewL2Engine()
	relayer = bridge.NewBridgeRelayer(l1, l2)
	l3 = l3_tenant.NewL3Manager()

	// Register Default L3 Tenant space for testing
	l3.RegisterTenantLease("TCSi_RECOVERY_HUB", "0x_USER_DEMO", 24)
	l3.SetSubStateValue("TCSi_RECOVERY_HUB", "DUAL_LEDGER_STATUS", "VERIFIED")

	// Set HTTP Endpoints
	http.HandleFunc("/", serveDashboard)
	http.HandleFunc("/api/action", handleAction)

	fmt.Println("\n[NODE ONLINE] Live Web Dashboard running at: http://localhost:8080")
	fmt.Println("Press Ctrl+C to stop.")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func serveDashboard(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/index.html")
	if err != nil {
		http.Error(w, "Dashboard template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func handleAction(w http.ResponseWriter, r *http.Request) {
	role := r.URL.Query().Get("role")
	action := r.URL.Query().Get("action")

	w.Header().Set("Content-Type", "application/json")
	var response map[string]interface{}

	switch action {
	case "mine":
		height, burned := l1.MineBlock(1)
		response = map[string]interface{}{
			"status":       "ok",
			"message":      fmt.Sprintf("Block #%d mined by %s.", height, role),
			"block_height": height,
			"burned":       burned,
		}
	case "bridge":
		err := relayer.LockL1AndMintL2("0x_FOUNDER_TREASURY", 10.0)
		if err != nil {
			response = map[string]interface{}{"status": "error", "message": err.Error()}
		} else {
			response = map[string]interface{}{
				"status":  "ok",
				"message": "Locked 10 HGK on L1 ---> Minted 50 HGXC on L2 via Bridge Relayer.",
			}
		}
	case "ai_job":
		success := l2.ExecuteWorkload("0x_USER_DEMO", l2_business.AI_Inference)
		height, burned := l1.MineBlock(1)
		response = map[string]interface{}{
			"status":       "ok",
			"message":      fmt.Sprintf("Executed AI Workload (Success: %t)", success),
			"block_height": height,
			"burned":       burned,
		}
	case "quantum_job":
		success := l2.ExecuteWorkload("0x_USER_DEMO", l2_business.Quantum_Circuit)
		height, burned := l1.MineBlock(1)
		response = map[string]interface{}{
			"status":       "ok",
			"message":      fmt.Sprintf("Executed Quantum Circuit Workload (Success: %t)", success),
			"block_height": height,
			"burned":       burned,
		}
	case "rwa_proof":
		success := l2.ExecuteWorkload("0x_USER_DEMO", l2_business.DualLedger_Verify)
		l3.SetSubStateValue("TCSi_RECOVERY_HUB", "LAST_PROOF_BLOCK", fmt.Sprintf("%d", l1.BlockHeight))
		height, burned := l1.MineBlock(1)
		response = map[string]interface{}{
			"status":       "ok",
			"message":      fmt.Sprintf("Submitted RWA Dual-Ledger Verification Proof on L2 (Success: %t)", success),
			"block_height": height,
			"burned":       burned,
		}
	default:
		response = map[string]interface{}{
			"status":  "ok",
			"message": fmt.Sprintf("Action [%s] processed for role [%s].", action, role),
		}
	}

	json.NewEncoder(w).Encode(response)
}
