# sharded-lru-cache

🌐 Context & Usage
- The Users: Other backend services (e.g., a microservice needing to cache database results or API responses).
- The Use Case: High-throughput systems where a single global lock would become a bottleneck (e.g., a session store or a metadata cache).
- Scalability: * Throughput: Designed for 100k+ requests per second.
  - Data Volume: Typically gigabytes of RAM.
  - Payloads: Small to medium objects (JSON blobs, protobufs).

🛠️ Minimum Viable Product (MVP)
1. LRU Logic: A doubly linked list 🔗 combined with a hash map 🗺️ for $O(1)$ access and eviction.
2. Sharding Strategy: A hashing function (like fnv64a) to map keys to specific shards.
3. Concurrency Control: Using sync.RWMutex per shard to allow concurrent reads.
4. Basic API: Get(key), Set(key, value), and Delete(key).

🚀 Optional Enhancements
- TTL (Time-to-Live): Automatically expiring keys after a duration.
- Prometheus Metrics: Tracking hit/miss ratios and eviction counts 📊.
- Custom Serialization: Supporting gob or protobuf for cross-network compatibility.
- Distributed Layer: Adding a gRPC or HTTP interface to turn it into a standalone service.

```
/sharded-cache
├── internal/
│   ├── lru/          # The core, non-thread-safe LRU logic
│   │   ├── lru.go
│   │   └── lru_test.go
│   └── shard/        # The sharding layer and locking logic
│       ├── manager.go
│       └── hasher.go
├── pkg/              # Public API for users
│   └── cache.go
├── main.go           # Example usage/CLI
└── go.mod
```