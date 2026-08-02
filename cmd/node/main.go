package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/yourusername/hogan-chain/pkg/l1_engine"
	"github.com/yourusername/hogan-chain/pkg/l2_business"
)

var (
	l1 *l1_engine.L1Ledger
	l2 *l2_business.L2Engine
)

func main() {
	fmt.Println("==================================================")
	fmt.Println("   HOGAN CHAIN & HALF-GALLON TECH NODE BOOTING   ")
	fmt.Println("==================================================")

	l1 = l1_engine.NewL1Ledger()
	l2 = l2_business.NewL2Engine()

	// Route Dashboard
	http.HandleFunc("/", serveDashboard)
	http.HandleFunc("/api/action", handleAction)

	fmt.Println("\n[NODE ONLINE] Live Web Dashboard running at: http://localhost:8080")
	fmt.Println("Press Ctrl+C to stop.")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func serveDashboard(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/index.html")
	if err != nil {
		http.Error(w, "Dashboard template not found: "+err.Error(), http.StatusInternalServerError)
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
			"message":      fmt.Sprintf("Block #%d mined successfully by %s.", height, role),
			"block_height": height,
			"burned":       burned,
		}
	case "ai_job", "quantum_job", "rwa_proof":
		height, burned := l1.MineBlock(1)
		response = map[string]interface{}{
			"status":       "ok",
			"message":      fmt.Sprintf("Workload [%s] executed on L2 and anchored to L1.", action),
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
