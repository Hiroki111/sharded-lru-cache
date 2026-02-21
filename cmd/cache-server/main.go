package main

import (
	"encoding/json"
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

// TODO: Implement this
func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {

}

func main() {
	mgr := shard.NewCacheManager[string, string](32, 1024)
	mgr.StartJanitor(10 * time.Second)

	srv := &Server{cache: mgr}

	http.HandleFunc("/get", srv.handleGet)
	http.HandleFunc("/set", srv.handleSet)

	println("Server starting on :8080...")
	http.ListenAndServe(":8080", nil)
}
