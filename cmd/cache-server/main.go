package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Hiroki111/sharded-lru-cache/pkg/shard"
)

type Server struct {
	cache *shard.CacheManager[string, any]
}

type setPayload struct {
	Key   string `json:"key"`
	Value []byte `json:"value"`
	TTL   int    `json:"ttl"`
}

func (s *Server) handleSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload setPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ttl := time.Duration(payload.TTL) * time.Second
	if ttl == 0 {
		ttl = 10 * time.Minute
	}

	s.cache.Set(payload.Key, payload.Value, ttl)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "stored"})
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

	value, found := s.cache.Get(key)
	if !found {
		http.Error(w, "Value not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"value": value})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := s.cache.GetStats()

	var hitRate float64
	totalRequests := stats.Hits + stats.Misses
	if totalRequests > 0 {
		hitRate = (float64(stats.Hits) / float64(totalRequests)) * 100
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"hits":      stats.Hits,
		"misses":    stats.Misses,
		"evictions": stats.Evictions,
		"hit_rate":  fmt.Sprintf("%.2f%%", hitRate),
	})
}

// You don't want to compact on every Set (that would be $O(N)$ and slow). You usually trigger it based on:
// Time: Once every hour.
// Size: When the AOF file exceeds 1GB.
// Manual: An admin endpoint /compact.
func (s *Server) handleCompact(w http.ResponseWriter, r *http.Request) {
	err := s.cache.Compact()
	if err != nil {
		http.Error(w, "Compaction failed", 500)
		return
	}
	w.Write([]byte("Compaction successful"))
}

func main() {
	// 1. Configuration
	var maxAofSize int64 = 50 * 1024 * 1024 // 50 MB

	// 2. Initialization
	mgr, err := shard.NewCacheManager[string, any](32, 1024, 3, "cache.aof", maxAofSize)
	if err != nil {
		// Use log.Fatalf for critical startup errors
		log.Fatalf("Critical Error: Failed to initialize cache manager: %v", err)
	}

	// 3. Recovery
	if err := mgr.LoadAOF(); err != nil {
		// A warning is appropriate here as the server can still function
		log.Printf("Warning: Recovery from AOF incomplete: %v", err)
	}

	// 4. Background Workers
	mgr.StartJanitor(10 * time.Second)
	mgr.StartAofSyncer()
	mgr.StartAofMonitor(30 * time.Second)

	srv := &Server{cache: mgr}

	// 5. Routing
	mux := http.NewServeMux() // Using a local mux is cleaner than global http.HandleFunc
	mux.HandleFunc("/get", srv.handleGet)
	mux.HandleFunc("/set", srv.handleSet)
	mux.HandleFunc("/stats", srv.handleStats)
	mux.HandleFunc("/compact", srv.handleCompact)

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// 6. Graceful Shutdown Logic
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down gracefully...")
		mgr.Stop()
		log.Println("AOF flushed. Goodbye!")
		os.Exit(0)
	}()

	log.Println("Server starting on :8080...")
	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
