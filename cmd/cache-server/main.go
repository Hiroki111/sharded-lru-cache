package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Hiroki111/sharded-lru-cache/internal/shard"
)

type Server struct {
	cache *shard.CacheManager[string, string]
}

type setPayload struct {
	Key   string `json:"key"`
	Value string `json:"value"`
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
	json.NewEncoder(w).Encode(map[string]string{"value": value})
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

func main() {
	mgr := shard.NewCacheManager[string, string](32, 1024, 3, "cache.aof")

	if err := mgr.LoadAOF(); err != nil {
		fmt.Println("Warning: Couldn't load AOF: %v\v", err)
	}

	mgr.StartJanitor(10 * time.Second)
	mgr.StartAofSyncer()

	srv := &Server{cache: mgr}

	http.HandleFunc("/get", srv.handleGet)
	http.HandleFunc("/set", srv.handleSet)
	http.HandleFunc("/stats", srv.handleStats)

	println("Server starting on :8080...")
	http.ListenAndServe(":8080", nil)
}
