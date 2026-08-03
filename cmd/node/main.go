package main

import (
	"fmt"
	"github.com/yourusername/hogan-chain/internal/app"
	webui "github.com/yourusername/hogan-chain/internal/web"
	"log"
	"net/http"
	"os"
)

func main() {
	genesis := env("GENESIS_PATH", "config/genesis.json")
	db := env("DB_PATH", "data/hogan-chain.db")
	port := env("PORT", "8080")
	a, err := app.Build(genesis, db)
	if err != nil {
		log.Fatal(err)
	}
	defer a.Close()
	server, err := webui.New(a)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("==================================================")
	fmt.Println(" HOGAN CHAIN & HALF-GALLON TECH SYSTEM MANAGER")
	fmt.Println("==================================================")
	fmt.Printf("Dashboard: http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, server.Routes()))
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
