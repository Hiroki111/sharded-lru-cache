package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Hiroki111/sharded-lru-cache/pkg/client"
)

func main() {
	// Initialize the client pointing to your local cache server
	c := client.NewClient("http://localhost:8080")

	fmt.Println("🚀 Starting Distributed Simulation...")

	// --- SIMULATED SERVICE A: THE WRITER ---
	fmt.Println("\n[Service A]: Storing product data...")
	err := c.Set("catalog:123", "High-End Laptop", 60)
	if err != nil {
		log.Fatalf("Service A failed: %v", err)
	}
	fmt.Println("[Service A]: Success.")

	// --- SIMULATED NETWORK DELAY ---
	time.Sleep(2 * time.Second)

	// --- SIMULATED SERVICE B: THE READER ---
	fmt.Println("\n[Service B]: Attempting to fetch product 123...")
	val, err := c.Get("catalog:123")
	if err != nil {
		log.Fatalf("Service B failed: %v", err)
	}

	if val == "High-End Laptop" {
		fmt.Printf("[Service B]: Success! Data retrieved: %s\n", val)
		fmt.Println("\n✅ VERIFIED: Data shared across independent service calls.")
	} else {
		fmt.Printf("[Service B]: Error! Data mismatch. Got: %s\n", val)
	}
}
