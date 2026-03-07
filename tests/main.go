package main

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/Hiroki111/sharded-lru-cache/pkg/client"
)

type Product struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	GPA     float64 `json:"gpa"`
	RoleIDs []int   `json:"role_ids"`
	IsAdmin bool    `json:"is_admin"`
}

func main() {
	c := client.NewClient("http://localhost:8080")
	fmt.Println("🚀 Starting Distributed Simulation...")

	// 1. Define our test cases
	jsonInput := map[string]any{
		"id":       json.Number("1"),
		"name":     "test",
		"gpa":      json.Number("3.8"),
		"role_ids": []any{json.Number("1"), json.Number("2"), json.Number("3")},
		"is_admin": true,
	}
	structInput := Product{ID: 2, Name: "another test", GPA: 3.9, RoleIDs: []int{4, 5}, IsAdmin: true}

	// 2. Run Write/Read cycles
	fmt.Println("\n--- Phase 1: Storage & Retrieval ---")

	// Test Case 1: Generic Map (using Get)
	runTest(c, "catalog:123", jsonInput, func(k string) (any, error) {
		return c.Get(k)
	})

	// Test Case 2: Typed Struct (using GetAs)
	runTest(c, "catalog:45", structInput, func(k string) (any, error) {
		return client.GetAs[Product](c, k)
	})

	// 3. Maintenance Operations
	fmt.Println("\n--- Phase 2: System Maintenance ---")

	executeStep("Service C: Stats", func() error {
		stats, err := c.Stats()
		if err == nil {
			fmt.Printf("   📈 Stats: %+v\n", stats)
		}
		return err
	})

	executeStep("Service D: AOF Compaction", func() error {
		return c.Compact()
	})

	fmt.Println("\n✨ Simulation Complete.")
}

// --- Helpers to keep main() clean ---

func runTest(c *client.Client, key string, input any, fetcher func(string) (any, error)) {
	fmt.Printf("\n[Testing Key: %s]\n", key)

	// Write
	if err := c.Set(key, input, 60); err != nil {
		log.Fatalf("  ❌ Set failed: %v", err)
	}
	fmt.Println("  ✅ Set Success")

	// Read
	val, err := fetcher(key)
	if err != nil {
		log.Fatalf("  ❌ Get failed: %v", err)
	}

	// Verify
	if reflect.DeepEqual(val, input) {
		fmt.Printf("  ✅ Verify Success: %v\n", val)
	} else {
		log.Fatalf("  ❌ Mismatch! \n     Want: %v\n     Got:  %v", input, val)
	}
}

func executeStep(name string, action func() error) {
	fmt.Printf("[%s]... ", name)
	if err := action(); err != nil {
		log.Fatalf("Failed: %v", err)
	}
	fmt.Println("Done.")
}
